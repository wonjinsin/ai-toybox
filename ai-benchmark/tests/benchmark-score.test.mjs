import assert from 'node:assert/strict';
import { readFile, mkdtemp, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
import test from 'node:test';

import { normalize } from '../skills/benchmark-review/scripts/score-review.mjs';
const root = fileURLToPath(new URL('../', import.meta.url));
const scorer = path.join(root, 'skills/benchmark-review/scripts/score-review.mjs');
const templatePath = path.join(root, 'skills/benchmark-review/assets/review.json');
const registry = JSON.parse(await readFile(
  path.join(root, 'skills/benchmark-review/assets/check-registry.json'), 'utf8'));

async function completeReview(format = '4.0') {
  const review = JSON.parse(await readFile(format === '4.0' ? templatePath
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
  assert.equal(result.value.total_score_upper_bound, 100);
  assert.equal(result.value.raw_total_score_upper_bound, 100);
  assert.ok(result.value.categories.every(category => category.score === 5 && category.score_upper_bound === 5));
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
  assert.equal(result.value.raw_total_score_upper_bound, 96.88);
  assert.equal(result.value.total_score_upper_bound, 69);
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

test('confirmed failure can determine the total without completing unknown evidence', async t => {
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
  assert.deepEqual(result.value.categories.find(category => category.id === 'delivery'),
    { id: 'delivery', weight: 10, score: 4, score_upper_bound: 5 });
  assert.equal(result.value.raw_total_score, 94.88);
  assert.equal(result.value.raw_total_score_upper_bound, 96.88);
  assert.equal(result.value.total_score, 69);
  assert.equal(result.value.total_score_upper_bound, 69);
  assert.deepEqual(result.value.checks, review.checks);
  assert.equal(result.value.score_cap, 69);
});

test('blocked review spans zero to one hundred without claiming observed quality', async t => {
  const blocked = JSON.parse(await readFile(templatePath, 'utf8'));
  blocked.status = 'blocked';
  blocked.notes = ['Synthetic unavailable-browser case.'];
  const result = await score(t, blocked);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.value.status, 'blocked');
  assert.equal(result.value.total_score, 0);
  assert.equal(result.value.total_score_upper_bound, 100);
  assert.equal(result.value.raw_total_score, 0);
  assert.equal(result.value.raw_total_score_upper_bound, 100);
  assert.equal(result.value.score_cap, null);
  assert.ok(result.value.categories.every(category => category.score === 0 && category.score_upper_bound === 5));
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
    review => { review.checks[0].evidence = 'An unsupported generic observation.'; },
    review => { review.checks[0].evidence = 'Action: Check. Expected: Works. Observed: Not observed.'; },
    review => { review.checks[0].evidence = 'Action: Check. Expected: Works. Observed: '; },
    review => { review.checks[0].method = 'static'; },
    review => { review.verification.viewports = ['1440x900']; },
    review => { review.verification.browser_method = 'none'; },
    review => { review.status = 'partial'; },
    review => { review.checks[0].status = 'unverified'; },
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
  assert.equal(Object.hasOwn(result.value, 'total_score_upper_bound'), false);
  assert.equal(Object.hasOwn(result.value, 'raw_total_score_upper_bound'), false);
  assert.ok(result.value.categories.every(category => !Object.hasOwn(category, 'score_upper_bound')));
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

for (const format of ['2.0', '4.0']) {
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
    if (format === '4.0') assert.equal(result.value.score_cap, null);
  });

  for (const status of ['partial', 'blocked']) {
    test(`format ${format} ${status} preserves honest unknown metadata and versioned scoring`, async t => {
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
      assert.equal(result.value.total_score, format === '4.0' ? 0 : null);
      assert.ok(result.value.categories.every(category => category.score === (format === '4.0' ? 0 : null)));
      assert.equal(result.value.requirements_verdict, 'unverified');
      assert.deepEqual(result.value.coverage, { passed: 0, failed: 0, unverified: review.checks.length });
      if (format === '4.0') {
        assert.equal(result.value.raw_total_score, 0);
        assert.equal(result.value.raw_total_score_upper_bound, 100);
        assert.equal(result.value.total_score_upper_bound, 100);
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

for (const format of ['2.0', '4.0']) {
  test(`format ${format} keeps observed findings when reviewer metadata is unavailable`, async t => {
    const review = await completeReview(format);
    review.status = 'partial';
    review.reviewer.effort = null;
    review.notes = ['Synthetic case. The current host does not expose reviewer effort.'];
    const failedId = format === '4.0' ? 'S7' : 'T3';
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
    assert.equal(result.value.total_score, format === '4.0' ? 69 : null);
    if (format === '4.0') {
      assert.equal(result.value.total_score_upper_bound, 69);
      assert.equal(result.value.raw_total_score, 96.88);
      assert.equal(result.value.raw_total_score_upper_bound, 96.88);
      assert.equal(result.value.score_cap, 69);
    }
  });
}

for (const [id, rawLower, totalLower, verdict] of [
  ['D3', 98, 98, 'pass'],
  ['D6', 98, 69, 'unverified'],
]) {
  test(`unknown ${id} preserves full denominators and confirmed-only caps`, async t => {
    const review = await completeReview();
    review.status = 'partial';
    Object.assign(review.checks.find(check => check.id === id), {
      status: 'unverified', rating: null, method: 'none', evidence: 'Synthetic missing observation.',
    });
    const result = await score(t, review);
    assert.equal(result.status, 0, result.stderr);
    assert.equal(result.value.raw_total_score, rawLower);
    assert.equal(result.value.raw_total_score_upper_bound, 100);
    assert.equal(result.value.total_score, totalLower);
    assert.equal(result.value.total_score_upper_bound, 100);
    assert.equal(result.value.score_cap, null);
    assert.equal(result.value.requirements_verdict, verdict);
    assert.deepEqual(result.value.categories.find(category => category.id === 'delivery'),
      { id: 'delivery', weight: 10, score: 4, score_upper_bound: 5 });
    assert.deepEqual(result.value.checks, review.checks);
  });
}

test('metadata-only partial review can have an exact uncapped score', async t => {
  const review = await completeReview();
  review.status = 'partial';
  review.reviewer.model = null;
  const result = await score(t, review);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.value.status, 'partial');
  assert.equal(result.value.total_score, 100);
  assert.equal(result.value.total_score_upper_bound, 100);
  assert.equal(result.value.raw_total_score, 100);
  assert.equal(result.value.raw_total_score_upper_bound, 100);
  assert.equal(result.value.reviewer.model, null);
  assert.equal(result.value.coverage.unverified, 0);
});

test('current scoring rejects missing, nullable, nonfinite, and out-of-range endpoints', async () => {
  const endpoints = ['total_score', 'total_score_upper_bound', 'raw_total_score', 'raw_total_score_upper_bound'];
  for (const field of endpoints) {
    for (const invalid of [undefined, null, NaN, Infinity, -Infinity, -0.01, 100.01, '0', false]) {
      const review = await completeReview();
      if (invalid === undefined) delete review[field];
      else review[field] = invalid;
      assert.throws(() => normalize(review, registry), `${field} must reject ${String(invalid)}`);
    }
  }
  for (const field of ['score', 'score_upper_bound']) {
    for (const invalid of [undefined, null, NaN, Infinity, -Infinity, -0.01, 5.01, '0', false]) {
      const review = await completeReview();
      if (invalid === undefined) delete review.categories[0][field];
      else review.categories[0][field] = invalid;
      assert.throws(() => normalize(review, registry), `${field} must reject ${String(invalid)}`);
    }
  }
});

test('normalization recomputes stale derived endpoints and is idempotent without altering evidence', async () => {
  const review = await completeReview();
  review.total_score = 100;
  review.total_score_upper_bound = 0;
  review.raw_total_score = 99;
  review.raw_total_score_upper_bound = 1;
  review.score_cap = 69;
  review.requirements_verdict = 'fail';
  review.coverage = { passed: 0, failed: 27, unverified: 0 };
  for (const category of review.categories) Object.assign(category, { score: 5, score_upper_bound: 0 });
  const before = structuredClone(review);
  const normalized = normalize(review, registry);
  assert.equal(normalized.total_score, 100);
  assert.equal(normalized.total_score_upper_bound, 100);
  assert.equal(normalized.raw_total_score, 100);
  assert.equal(normalized.raw_total_score_upper_bound, 100);
  assert.equal(normalized.score_cap, null);
  assert.equal(normalized.requirements_verdict, 'pass');
  assert.deepEqual(normalized.coverage, { passed: 27, failed: 0, unverified: 0 });
  assert.ok(normalized.categories.every(category => category.score === 5 && category.score_upper_bound === 5));
  assert.deepEqual(normalize(normalized, registry), normalized);
  assert.deepEqual(review, before);
  assert.deepEqual(normalized.checks, before.checks);
});

test('half-cent endpoints round half up and totals use unrounded category means', async () => {
  const review = await completeReview();
  Object.assign(review.checks.find(check => check.id === 'S7'), { status: 'fail', rating: 0 });
  Object.assign(review.checks.find(check => check.id === 'U3'), { status: 'fail', rating: 1 });
  let normalized = normalize(review, registry);
  assert.equal(normalized.categories.find(category => category.id === 'simulation').score, 4.38);
  assert.equal(normalized.raw_total_score, 94.48);
  assert.equal(normalized.raw_total_score_upper_bound, 94.48);
  review.status = 'partial';
  Object.assign(review.checks.find(check => check.id === 'D3'), {
    status: 'unverified', rating: null, method: 'none', evidence: 'Synthetic missing observation.',
  });
  normalized = normalize(review, registry);
  assert.equal(normalized.raw_total_score, 92.48);
  assert.equal(normalized.raw_total_score_upper_bound, 94.48);
});

test('resolving any unknown rating contracts raw, category, and published endpoints', async () => {
  for (const definition of registry.rubrics['solar-system-4.0'].checks) {
    const review = await completeReview();
    review.status = 'partial';
    const check = review.checks.find(check => check.id === definition.id);
    Object.assign(check, { status: 'unverified', rating: null });
    const before = normalize(review, registry);
    for (const rating of definition.qualitative ? [0, 1, 2, 3, 4, 5] : [0, 5]) {
      review.status = 'complete';
      Object.assign(check, {
        rating, status: (definition.qualitative ? rating >= 3 : rating === 5) ? 'pass' : 'fail',
      });
      const after = normalize(review, registry);
      assert.ok(after.raw_total_score >= before.raw_total_score);
      assert.ok(after.raw_total_score_upper_bound <= before.raw_total_score_upper_bound);
      assert.ok(after.total_score >= before.total_score);
      assert.ok(after.total_score_upper_bound <= before.total_score_upper_bound);
      assert.equal(after.total_score, after.total_score_upper_bound);
      for (const [index, category] of after.categories.entries()) {
        assert.ok(category.score >= before.categories[index].score);
        assert.ok(category.score_upper_bound <= before.categories[index].score_upper_bound);
      }
    }
  }
});

test('obsolete current-policy version is rejected instead of silently reinterpreted', async t => {
  const review = await completeReview();
  review.format_version = '3.0';
  review.rubric_version = 'solar-system-3.0';
  const result = await score(t, review);
  assert.notEqual(result.status, 0);
  assert.equal(result.stdout, '');
});

test('blocked current review recomputes stale numeric fields rather than retaining an earlier score', async () => {
  const review = JSON.parse(await readFile(templatePath, 'utf8'));
  review.status = 'blocked';
  review.total_score = 69;
  review.total_score_upper_bound = 69;
  review.raw_total_score = 90;
  review.raw_total_score_upper_bound = 90;
  review.score_cap = 69;
  review.coverage = { passed: 27, failed: 0, unverified: 0 };
  for (const category of review.categories) Object.assign(category, { score: 5, score_upper_bound: 5 });
  const result = normalize(review, registry);
  assert.deepEqual([result.total_score, result.total_score_upper_bound], [0, 100]);
  assert.deepEqual([result.raw_total_score, result.raw_total_score_upper_bound], [0, 100]);
  assert.equal(result.score_cap, null);
  assert.deepEqual(result.coverage, { passed: 0, failed: 0, unverified: 27 });
  assert.ok(result.categories.every(category => category.score === 0 && category.score_upper_bound === 5));
  assert.deepEqual(normalize(result, registry), result);
});

test('a category with no positive observation still includes every unknown rating in its upper bound', async () => {
  const review = JSON.parse(await readFile(templatePath, 'utf8'));
  review.verification.browser_method = 'http';
  Object.assign(review.checks[0], {
    status: 'fail', rating: 0, method: 'http',
    evidence: 'Action: Synthetic scene check. Expected: Fixture renders. Observed: Fixture does not render.',
  });
  const result = normalize(review, registry);
  assert.deepEqual(result.categories[0], { id: 'simulation', weight: 25, score: 0, score_upper_bound: 4.38 });
  assert.deepEqual([result.raw_total_score, result.raw_total_score_upper_bound], [0, 96.88]);
  assert.deepEqual([result.total_score, result.total_score_upper_bound], [0, 69]);
  assert.equal(result.score_cap, 69);
  assert.equal(result.requirements_verdict, 'fail');
});

test('fully observed zero quality is an exact zero rather than an unknown range', async () => {
  const review = await completeReview();
  for (const check of review.checks) Object.assign(check, { status: 'fail', rating: 0 });
  const result = normalize(review, registry);
  assert.equal(result.status, 'complete');
  assert.deepEqual([result.raw_total_score, result.raw_total_score_upper_bound], [0, 0]);
  assert.deepEqual([result.total_score, result.total_score_upper_bound], [0, 0]);
  assert.ok(result.categories.every(category => category.score === 0 && category.score_upper_bound === 0));
  assert.equal(result.coverage.unverified, 0);
});

test('frozen historical partial evidence keeps nullable totals and observed category scores', async t => {
  const review = await completeReview('2.0');
  review.status = 'partial';
  Object.assign(review.checks.find(check => check.id === 'D3'), {
    status: 'unverified', rating: null, method: 'none', evidence: 'Synthetic historical missing observation.',
  });
  const result = await score(t, review);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.value.total_score, null);
  assert.equal(result.value.categories.find(category => category.id === 'delivery').score, null);
  assert.ok(result.value.categories.filter(category => category.id !== 'delivery').every(category => category.score === 5));
  review.total_score = 95;
  assert.notEqual((await score(t, review)).status, 0);
});
