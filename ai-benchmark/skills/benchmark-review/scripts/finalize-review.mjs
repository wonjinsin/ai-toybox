#!/usr/bin/env node
import { constants } from 'node:fs';
import { lstat, mkdir, open, readFile, realpath, rename, unlink } from 'node:fs/promises';
import { isAbsolute, join, relative, resolve, sep } from 'node:path';
import { randomUUID } from 'node:crypto';
import { normalize } from './score-review.mjs';

function requireValue(condition, message) {
  if (!condition) throw new Error(message);
}

function inside(root, path) {
  const suffix = relative(root, path);
  return suffix === '' || (!isAbsolute(suffix) && suffix !== '..' && !suffix.startsWith(`..${sep}`));
}

async function inspect(path) {
  try {
    return await lstat(path);
  } catch (error) {
    if (error.code === 'ENOENT') return null;
    throw error;
  }
}

function literal(value) {
  return String(value).replace(/[&<>"'\\`*_{}\[\]()#+.!|~\-\r\n\u0000-\u001f\u007f]/g, character =>
    character === '\n' ? '<br>' : `&#${character.charCodeAt(0)};`);
}

function markdown(review, registry) {
  const value = item => item === null ? 'Unscored' : literal(item);
  const current = review.format_version === '5.0';
  const bounded = current || review.format_version === '4.0';
  const range = (lower, upper) => lower === upper ? lower.toFixed(2) : `${lower.toFixed(2)} to ${upper.toFixed(2)}`;
  const definitions = registry.rubrics[review.rubric_version].checks;
  const missingChecks = review.checks.filter(check => check.status === 'unverified').map(check => check.id);
  const lines = [
    '# Benchmark review', '', '## Conclusion', '',
    `Status: ${literal(review.status)}.`,
    `Mandatory verdict: ${literal(review.requirements_verdict)}.`,
    current
      ? `Quality total: ${review.total_score.toFixed(2)} / 100. ${review.total_score === review.total_score_upper_bound ? 'Determined score' : 'Conservative lower bound'}. Range: ${range(review.total_score, review.total_score_upper_bound)}.`
      : bounded
        ? `Published total: ${review.total_score.toFixed(2)} / 100. ${review.total_score === review.total_score_upper_bound ? 'Determined score' : 'Conservative lower bound'}. Range: ${range(review.total_score, review.total_score_upper_bound)}.`
        : `Published total: ${value(review.total_score)}.`,
  ];
  if (current) {
    const mandatory = status => definitions
      .filter((definition, index) => definition.mandatory && review.checks[index].status === status)
      .map(definition => definition.id);
    const failed = mandatory('fail');
    const unknown = mandatory('unverified');
    lines.push('Quality is uncapped. The mandatory verdict is independent of the quality score.',
      `Mandatory failures (${failed.length}): ${failed.length ? failed.join(', ') : 'None'}.`,
      `Mandatory unverified (${unknown.length}): ${unknown.length ? unknown.join(', ') : 'None'}.`);
  }
  if (bounded && !current) {
    lines.push(`Raw total: ${range(review.raw_total_score, review.raw_total_score_upper_bound)}.`,
      `Confirmed cap: ${review.score_cap === null ? 'None' : review.score_cap}.`);
    if (review.score_cap !== null) lines.push('The confirmed 69 cap is existing review policy for a mandatory failure, not a calibrated quality measurement.');
    else if (definitions.some((definition, index) => definition.mandatory && review.checks[index].status === 'unverified')) {
      lines.push('An unverified mandatory check may fail. The lower endpoint includes the possible 69 cap; no failure or cap is confirmed.');
    }
  }
  if (bounded) {
    const missingMetadata = [
      ...(review.reviewed_at === null ? ['reviewed_at'] : []),
      ...Object.entries(review.reviewer).filter(([, item]) => item === null).map(([field]) => `reviewer.${field}`),
    ];
    lines.push(`Missing checks (${missingChecks.length}): ${missingChecks.length ? missingChecks.join(', ') : 'None'}.`,
      `Missing reviewer metadata: ${missingMetadata.length ? missingMetadata.join(', ') : 'None'}.`);
  }
  lines.push(
    `Format: ${literal(review.format_version)}. Rubric: ${literal(review.rubric_version)}.`,
    `Reviewed at: ${review.reviewed_at === null ? 'Unavailable' : literal(review.reviewed_at)}.`,
    `Reviewer model: ${review.reviewer.model === null ? 'Unavailable' : literal(review.reviewer.model)}.`,
    `Reviewer effort: ${review.reviewer.effort === null ? 'Unavailable' : literal(review.reviewer.effort)}.`,
    '', '## Scores', '', '| Category | Score / 5 | Weight | Basis |', '| --- | --- | --- | --- |',
  );
  for (const category of review.categories) {
    const ids = definitions.filter(check => check.category === category.id).map(check => check.id);
    const categoryScore = bounded ? range(category.score, category.score_upper_bound) : category.score === null ? 'Unverified' : category.score;
    const basis = bounded
      ? category.score === category.score_upper_bound ? 'determined score; full category denominator' : 'conservative lower bound; unknown ratings range from 0 to 5 with the full category denominator'
      : category.score === null ? 'incomplete evidence' : 'mean of observed ratings';
    lines.push(`| ${literal(category.id)} | ${categoryScore} | ${category.weight} | ${ids.join(', ')}; ${basis} |`);
  }
  lines.push('', '## Findings', '');
  for (const [status, heading] of [['fail', 'Confirmed failures'], ['unverified', 'Missing proof'], ['pass', 'Verified passes']]) {
    lines.push(`### ${heading}`, '');
    const checks = review.checks.filter(check => check.status === status);
    if (checks.length === 0) lines.push('None.', '');
    for (const check of checks) {
      lines.push(`#### ${literal(check.id)}`, '',
        `Status: ${literal(check.status)}. Rating: ${check.rating === null ? 'Unverified' : check.rating}. Method: ${literal(check.method)}.`, '',
        literal(check.evidence), '');
    }
  }
  lines.push('## Verification and limits', '',
    `Browser method: ${literal(review.verification.browser_method)}.`,
    `Viewports: ${review.verification.viewports.length ? review.verification.viewports.map(literal).join(', ') : 'None observed'}.`,
    `Coverage: ${review.coverage.passed} passed, ${review.coverage.failed} failed, ${review.coverage.unverified} unverified.`, '');
  if (review.verification.browser_method === 'none') lines.push('No browser execution occurred.', '');
  lines.push('### Notes', '');
  if (review.notes.length === 0) lines.push('None.');
  for (const note of review.notes) lines.push(`- ${literal(note)}`);
  return `${lines.join('\n')}\n`;
}

const result = { ok: false, outputs: { 'review.json': 'unknown', 'review.md': 'unknown' } };
const temporary = new Set();
let run;
let directory;
let runIdentity;
let directoryIdentity;

async function safeDirectory() {
  const target = await lstat(run);
  requireValue(target.isDirectory() && !target.isSymbolicLink() && target.dev === runIdentity.dev && target.ino === runIdentity.ino && await realpath(run) === run, 'target directory changed or is unsafe');
  const stats = await lstat(directory);
  requireValue(stats.isDirectory() && !stats.isSymbolicLink() && await realpath(directory) === directory, 'reviews must be a real directory inside the target');
  if (directoryIdentity) requireValue(stats.dev === directoryIdentity.dev && stats.ino === directoryIdentity.ino, 'reviews directory changed');
  directoryIdentity = stats;
}

async function safeOutput(name) {
  await safeDirectory();
  const path = join(directory, name);
  const stats = await inspect(path);
  requireValue(!stats || (stats.isFile() && !stats.isSymbolicLink() && stats.nlink === 1), `${name} must be a regular, non-linked file or absent`);
  if (stats) requireValue(await realpath(path) === path, `${name} escapes the output directory`);
  return stats;
}

try {
  requireValue(process.argv.length === 4, 'usage: node skills/benchmark-review/scripts/finalize-review.mjs <run-directory> <draft-review.json>');
  const requestedRun = resolve(process.argv[2]);
  const target = await lstat(requestedRun);
  requireValue(target.isDirectory() && !target.isSymbolicLink(), 'run-directory must be an existing non-symlink directory');
  run = await realpath(requestedRun);
  runIdentity = target;
  directory = join(run, 'reviews');
  result.paths = { 'review.json': join(directory, 'review.json'), 'review.md': join(directory, 'review.md') };
  const requestedDraft = resolve(process.argv[3]);
  const draftStats = await lstat(requestedDraft);
  requireValue(draftStats.isFile() && !draftStats.isSymbolicLink() && draftStats.nlink === 1, 'draft must be a regular, non-linked JSON file');
  const draft = await realpath(requestedDraft);
  requireValue(!inside(run, draft), 'draft must be outside the target run directory');
  const registry = JSON.parse(await readFile(new URL('../assets/check-registry.json', import.meta.url), 'utf8'));
  const normalized = normalize(JSON.parse(await readFile(draft, 'utf8')), registry);
  if (normalized.status === 'complete') {
    const output = await lstat(join(run, 'output'));
    requireValue(output.isDirectory() && !output.isSymbolicLink(), 'complete review requires a real output directory');
    for (const name of ['prompt.md', 'run.json', 'output/index.html']) {
      const source = join(run, name);
      const stats = await lstat(source);
      requireValue(stats.isFile() && !stats.isSymbolicLink() && stats.size > 0, `complete review requires a regular nonempty ${name}`);
    }
    const sourceRecord = JSON.parse(await readFile(join(run, 'run.json'), 'utf8'));
    requireValue(sourceRecord !== null && typeof sourceRecord === 'object' && !Array.isArray(sourceRecord), 'complete review requires a run JSON object');
  }
  const json = `${JSON.stringify(normalized, null, 2)}\n`;
  const report = markdown(normalized, registry);
  requireValue(JSON.stringify(normalize(JSON.parse(json), registry)) === JSON.stringify(normalized), 'normalized JSON is not stable');
  const contents = { 'review.json': json, 'review.md': report };
  const existingDirectory = await inspect(directory);
  if (!existingDirectory) await mkdir(directory);
  await safeDirectory();
  for (const name of Object.keys(contents)) result.outputs[name] = await safeOutput(name) ? 'unchanged' : 'missing';
  const staged = {};
  for (const [name, content] of Object.entries(contents)) {
    await safeDirectory();
    const path = join(directory, `.${name}.${randomUUID()}.tmp`);
    const handle = await open(path, constants.O_WRONLY | constants.O_CREAT | constants.O_EXCL | constants.O_NOFOLLOW, 0o600);
    temporary.add(path);
    try {
      await handle.writeFile(content, 'utf8');
      await handle.sync();
    } finally {
      await handle.close();
    }
    staged[name] = path;
  }
  for (const name of Object.keys(contents)) await safeOutput(name);
  for (const name of Object.keys(contents)) {
    await safeOutput(name);
    await rename(staged[name], join(directory, name));
    temporary.delete(staged[name]);
    result.outputs[name] = 'saved';
  }
  for (const [name, expected] of Object.entries(contents)) {
    await safeOutput(name);
    requireValue(await readFile(join(directory, name), 'utf8') === expected, `${name} readback does not match published bytes`);
  }
  const saved = normalize(JSON.parse(await readFile(join(directory, 'review.json'), 'utf8')), registry);
  requireValue(JSON.stringify(saved) === JSON.stringify(normalized) && markdown(saved, registry) === report, 'saved pair failed normalization or Markdown consistency');
  result.ok = true;
} catch (error) {
  result.error = error.message;
  process.stderr.write(`finalize-review: ${error.message}\n`);
  process.exitCode = 1;
} finally {
  for (const path of temporary) {
    try {
      await safeDirectory();
      await unlink(path);
    } catch (error) {
      result.cleanup_errors ??= [];
      result.cleanup_errors.push({ path, error: error.message });
      result.ok = false;
      process.exitCode = 1;
    }
  }
  process.stdout.write(`${JSON.stringify(result, null, 2)}\n`);
}
