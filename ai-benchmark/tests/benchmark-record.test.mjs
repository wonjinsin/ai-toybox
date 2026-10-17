import assert from 'node:assert/strict';
import { readFile, mkdtemp, writeFile, rm, copyFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
import test from 'node:test';

const root = fileURLToPath(new URL('../', import.meta.url));
const validator = path.join(root, 'skills/benchmark-record/scripts/validate-run.mjs');
const requiredValues = ['session_id', 'model', 'effort', 'duration_ms', 'input_tokens', 'output_tokens'];

function completeRun() {
  return {
    format_version: '2.1',
    status: 'completed',
    session_id: 'host-thread-42',
    model: 'sonnet',
    effort: 'high',
    duration_ms: 1234.5,
    input_tokens: 321,
    output_tokens: 123,
    notes: ['Synthetic CLI fixture with known metadata and measurements.'],
  };
}

async function validate(t, run, script = validator) {
  const directory = await mkdtemp(path.join(tmpdir(), 'benchmark-record-'));
  t.after(() => rm(directory, { recursive: true, force: true }));
  const input = path.join(directory, 'run.json');
  const source = typeof run === 'string' ? run : `${JSON.stringify(run, null, 2)}\n`;
  await writeFile(input, source);
  const result = spawnSync(process.execPath, [script, input], { encoding: 'utf8' });
  assert.equal(result.error, undefined);
  assert.equal(await readFile(input, 'utf8'), source, 'validation must not rewrite input');
  if (result.status === 0) {
    assert.equal(result.stdout, source, 'successful validation must preserve the entire record');
    assert.equal(result.stderr, '');
  } else {
    assert.notEqual(result.status, null, 'validation must exit normally with a failure status');
    assert.equal(result.stdout, '', 'invalid input must not be published');
    assert.ok(result.stderr.trim(), 'rejection must explain the failure');
  }
  return result;
}

async function rejects(t, run, field) {
  const result = await validate(t, run);
  assert.notEqual(result.status, 0, `invalid ${field} must be rejected`);
  assert.match(result.stderr, new RegExp(field));
}

test('completed known values preserve aliases and non-UUID session identity', async t => {
  const run = completeRun();
  assert.equal((await validate(t, run)).status, 0);
  run.notes = [];
  assert.equal((await validate(t, run)).status, 0);
  run.effort = 'none';
  assert.equal((await validate(t, run)).status, 0);
});

test('real zero measurements and maximum safe token counts are valid', async t => {
  const run = completeRun();
  Object.assign(run, { duration_ms: 0, input_tokens: 0, output_tokens: 0 });
  assert.equal((await validate(t, run)).status, 0);
  Object.assign(run, { input_tokens: Number.MAX_SAFE_INTEGER, output_tokens: Number.MAX_SAFE_INTEGER });
  assert.equal((await validate(t, run)).status, 0);
});

for (const field of requiredValues) {
  test(`completed ${field} cannot be null or omitted even with an explanatory note`, async t => {
    const run = completeRun();
    run.notes = [`${field} is unavailable in this synthetic fixture.`];
    run[field] = null;
    await rejects(t, run, field);
    delete run[field];
    await rejects(t, run, field);
  });
}

for (const field of ['session_id', 'model', 'effort']) {
  test(`${field} rejects blank, placeholder and non-string metadata`, async t => {
    const invalid = ['', '  ', 'unknown', 'null', 'undefined', 'unavailable', 'n/a', 'default', 'auto', 'unspecified', 'unset', 'not available', 'not applicable', 'tbd', 'todo', 'placeholder', ' UnKnOwN ', '${CLAUDE_SESSION_ID}', 'thread-${ID}', 42, false, [], {}];
    if (field !== 'effort') invalid.push('none');
    for (const value of invalid) {
      await rejects(t, { ...completeRun(), [field]: value }, field);
    }
  });
}

for (const field of ['duration_ms', 'input_tokens', 'output_tokens']) {
  test(`${field} rejects invalid measurement types and bounds`, async t => {
    const invalid = [-1, '0', '', true, false, [], {}];
    if (field !== 'duration_ms') invalid.push(0.5, Number.MAX_SAFE_INTEGER + 1);
    for (const value of invalid) {
      await rejects(t, { ...completeRun(), [field]: value }, field);
    }
    for (const literal of ['1e309', '-1e309']) {
      const source = JSON.stringify({ ...completeRun(), [field]: 'overflow' }).replace('"overflow"', literal);
      await rejects(t, source, field);
    }
  });
}

for (const status of ['running', 'failed', 'interrupted']) {
  test(`${status} retains honest nulls with notes and still checks known values`, async t => {
    const run = completeRun();
    run.status = status;
    for (const field of requiredValues) run[field] = null;
    run.notes = ['Identity, settings and measurements are unavailable for this synthetic state.'];
    assert.equal((await validate(t, run)).status, 0);
    run.notes = [];
    await rejects(t, run, 'notes');
    run.notes = ['Missing metadata and measurements remain unresolved.'];
    for (const field of ['session_id', 'model', 'effort']) {
      await rejects(t, { ...run, [field]: 'unknown' }, field);
    }
    for (const field of ['duration_ms', 'input_tokens', 'output_tokens']) {
      await rejects(t, { ...run, [field]: -1 }, field);
    }
    assert.equal((await validate(t, { ...completeRun(), status })).status, 0);
  });
}

test('notes have the fixed string-array shape', async t => {
  for (const notes of [null, '', {}, [null], [42], [''], ['   '], ['valid', false]]) {
    await rejects(t, { ...completeRun(), notes }, 'notes');
  }
});

test('the exact format2.1 fields and recognized status are required', async t => {
  for (const field of ['format_version', 'status', 'notes']) {
    const run = completeRun();
    delete run[field];
    await rejects(t, run, field);
  }
  for (const format_version of ['2.0', 2.1, null]) {
    await rejects(t, { ...completeRun(), format_version }, 'format_version');
  }
  for (const status of ['complete', 'unknown', null, false]) {
    await rejects(t, { ...completeRun(), status }, 'status');
  }
  await rejects(t, { ...completeRun(), task_version: 'extra' }, 'task_version');
  for (const source of ['null', '[]', '42', '"record"']) {
    await rejects(t, source, 'run');
  }
});

test('malformed JSON emits only an error and leaves the input untouched', async t => {
  for (const source of ['', '{', '{"duration_ms": NaN}', '{"duration_ms": Infinity}']) {
    assert.notEqual((await validate(t, source)).status, 0);
  }
});

test('the CLI rejects missing paths and extra arguments without stdout', async t => {
  const directory = await mkdtemp(path.join(tmpdir(), 'benchmark-record-cli-'));
  t.after(() => rm(directory, { recursive: true, force: true }));
  for (const args of [[], [path.join(directory, 'missing.json')], ['one.json', 'two.json']]) {
    const result = spawnSync(process.execPath, [validator, ...args], { encoding: 'utf8' });
    assert.equal(result.error, undefined);
    assert.notEqual(result.status, 0);
    assert.equal(result.stdout, '');
    assert.ok(result.stderr.trim());
  }
});

test('the validator works when copied outside the repository', async t => {
  const directory = await mkdtemp(path.join(tmpdir(), 'benchmark-record-standalone-'));
  t.after(() => rm(directory, { recursive: true, force: true }));
  const script = path.join(directory, 'validate-run.mjs');
  await copyFile(validator, script);
  assert.equal((await validate(t, completeRun(), script)).status, 0);
});
