import assert from 'node:assert/strict';
import { copyFile, mkdir, mkdtemp, readFile, readdir, rm, symlink, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import test from 'node:test';

const root = fileURLToPath(new URL('../', import.meta.url));
const reviewRoot = path.join(root, 'skills/benchmark-review');
const recordRoot = path.join(root, 'skills/benchmark-record');

function invoke(script, ...args) {
  const result = spawnSync(process.execPath, [script, ...args], { encoding: 'utf8' });
  assert.equal(result.error, undefined);
  return result;
}

function runRecord(status = 'running') {
  return {
    format_version: '2.1', status, session_id: 'fixture-session', model: 'fixture-model',
    effort: 'high', duration_ms: 12, input_tokens: 34, output_tokens: 56,
    notes: ['Synthetic publication fixture, not a measured benchmark.'],
  };
}

async function fixture(t, format = '4.0') {
  const directory = await mkdtemp(path.join(tmpdir(), 'benchmark-publication-'));
  t.after(() => rm(directory, { recursive: true, force: true }));
  const run = path.join(directory, 'run');
  await mkdir(path.join(run, 'output'), { recursive: true });
  const prompt = format === '2.0' ? 'solar-system-prompt-v1.md' : 'solar-system-prompt.md';
  await copyFile(path.join(recordRoot, 'assets', prompt), path.join(run, 'prompt.md'));
  await writeFile(path.join(run, 'output/index.html'), '<!doctype html><title>Publication fixture</title>');
  await writeFile(path.join(run, 'run.json'), JSON.stringify(runRecord()));
  return { directory, run, draft: path.join(directory, 'draft.json') };
}

async function reviewRecord(format = '4.0') {
  const template = format === '5.0' ? 'review.json' : `review-${format}.json`;
  const review = JSON.parse(await readFile(path.join(reviewRoot, 'assets', template), 'utf8'));
  const registry = JSON.parse(await readFile(path.join(reviewRoot, 'assets/check-registry.json'), 'utf8'));
  review.status = 'complete';
  review.reviewer = { model: 'fixture-reviewer', effort: 'high' };
  review.reviewed_at = '2026-10-18T00:00:00Z';
  review.notes = ['Synthetic observations. No browser execution is claimed by this test.'];
  review.verification = { browser_method: 'http', viewports: ['1440x900', '390x844'] };
  review.checks = registry.rubrics[review.rubric_version].checks.map(({ id, method }) => ({
    id, status: 'pass', rating: 5, method,
    evidence: 'Action: Synthetic publication case. Expected: Persist supplied observations. Observed: Fixture passes.',
  }));
  return review;
}

async function sourceBytes(run) {
  return Promise.all(['run.json', 'prompt.md', 'output/index.html'].map(file => readFile(path.join(run, file), 'utf8')));
}

for (const format of ['2.0', '4.0', '5.0']) {
  test(`review publication produces consistent format ${format} reports without changing the artifact`, async t => {
    const { run, draft } = await fixture(t, format);
    const before = await sourceBytes(run);
    const review = await reviewRecord(format);
    await writeFile(draft, JSON.stringify(review));
    const result = invoke(path.join(reviewRoot, 'scripts/finalize-review.mjs'), run, draft);
    assert.equal(result.status, 0, result.stderr);
    const saved = JSON.parse(await readFile(path.join(run, 'reviews/review.json'), 'utf8'));
    assert.equal(saved.status, 'complete');
    assert.equal(saved.total_score, 100);
    if (format !== '2.0') {
      assert.equal(saved.total_score_upper_bound, 100);
      if (format === '4.0') assert.equal(saved.raw_total_score_upper_bound, 100);
      else {
        assert.equal(Object.hasOwn(saved, 'raw_total_score'), false);
        assert.equal(Object.hasOwn(saved, 'raw_total_score_upper_bound'), false);
        assert.equal(Object.hasOwn(saved, 'score_cap'), false);
      }
    } else {
      assert.equal(Object.hasOwn(saved, 'total_score_upper_bound'), false);
    }
    assert.deepEqual(saved.checks, review.checks);
    const markdown = await readFile(path.join(run, 'reviews/review.md'), 'utf8');
    assert.match(markdown, /100/);
    assert.ok(markdown.replace(/&#(\d+);/g, (_, code) => String.fromCharCode(Number(code))).includes(review.reviewer.model));
    for (const check of review.checks) assert.ok(markdown.includes(check.id));
    const normalized = invoke(path.join(reviewRoot, 'scripts/score-review.mjs'), path.join(run, 'reviews/review.json'));
    assert.equal(normalized.status, 0, normalized.stderr);
    assert.deepEqual(saved, JSON.parse(normalized.stdout));
    assert.deepEqual(await sourceBytes(run), before);
    const pair = await Promise.all(['review.json', 'review.md'].map(file => readFile(path.join(run, 'reviews', file), 'utf8')));
    assert.equal(invoke(path.join(reviewRoot, 'scripts/finalize-review.mjs'), run, draft).status, 0);
    assert.deepEqual(await Promise.all(['review.json', 'review.md'].map(file => readFile(path.join(run, 'reviews', file), 'utf8'))), pair);
    assert.deepEqual((await readdir(path.join(run, 'reviews'))).sort(), ['review.json', 'review.md']);
  });
}

test('review publication preserves incomplete evidence even when a confirmed cap determines the score', async t => {
  const { run, draft } = await fixture(t);
  const review = await reviewRecord();
  review.status = 'partial';
  review.reviewer.effort = null;
  review.notes.push('Reviewer effort and mobile browser input are unavailable.');
  Object.assign(review.checks.find(check => check.id === 'S7'), { status: 'fail', rating: 0 });
  Object.assign(review.checks.find(check => check.id === 'L2'), {
    status: 'unverified', rating: null, method: 'none', evidence: 'Mobile viewport capability is unavailable.',
  });
  await writeFile(draft, JSON.stringify(review));
  const result = invoke(path.join(reviewRoot, 'scripts/finalize-review.mjs'), run, draft);
  assert.equal(result.status, 0, result.stderr);
  const saved = JSON.parse(await readFile(path.join(run, 'reviews/review.json'), 'utf8'));
  assert.equal(saved.status, 'partial');
  assert.equal(saved.requirements_verdict, 'fail');
  assert.equal(saved.total_score, 69);
  assert.equal(saved.total_score_upper_bound, 69);
  assert.equal(saved.raw_total_score, 89.38);
  assert.equal(saved.raw_total_score_upper_bound, 96.88);
  assert.equal(saved.score_cap, 69);
  assert.deepEqual(saved.coverage, { passed: 25, failed: 1, unverified: 1 });
  assert.deepEqual(saved.checks, review.checks);
  assert.equal(saved.reviewer.effort, null);
  const markdown = await readFile(path.join(run, 'reviews/review.md'), 'utf8');
  assert.match(markdown, /S7/);
  assert.match(markdown, /Mobile viewport capability is unavailable/);
});

test('invalid review proposal preserves the previous pair and unrelated files', async t => {
  const { run, draft } = await fixture(t);
  const reviews = path.join(run, 'reviews');
  await mkdir(reviews);
  await writeFile(path.join(reviews, 'review.json'), 'old json');
  await writeFile(path.join(reviews, 'review.md'), 'old markdown');
  await writeFile(path.join(reviews, 'user.txt'), 'user data');
  const invalid = await reviewRecord();
  invalid.checks[0].evidence = '';
  await writeFile(draft, JSON.stringify(invalid));
  const result = invoke(path.join(reviewRoot, 'scripts/finalize-review.mjs'), run, draft);
  assert.notEqual(result.status, 0);
  assert.equal(await readFile(path.join(reviews, 'review.json'), 'utf8'), 'old json');
  assert.equal(await readFile(path.join(reviews, 'review.md'), 'utf8'), 'old markdown');
  assert.equal(await readFile(path.join(reviews, 'user.txt'), 'utf8'), 'user data');
});

for (const target of ['directory', 'json', 'markdown']) {
  test(`review publication refuses a symlinked ${target} without touching its destination`, async t => {
    const { directory, run, draft } = await fixture(t);
    const outside = path.join(directory, 'outside');
    await mkdir(outside);
    const sentinel = path.join(outside, 'sentinel.txt');
    await writeFile(sentinel, 'protected');
    const reviews = path.join(run, 'reviews');
    if (target === 'directory') await symlink(outside, reviews);
    else {
      await mkdir(reviews);
      await symlink(sentinel, path.join(reviews, target === 'json' ? 'review.json' : 'review.md'));
    }
    await writeFile(draft, JSON.stringify(await reviewRecord()));
    const result = invoke(path.join(reviewRoot, 'scripts/finalize-review.mjs'), run, draft);
    assert.notEqual(result.status, 0);
    assert.equal(await readFile(sentinel, 'utf8'), 'protected');
    assert.deepEqual(await readdir(outside), ['sentinel.txt']);
  });
}

async function recordDigest(run) {
  return createHash('sha256').update(await readFile(path.join(run, 'run.json'))).digest('hex');
}

test('an early checkpoint can finish with honest null measurements after unavailable usage', async t => {
  const directory = await mkdtemp(path.join(tmpdir(), 'benchmark-checkpoint-'));
  t.after(() => rm(directory, { recursive: true, force: true }));
  const run = path.join(directory, 'run');
  const draft = path.join(directory, 'draft.json');
  const script = path.join(recordRoot, 'scripts/validate-run.mjs');
  await mkdir(run);
  const record = runRecord();
  record.input_tokens = null;
  record.output_tokens = null;
  record.notes = ['Input and output usage are unavailable from the current host.'];
  await writeFile(draft, JSON.stringify(record));
  const checkpoint = invoke(script, '--checkpoint', draft, run);
  assert.equal(checkpoint.status, 0, checkpoint.stderr);
  assert.deepEqual(JSON.parse(await readFile(path.join(run, 'run.json'), 'utf8')), record);
  const digest = await recordDigest(run);
  record.status = 'failed';
  await writeFile(draft, JSON.stringify(record));
  const final = invoke(script, '--finalize', draft, run, digest);
  assert.equal(final.status, 0, final.stderr);
  const saved = JSON.parse(await readFile(path.join(run, 'run.json'), 'utf8'));
  assert.equal(saved.status, 'failed');
  assert.equal(saved.input_tokens, null);
  assert.equal(saved.output_tokens, null);
  assert.equal(invoke(script, path.join(run, 'run.json')).status, 0);
});

test('a measured artifact with an unavailable browser check finalizes without pretending that check passed', async t => {
  const { run, draft } = await fixture(t);
  const record = runRecord('completed');
  record.notes.push('Mobile zoom could not be verified because the browser input API was unavailable.');
  await writeFile(draft, JSON.stringify(record));
  const result = invoke(path.join(recordRoot, 'scripts/validate-run.mjs'), '--finalize', draft, run, await recordDigest(run));
  assert.equal(result.status, 0, result.stderr);
  assert.deepEqual(JSON.parse(await readFile(path.join(run, 'run.json'), 'utf8')), record);
});

for (const scenario of ['invalid-value', 'stale-write', 'changed-session', 'terminal-record', 'missing-artifact']) {
  test(`record publication rejects ${scenario} without corrupting the saved record`, async t => {
    const { run, draft } = await fixture(t);
    const script = path.join(recordRoot, 'scripts/validate-run.mjs');
    const proposal = runRecord('completed');
    let digest = await recordDigest(run);
    if (scenario === 'invalid-value') proposal.input_tokens = '34';
    if (scenario === 'stale-write') digest = '0'.repeat(64);
    if (scenario === 'changed-session') proposal.session_id = 'another-session';
    if (scenario === 'terminal-record') {
      await writeFile(path.join(run, 'run.json'), JSON.stringify(runRecord('completed')));
      digest = await recordDigest(run);
    }
    if (scenario === 'missing-artifact') await rm(path.join(run, 'output/index.html'));
    const before = await readFile(path.join(run, 'run.json'), 'utf8');
    await writeFile(draft, JSON.stringify(proposal));
    const result = invoke(script, '--finalize', draft, run, digest);
    assert.notEqual(result.status, 0);
    assert.equal(await readFile(path.join(run, 'run.json'), 'utf8'), before);
  });
}

test('record publication rejects a symlinked record without replacing the linked data', async t => {
  const { directory, run, draft } = await fixture(t);
  const outside = path.join(directory, 'outside.json');
  const source = JSON.stringify(runRecord());
  await writeFile(outside, source);
  await rm(path.join(run, 'run.json'));
  await symlink(outside, path.join(run, 'run.json'));
  await writeFile(draft, JSON.stringify(runRecord('completed')));
  const result = invoke(path.join(recordRoot, 'scripts/validate-run.mjs'), '--finalize', draft, run, createHash('sha256').update(source).digest('hex'));
  assert.notEqual(result.status, 0);
  assert.equal(await readFile(outside, 'utf8'), source);
});

test('a safe target with missing artifacts still receives a complete blocked review record', async t => {
  const { run, draft } = await fixture(t);
  await rm(path.join(run, 'output'), { recursive: true });
  await writeFile(path.join(run, 'run.json'), '{');
  const review = JSON.parse(await readFile(path.join(reviewRoot, 'assets/review.json'), 'utf8'));
  review.status = 'blocked';
  review.notes = ['The saved run JSON is malformed and output is missing. Reviewer metadata and timestamp are unavailable.'];
  await writeFile(draft, JSON.stringify(review));
  const result = invoke(path.join(reviewRoot, 'scripts/finalize-review.mjs'), run, draft);
  assert.equal(result.status, 0, result.stderr);
  const saved = JSON.parse(await readFile(path.join(run, 'reviews/review.json'), 'utf8'));
  assert.equal(saved.status, 'blocked');
  assert.equal(saved.total_score, 0);
  assert.equal(saved.total_score_upper_bound, 100);
  assert.equal(saved.format_version, '5.0');
  assert.equal(Object.hasOwn(saved, 'raw_total_score'), false);
  assert.equal(Object.hasOwn(saved, 'raw_total_score_upper_bound'), false);
  assert.equal(Object.hasOwn(saved, 'score_cap'), false);
  assert.deepEqual(saved.coverage, { passed: 0, failed: 0, unverified: 27 });
  assert.equal(await readFile(path.join(run, 'run.json'), 'utf8'), '{');
  assert.match(await readFile(path.join(run, 'reviews/review.md'), 'utf8'), /output is missing/);
});

test('a completed review cannot be published without its artifact', async t => {
  const { run, draft } = await fixture(t);
  await rm(path.join(run, 'output/index.html'));
  await writeFile(draft, JSON.stringify(await reviewRecord()));
  const result = invoke(path.join(reviewRoot, 'scripts/finalize-review.mjs'), run, draft);
  assert.notEqual(result.status, 0);
  assert.equal((await readdir(run)).includes('reviews'), false);
});

test('failure recording does not depend on unsafe or broken artifact links', async t => {
  const { directory, run, draft } = await fixture(t);
  await symlink(path.join(directory, 'missing-resource'), path.join(run, 'output/broken-link'));
  const proposal = runRecord('failed');
  proposal.notes.push('Output contains a broken resource link; implementation could not finish.');
  await writeFile(draft, JSON.stringify(proposal));
  const result = invoke(path.join(recordRoot, 'scripts/validate-run.mjs'), '--finalize', draft, run, await recordDigest(run));
  assert.equal(result.status, 0, result.stderr);
  assert.equal(JSON.parse(await readFile(path.join(run, 'run.json'), 'utf8')).status, 'failed');
});

for (const scenario of [
  { name: 'optional unknown', unknown: 'D3', total: [98, 100], raw: [98, 100], cap: null, verdict: 'pass' },
  { name: 'mandatory unknown', unknown: 'D6', total: [69, 100], raw: [98, 100], cap: null, verdict: 'unverified' },
  { name: 'confirmed failure with unknown', unknown: 'D3', failed: 'S7', total: [69, 69], raw: [94.88, 96.88], cap: 69, verdict: 'fail' },
  { name: 'metadata only', total: [100, 100], raw: [100, 100], cap: null, verdict: 'pass' },
  { name: 'blocked', blocked: true, total: [0, 100], raw: [0, 100], cap: null, verdict: 'unverified' },
]) {
  test(`published ${scenario.name} bounds and evidence are readable without treating the review as complete`, async t => {
    const { run, draft } = await fixture(t);
    const before = await sourceBytes(run);
    const review = await reviewRecord();
    review.status = scenario.blocked ? 'blocked' : 'partial';
    review.reviewer.effort = null;
    review.notes.push('Synthetic unavailable metadata and observations.');
    if (scenario.failed) Object.assign(review.checks.find(check => check.id === scenario.failed), { status: 'fail', rating: 0 });
    for (const check of review.checks) {
      if (scenario.blocked || check.id === scenario.unknown) {
        Object.assign(check, { status: 'unverified', rating: null, method: 'none', evidence: `Synthetic unavailable observation for ${check.id}.` });
      }
    }
    if (scenario.blocked) {
      review.verification = { browser_method: 'none', viewports: [] };
    }
    await writeFile(draft, JSON.stringify(review));
    const published = invoke(path.join(reviewRoot, 'scripts/finalize-review.mjs'), run, draft);
    assert.equal(published.status, 0, published.stderr);
    const receipt = JSON.parse(published.stdout);
    assert.equal(receipt.ok, true);
    assert.deepEqual(receipt.outputs, { 'review.json': 'saved', 'review.md': 'saved' });
    const saved = JSON.parse(await readFile(path.join(run, 'reviews/review.json'), 'utf8'));
    assert.deepEqual([saved.total_score, saved.total_score_upper_bound], scenario.total);
    assert.deepEqual([saved.raw_total_score, saved.raw_total_score_upper_bound], scenario.raw);
    assert.equal(saved.score_cap, scenario.cap);
    assert.equal(saved.requirements_verdict, scenario.verdict);
    assert.equal(saved.status, review.status);
    assert.deepEqual(saved.checks, review.checks);
    assert.equal(saved.reviewer.effort, null);
    const markdown = (await readFile(path.join(run, 'reviews/review.md'), 'utf8'))
      .replace(/&#(\d+);/g, (_, code) => String.fromCharCode(Number(code)));
    const numbers = value => (value.match(/\d+(?:\.\d+)?/g) ?? []).map(Number);
    const headline = markdown.split('\n').find(line => line.includes(' / 100.'));
    assert.deepEqual(numbers(headline), [
      scenario.total[0], 100,
      ...(scenario.total[0] === scenario.total[1] ? [scenario.total[0]] : scenario.total),
    ]);
    for (const category of saved.categories) {
      const row = markdown.split('\n').find(line => line.startsWith('| ') && line.includes(category.id));
      assert.ok(row, `report must expose ${category.id}`);
      assert.deepEqual(numbers(row.split('|')[2]), category.score === category.score_upper_bound
        ? [category.score] : [category.score, category.score_upper_bound]);
    }
    const scored = invoke(path.join(reviewRoot, 'scripts/score-review.mjs'), path.join(run, 'reviews/review.json'));
    assert.equal(scored.status, 0, scored.stderr);
    assert.deepEqual(JSON.parse(scored.stdout), saved);
    assert.deepEqual(await sourceBytes(run), before);
  });
}

for (const scenario of [
  { name: 'optional unknown', unknown: 'D3', total: [98, 100], verdict: 'pass', failed: [], mandatoryUnknown: [] },
  { name: 'mandatory unknown', unknown: 'D6', total: [98, 100], verdict: 'unverified', failed: [], mandatoryUnknown: ['D6'] },
  { name: 'mandatory failure and unknown', unknown: 'D6', total: [94.88, 96.88], verdict: 'fail', failed: ['S7'], mandatoryUnknown: ['D6'] },
  { name: 'complete mandatory failure', total: [96.88, 96.88], verdict: 'fail', failed: ['S7'], mandatoryUnknown: [], complete: true },
  { name: 'metadata only', total: [100, 100], verdict: 'pass', failed: [], mandatoryUnknown: [] },
]) {
  test(`current publication exposes uncapped quality and independent verdict for ${scenario.name}`, async t => {
    const { run, draft } = await fixture(t, '5.0');
    const before = await sourceBytes(run);
    const review = await reviewRecord('5.0');
    review.status = scenario.complete ? 'complete' : 'partial';
    if (!scenario.complete) review.reviewer.effort = null;
    for (const check of review.checks) {
      if (scenario.failed.includes(check.id)) Object.assign(check, { status: 'fail', rating: 0 });
      if (check.id === scenario.unknown) Object.assign(check, {
        status: 'unverified', rating: null, method: 'none', evidence: 'Synthetic observation unavailable.',
      });
    }
    const draftBytes = JSON.stringify(review);
    await writeFile(draft, draftBytes);
    const published = invoke(path.join(reviewRoot, 'scripts/finalize-review.mjs'), run, draft);
    assert.equal(published.status, 0, published.stderr);
    assert.equal(JSON.parse(published.stdout).ok, true);
    const saved = JSON.parse(await readFile(path.join(run, 'reviews/review.json'), 'utf8'));
    assert.deepEqual([saved.total_score, saved.total_score_upper_bound], scenario.total);
    assert.equal(saved.requirements_verdict, scenario.verdict);
    assert.equal(saved.status, review.status);
    assert.deepEqual(saved.checks, review.checks);
    assert.deepEqual(saved.reviewer, review.reviewer);
    assert.equal(saved.reviewed_at, review.reviewed_at);
    assert.deepEqual(saved.notes, review.notes);
    for (const field of ['raw_total_score', 'raw_total_score_upper_bound', 'score_cap']) {
      assert.equal(Object.hasOwn(saved, field), false);
    }
    const markdown = (await readFile(path.join(run, 'reviews/review.md'), 'utf8'))
      .replace(/&#(\d+);/g, (_, code) => String.fromCharCode(Number(code)));
    const headline = markdown.split('\n').find(line => line.includes(' / 100.'));
    assert.deepEqual((headline.match(/\d+(?:\.\d+)?/g) ?? []).map(Number), [
      scenario.total[0], 100,
      ...(scenario.total[0] === scenario.total[1] ? [scenario.total[0]] : scenario.total),
    ]);
    assert.equal(await readFile(draft, 'utf8'), draftBytes);
    assert.deepEqual(await sourceBytes(run), before);
  });
}
