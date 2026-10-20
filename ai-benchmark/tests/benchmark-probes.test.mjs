import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { compareProbeRuns } from '../skills/benchmark-review/scripts/compare-probes.mjs';

const experiment = new URL('../experiments/observatory-improvement/terra-xhigh/', import.meta.url);
const keys = ['mobile-quota', 'mobile-persistence', 'mobile-invalid-storage', 'mobile-diagnostics',
  'mobile-overview-5.2-false', 'desktop-overview-73-false', 'desktop-overview-120-false', 'mobile-overview-17-true'];
const observations = JSON.parse(await readFile(new URL('manifest.json', experiment), 'utf8')).visual_observations;
const recorded = {};
for (const side of ['baseline', 'candidate']) {
  recorded[side] = Object.fromEntries(await Promise.all(keys.map(async key => [key,
    JSON.parse(await readFile(new URL(`evidence/${side}-${key}.json`, experiment), 'utf8'))])));
}
function fixture() {
  return [structuredClone(recorded.baseline), structuredClone(recorded.candidate), structuredClone(observations)];
}

test('paired evidence separates repaired scenarios from unavailable diagnostics and a full score', () => {
  const result = compareProbeRuns(...fixture());
  assert.equal(result.verified, true);
  assert.deepEqual(result.pairs.filter(pair => pair.baseline === 'fail' && pair.candidate === 'pass').map(pair => pair.scenario_id), [
    'quota-capture-warning', 'quota-delete-feedback', 'quota-import-feedback',
    'overview-framing-5.2-paused', 'overview-framing-73-paused', 'overview-framing-120-paused', 'overview-framing-17-playing',
  ]);
  assert.deepEqual(result.pairs.filter(pair => pair.candidate === 'unverified').map(pair => pair.scenario_id), ['diagnostic-positive-control']);
  assert.equal(result.full_benchmark_score, null);
});

test('a remaining failure or lost observation prevents an improvement verdict', () => {
  for (const status of ['fail', 'unverified']) {
    const [before, after, visuals] = fixture();
    after['mobile-persistence'].observations[0].status = status;
    const result = compareProbeRuns(before, after, visuals);
    assert.equal(result.verified, false);
    assert.equal(status === 'fail' ? result.remaining_failures : result.coverage_lost, 1);
  }
});

test('missing or duplicate evidence cannot silently shrink the comparison', () => {
  for (const mutate of [
    (before, after) => { delete before['mobile-quota']; delete after['mobile-quota']; },
    (_, after) => { after['mobile-quota'].observations.pop(); },
    (_, after) => { after['mobile-quota'].observations[1] = after['mobile-quota'].observations[0]; },
    (_, after) => { after['mobile-persistence'].observations[0].observed = ''; },
    (_, after, visuals) => { visuals.pop(); },
    (_, after, visuals) => { visuals.push(visuals[0]); },
  ]) {
    const values = fixture();
    mutate(...values);
    assert.throws(() => compareProbeRuns(...values));
  }
});

test('unverified framing cannot become a pass without an independent rendered observation', () => {
  const [before, after, visuals] = fixture();
  visuals.find(item => item.side === 'candidate').status = 'unverified';
  assert.equal(compareProbeRuns(before, after, visuals).verified, false);
  after['mobile-overview-5.2-false'].observations[1].status = 'pass';
  assert.throws(() => compareProbeRuns(before, after, visuals), /must not certify rendered framing/);
});

test('wrong viewports, check identities and incomplete storage cleanup invalidate a report', () => {
  for (const mutate of [
    report => { report.viewport.width = 1440; },
    report => { report.viewport.secure = false; },
    report => { report.observations[0].check_id = 'T4'; },
    report => { report.cleanup.storage_restored = false; },
    report => { delete report.cleanup.filler_removed; },
  ]) {
    const [before, after, visuals] = fixture();
    mutate(after['mobile-quota']);
    assert.throws(() => compareProbeRuns(before, after, visuals));
  }
});

test('an unchanged artifact outcome is not reported as a behavioral improvement', () => {
  const [before, after, visuals] = fixture();
  for (const key of keys) before[key] = structuredClone(after[key]);
  for (const item of visuals.filter(item => item.side === 'baseline')) item.status = 'pass';
  const result = compareProbeRuns(before, after, visuals);
  assert.equal(result.repaired_scenarios, 0);
  assert.equal(result.verified, false);
});
