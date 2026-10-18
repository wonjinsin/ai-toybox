import assert from 'node:assert/strict';
import { readFile, mkdtemp, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
import test from 'node:test';

const root = fileURLToPath(new URL('../', import.meta.url));
const scorer = path.join(root, 'skills/benchmark-review/scripts/score-review.mjs');
const templatePath = path.join(root, 'skills/benchmark-review/assets/review.json');

async function completeReview(format = '3.0') {
  const review = JSON.parse(await readFile(format === '3.0' ? templatePath
    : path.join(root, 'skills/benchmark-review/assets/review-2.0.json'), 'utf8'));
  if (format === '2.0') {
    const registry = JSON.parse(await readFile(
      path.join(root, 'skills/benchmark-review/assets/check-registry.json'), 'utf8'));
    review.checks = registry.rubrics['solar-system-2.0'].checks.map(({ id }) => ({ id }));
  }
  review.status = 'complete';
  review.reviewer = { model: 'synthetic-reviewer', effort: 'high' };
  review.reviewed_at = '2026-10-17T09:00:00Z';
  review.notes = ['Synthetic fixture. Metadata and observations are test data, not an actual review.'];
  review.verification = { browser_method: 'http', viewports: ['1440x900', '390x844'] };
  review.checks = review.checks.map(check => ({
    ...check, status: 'pass', rating: 5, method: check.id === 'D2' ? 'static' : 'http',
    evidence: `Action: Synthetic scoring case for ${check.id}. Expected: The supplied fixture outcome. Observed: Synthetic outcome only, not a browser observation.`,
  }));
  return review;
}

async function score(t, review) {
  const directory = await mkdtemp(path.join(tmpdir(), 'benchmark-score-'));
  t.after(() => rm(directory, { recursive: true, force: true }));
  const input = path.join(directory, 'review.json');
  const source = JSON.stringify(review, null, 2);
  await writeFile(input, source);
  const result = spawnSync(process.execPath, [scorer, input], { encoding: 'utf8' });
  assert.equal(await readFile(input, 'utf8'), source, 'scoring must not rewrite evidence');
  return { ...result, value: result.status === 0 ? JSON.parse(result.stdout) : null };
}

test('complete satisfied requirements earn the uncapped weighted score', async t => {
  const result = await score(t, await completeReview());
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.value.total_score, 100);
  assert.equal(result.value.raw_total_score, 100);
  assert.equal(result.value.score_cap, null);
  assert.equal(result.value.requirements_verdict, 'pass');
  assert.deepEqual(result.value.coverage, { passed: 27, failed: 0, unverified: 0 });
});

test('a mandatory rendering failure caps the headline while preserving its raw score', async t => {
  const review = await completeReview();
  Object.assign(review.checks.find(check => check.id === 'S7'), { status: 'fail', rating: 0 });
  const result = await score(t, review);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.value.requirements_verdict, 'fail');
  assert.equal(result.value.raw_total_score, 96.88);
  assert.equal(result.value.score_cap, 69);
  assert.equal(result.value.total_score, 69);
});

test('qualitative failure does not invent a mandatory failure or cap', async t => {
  const review = await completeReview();
  Object.assign(review.checks.find(check => check.id === 'V1'), { status: 'fail', rating: 0 });
  const result = await score(t, review);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.value.requirements_verdict, 'pass');
  assert.equal(result.value.score_cap, null);
  assert.equal(result.value.total_score, 92.5);
});

test('unknown evidence keeps both totals null even with a confirmed failure', async t => {
  const review = await completeReview();
  review.status = 'partial';
  Object.assign(review.checks.find(check => check.id === 'S7'), { status: 'fail', rating: 0 });
  Object.assign(review.checks.find(check => check.id === 'D6'), {
    status: 'unverified', rating: null, method: 'none', evidence: 'Persistence was not observed.',
  });
  const result = await score(t, review);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.value.status, 'partial');
  assert.equal(result.value.requirements_verdict, 'fail');
  assert.equal(result.value.categories.find(category => category.id === 'delivery').score, null);
  assert.equal(result.value.raw_total_score, null);
  assert.equal(result.value.total_score, null);
  assert.equal(result.value.score_cap, 69);
});

test('blocked review remains unscored and cannot retain earlier verified checks', async t => {
  const blocked = JSON.parse(await readFile(templatePath, 'utf8'));
  blocked.status = 'blocked';
  blocked.notes = ['Synthetic unavailable-browser case.'];
  const result = await score(t, blocked);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.value.status, 'blocked');
  assert.equal(result.value.total_score, null);
  assert.equal(result.value.requirements_verdict, 'unverified');
  assert.deepEqual(result.value.coverage, { passed: 0, failed: 0, unverified: 27 });
  Object.assign(blocked.checks[0], { status: 'pass', rating: 5, method: 'http' });
  assert.notEqual((await score(t, blocked)).status, 0);
});

test('malformed evidence cannot enter the scoring calculation', async t => {
  const mutations = [
    review => { review.checks[1].id = review.checks[0].id; },
    review => { review.checks.pop(); },
    review => { review.checks[0].id = 'UNKNOWN'; },
    review => { review.checks[0].rating = 4; },
    review => { review.checks[0].rating = null; },
    review => { review.checks[0].evidence = ''; },
    review => { review.rubric_version = 'solar-system-2.0'; },
    review => { review.categories[0].weight = 99; },
  ];
  for (const mutate of mutations) {
    const review = await completeReview();
    mutate(review);
    const result = await score(t, review);
    assert.notEqual(result.status, 0, 'invalid input must be rejected');
    assert.equal(result.stdout.trim(), '', 'invalid input must not emit a scored report');
  }
});

test('historical reviews retain their original score and schema without the new cap', async t => {
  const review = await completeReview('2.0');
  for (const check of review.checks) {
    const rating = ['T3', 'U2'].includes(check.id) ? 0 : ['U3', 'V1', 'V2'].includes(check.id) ? 3 : 5;
    Object.assign(check, {
      status: rating === 0 ? 'fail' : 'pass', rating,
      evidence: `Synthetic historical fixture for ${check.id}; not a browser observation.`,
    });
  }
  const result = await score(t, review);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.value.total_score, 82);
  assert.equal(result.value.requirements_verdict, 'fail');
  assert.equal(Object.hasOwn(result.value, 'raw_total_score'), false);
  assert.equal(Object.hasOwn(result.value, 'score_cap'), false);
});

test('the mandatory-failure cap never raises a low raw score', async t => {
  const review = await completeReview();
  for (const check of review.checks) {
    if (!['U3', 'V1', 'V2'].includes(check.id)) {
      Object.assign(check, { status: 'fail', rating: 0 });
    }
  }
  const result = await score(t, review);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.value.raw_total_score, 18);
  assert.equal(result.value.score_cap, 69);
  assert.equal(result.value.total_score, 18);
});

for (const format of ['2.0', '3.0']) {
  for (const field of ['model', 'effort', 'reviewed_at']) {
    test(`format ${format} complete requires concrete ${field} even with a note`, async t => {
      const invalid = [
        undefined, null, '', '   ', 0, false, {}, [],
        'unknown', 'null', 'undefined', 'unavailable', 'n/a', 'default', 'auto',
        'unspecified', 'unset', 'not available', 'not applicable', 'tbd', 'todo',
        'placeholder', ' UNKNOWN ', ' Default ', ' AUTO ', '${REVIEW_METADATA}',
      ];
      if (field !== 'effort') invalid.push('none');
      if (field === 'reviewed_at') {
        invalid.push('2026-10-17', '2026-10-17T09:00:00', 'not-a-date',
          '2026-02-30T09:00:00Z', '2025-02-29T09:00:00Z', '2026-10-17T24:00:00Z');
      }
      for (const value of invalid) {
        const review = await completeReview(format);
        const target = field === 'reviewed_at' ? review : review.reviewer;
        if (value === undefined) delete target[field];
        else target[field] = value;
        review.notes = ['Synthetic case. Missing metadata is explained but cannot waive completion requirements.'];
        const result = await score(t, review);
        assert.notEqual(result.status, 0, `${field}=${JSON.stringify(value)} must be rejected`);
        assert.equal(result.stdout, '', 'invalid metadata must not emit a scored report');
        assert.ok(result.stderr.includes(field === 'reviewed_at' ? field : `reviewer.${field}`),
          `field-specific diagnostic required: ${result.stderr}`);
      }
    });
  }

  test(`format ${format} accepts concrete metadata without a note and preserves it`, async t => {
    const review = await completeReview(format);
    review.notes = [];
    review.reviewer.effort = 'none';
    review.reviewed_at = '2024-02-29T18:00:00.123+09:00';
    const result = await score(t, review);
    assert.equal(result.status, 0, result.stderr);
    assert.equal(result.value.status, 'complete');
    assert.deepEqual(result.value.reviewer, review.reviewer);
    assert.equal(result.value.reviewed_at, review.reviewed_at);
    assert.equal(result.value.total_score, 100);
    if (format === '3.0') assert.equal(result.value.score_cap, null);
  });

  for (const status of ['partial', 'blocked']) {
    test(`format ${format} ${status} preserves honest unknown metadata and null scores`, async t => {
      const review = await completeReview(format);
      review.status = status;
      review.reviewer = { model: null, effort: null };
      review.reviewed_at = null;
      review.notes = ['Synthetic case. Reviewer metadata, time, and browser observations are unavailable.'];
      review.verification = { browser_method: 'none', viewports: [] };
      review.checks = review.checks.map(check => ({
        ...check, status: 'unverified', rating: null, method: 'none', evidence: 'Not observed in this synthetic case.',
      }));
      const result = await score(t, review);
      assert.equal(result.status, 0, result.stderr);
      assert.equal(result.value.status, status);
      assert.deepEqual(result.value.reviewer, { model: null, effort: null });
      assert.equal(result.value.reviewed_at, null);
      assert.equal(result.value.total_score, null);
      assert.ok(result.value.categories.every(category => category.score === null));
      assert.equal(result.value.requirements_verdict, 'unverified');
      assert.deepEqual(result.value.coverage, { passed: 0, failed: 0, unverified: review.checks.length });
      if (format === '3.0') {
        assert.equal(result.value.raw_total_score, null);
        assert.equal(result.value.score_cap, null);
      }
      review.notes = [];
      const missingReason = await score(t, review);
      assert.notEqual(missingReason.status, 0);
      assert.match(missingReason.stderr, /missing metadata requires a note/);
      assert.equal(missingReason.stdout, '');
    });
  }
}

for (const format of ['2.0', '3.0']) {
  test(`format ${format} keeps observed findings when reviewer metadata is unavailable`, async t => {
    const review = await completeReview(format);
    review.status = 'partial';
    review.reviewer.effort = null;
    review.notes = ['Synthetic case. The current host does not expose reviewer effort.'];
    const failedId = format === '3.0' ? 'S7' : 'T3';
    Object.assign(review.checks.find(check => check.id === failedId), { status: 'fail', rating: 0 });
    const result = await score(t, review);
    assert.equal(result.status, 0, result.stderr);
    assert.equal(result.value.status, 'partial');
    assert.equal(result.value.requirements_verdict, 'fail');
    assert.equal(result.value.reviewer.effort, null);
    assert.deepEqual(result.value.checks, review.checks);
    assert.deepEqual(result.value.coverage, {
      passed: review.checks.length - 1, failed: 1, unverified: 0,
    });
    assert.equal(result.value.total_score, null);
    if (format === '3.0') {
      assert.equal(result.value.raw_total_score, null);
      assert.equal(result.value.score_cap, 69);
    }
  });
}
