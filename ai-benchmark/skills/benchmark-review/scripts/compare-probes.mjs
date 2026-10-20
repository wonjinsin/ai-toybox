#!/usr/bin/env node
import { readFile, readdir, lstat } from 'node:fs/promises';
import { createHash } from 'node:crypto';
import { fileURLToPath, pathToFileURL } from 'node:url';
import path from 'node:path';
import { projectCurrentPolicy } from './score-review.mjs';

const root = fileURLToPath(new URL('../../../', import.meta.url));
const hash = value => createHash('sha256').update(value).digest('hex');
const requireValue = (condition, message) => { if (!condition) throw new Error(message); };
const plans = [
  { key: 'mobile-quota', scenario: 'quota', size: [390, 844], ids: ['quota-capture-warning', 'quota-delete-feedback', 'quota-import-feedback', 'quota-memory-restore'] },
  { key: 'mobile-persistence', scenario: 'persistence', size: [390, 844], ids: ['capture-persistence', 'reload-live-defaults', 'import-persistence', 'delete-persistence'] },
  { key: 'mobile-invalid-storage', scenario: 'invalid-storage', size: [390, 844], ids: ['invalid-storage-malformed', 'invalid-storage-version', 'invalid-storage-unknown-planet', 'invalid-storage-duplicate-id'] },
  { key: 'mobile-diagnostics', scenario: 'diagnostics', size: [390, 844], ids: ['diagnostic-positive-control'] },
  ...[['mobile', [390, 844], 5.2, false], ['desktop', [1440, 900], 73, false], ['desktop', [1440, 900], 120, false], ['mobile', [390, 844], 17, true]].map(([size, dimensions, time, playing]) => ({
    key: `${size}-overview-${time}-${playing}`, scenario: 'overview', size: dimensions,
    ids: [`overview-state-${time}-${playing ? 'playing' : 'paused'}`, `overview-framing-${time}-${playing ? 'playing' : 'paused'}`],
  })),
];

export function compareProbeRuns(baseline, candidate, visualObservations) {
  const pairs = [];
  requireValue(Array.isArray(visualObservations), 'Visual observations are required');
  const visuals = new Map();
  for (const observation of visualObservations) {
    const key = `${observation.side}:${observation.plan}`;
    requireValue(['baseline', 'candidate'].includes(observation.side) && plans.some(plan => plan.key === observation.plan && plan.scenario === 'overview'), 'Unknown visual observation target');
    requireValue(!visuals.has(key), 'Duplicate visual observation');
    requireValue(['pass', 'fail', 'unverified'].includes(observation.status) && typeof observation.observed === 'string' && observation.observed.trim(), 'Visual observation needs a status and concrete evidence');
    visuals.set(key, observation);
  }
  for (const [side, reports] of [['baseline', baseline], ['candidate', candidate]]) {
    requireValue(reports && Object.keys(reports).length === plans.length && plans.every(plan => Object.hasOwn(reports, plan.key)), `${side} must contain every required probe exactly once`);
    for (const plan of plans) {
      const report = reports[plan.key];
      requireValue(report.format_version === '1.0' && report.scope === 'focused-scenarios-only' && report.profile === 'terra-xhigh-observatory-2' && report.scenario === plan.scenario, `Incompatible ${side} ${plan.key} report`);
      requireValue(report.viewport?.secure === true && report.viewport.width === plan.size[0] && report.viewport.height === plan.size[1], `Wrong viewport for ${side} ${plan.key}`);
      requireValue(Array.isArray(report.observations) && report.observations.length === plan.ids.length && plan.ids.every((id, index) => report.observations[index].scenario_id === id), `Missing, reordered or duplicate observations in ${side} ${plan.key}`);
      requireValue(report.cleanup && report.cleanup.storage_restored !== false && report.cleanup.filler_removed !== false, `Cleanup failed for ${side} ${plan.key}`);
      if (['quota', 'persistence', 'invalid-storage'].includes(plan.scenario)) requireValue(report.cleanup.storage_restored === true, `Storage restoration not proved for ${side} ${plan.key}`);
      if (plan.scenario === 'quota') requireValue(report.cleanup.filler_removed === true, `Quota cleanup not proved for ${side}`);
      for (const observation of report.observations) {
        requireValue(['pass', 'fail', 'unverified'].includes(observation.status), 'Invalid probe outcome');
        requireValue(observation.check_id === (plan.scenario === 'overview' ? 'T4' : plan.scenario === 'diagnostics' ? 'D3' : 'D6'), 'Probe check ownership mismatch');
        requireValue(['action', 'expected', 'observed'].every(field => typeof observation[field] === 'string' && observation[field].trim()), 'Probe evidence is incomplete');
      }
    }
  }
  for (const plan of plans) {
    plan.ids.forEach((id, index) => {
      const original = baseline[plan.key].observations[index], repaired = candidate[plan.key].observations[index];
      let before = original.status, after = repaired.status;
      const visual = id.startsWith('overview-framing-');
      if (visual) {
        requireValue(original.status === 'unverified' && repaired.status === 'unverified', 'The automated probe must not certify rendered framing');
        requireValue(visuals.has(`baseline:${plan.key}`) && visuals.has(`candidate:${plan.key}`), `Missing visual evidence for ${plan.key}`);
        before = visuals.get(`baseline:${plan.key}`).status;
        after = visuals.get(`candidate:${plan.key}`).status;
      }
      pairs.push({ plan: plan.key, scenario_id: id, check_id: original.check_id,
        method: visual ? 'visual' : 'automated', baseline: before, candidate: after });
    });
  }
  const improved = pairs.filter(pair => pair.baseline === 'fail' && pair.candidate === 'pass');
  const failures = pairs.filter(pair => pair.candidate === 'fail');
  const coverageLost = pairs.filter(pair => pair.baseline !== 'unverified' && pair.candidate === 'unverified');
  const unknown = pairs.filter(pair => pair.candidate === 'unverified');
  const unprovedRequired = unknown.filter(pair => pair.check_id !== 'D3');
  return {
    scope: 'focused-scenarios-only',
    verified: improved.length > 0 && failures.length === 0 && coverageLost.length === 0 && unprovedRequired.length === 0,
    repaired_scenarios: improved.length, remaining_failures: failures.length,
    coverage_lost: coverageLost.length, unverified_scenarios: unknown.length, pairs,
    full_benchmark_score: null,
    limits: ['This gate proves only the listed regression scenarios, not a complete T4/D6 check or a 27-check review.',
      'D3 diagnostic transport may remain unavailable. No clean-console claim is made.',
      'Historical score projection is a policy effect, not a new score for the repaired artifact.'],
  };
}

async function fileManifest(directory) {
  const files = [];
  async function visit(relative) {
    for (const entry of await readdir(path.join(directory, relative), { withFileTypes: true })) {
      const name = path.posix.join(relative, entry.name);
      const info = await lstat(path.join(directory, name));
      requireValue(!info.isSymbolicLink(), 'Symlinks are not valid experiment artifacts');
      if (info.isDirectory()) await visit(name);
      else {
        requireValue(info.isFile(), 'Experiment artifacts must be regular files');
        files.push({ path: name, sha256: hash(await readFile(path.join(directory, name))) });
      }
    }
  }
  await visit('');
  return files.sort((a, b) => a.path.localeCompare(b.path, 'en'));
}

async function main() {
  requireValue(process.argv.length === 3, 'Usage: node compare-probes.mjs <experiment-manifest.json>');
  const filename = path.resolve(process.argv[2]);
  const manifest = JSON.parse(await readFile(filename, 'utf8'));
  requireValue(manifest.format_version === '1.0', 'Unsupported experiment manifest');
  const originalFiles = await fileManifest(path.join(root, 'runs/v1'));
  const originals = Object.fromEntries(originalFiles.map(file => [`runs/v1/${file.path}`, file.sha256]));
  requireValue(Object.keys(originals).length === Object.keys(manifest.original_files).length
    && Object.entries(originals).every(([name, digest]) => manifest.original_files[name] === digest), 'Original run files changed or their inventory differs');
  const runnerHash = hash(await readFile(new URL('./probe-observatory.js', import.meta.url)));
  const profileHash = hash(await readFile(new URL('../assets/probe-terra-xhigh.json', import.meta.url)));
  const reports = {};
  const artifactHashes = [];
  for (const side of ['baseline', 'candidate']) {
    const directory = path.resolve(root, manifest[`${side}_output`]);
    const files = await fileManifest(directory);
    const digest = hash(JSON.stringify(files));
    artifactHashes.push(digest);
    reports[side] = {};
    for (const plan of plans) {
      const reportPath = path.join(path.dirname(filename), 'evidence', `${side}-${plan.key}.json`);
      const reportBytes = await readFile(reportPath);
      const report = JSON.parse(reportBytes);
      if (plan.scenario === 'overview') {
        const visual = manifest.visual_observations.find(item => item.side === side && item.plan === plan.key);
        requireValue(visual?.report_sha256 === hash(reportBytes), `Visual evidence is missing or stale for ${side} ${plan.key}`);
      }
      requireValue(report.provenance?.artifact_sha256 === digest
        && JSON.stringify(report.provenance.files) === JSON.stringify(files), `Artifact identity mismatch in ${reportPath}`);
      requireValue(report.provenance.runner_sha256 === runnerHash && report.provenance.profile_sha256 === profileHash, `Probe version mismatch in ${reportPath}`);
      requireValue(report.served_files?.url === report.url
        && JSON.stringify(report.served_files.before) === JSON.stringify(files)
        && JSON.stringify(report.served_files.after) === JSON.stringify(files), `Served-byte identity missing or mismatched in ${reportPath}`);
      reports[side][plan.key] = report;
    }
  }
  requireValue(artifactHashes[0] !== artifactHashes[1], 'The candidate must differ from the baseline artifact');
  const registry = JSON.parse(await readFile(new URL('../assets/check-registry.json', import.meta.url), 'utf8'));
  const policy = [];
  for (const name of Object.keys(originals).filter(name => name.endsWith('/reviews/review.json')).sort()) {
    const review = JSON.parse(await readFile(path.join(root, name), 'utf8'));
    const projection = projectCurrentPolicy(review, registry);
    policy.push({ source: name, source_sha256: originals[name], published: { lower: review.total_score, upper: review.total_score_upper_bound ?? review.total_score }, projection });
  }
  const behavior = compareProbeRuns(reports.baseline, reports.candidate, manifest.visual_observations);
  process.stdout.write(`${JSON.stringify({ format_version: '1.0', originals_verified: originalFiles.length,
    policy_only: policy, behavior }, null, 2)}\n`);
  if (!behavior.verified) process.exitCode = 1;
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  main().catch(error => { console.error(`compare-probes: ${error.message}`); process.exitCode = 1; });
}
