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

const labels = {
  complete: '완료', partial: '부분 완료', blocked: '진행 불가',
  pass: '통과', fail: '실패', unverified: '미검증',
  http: 'HTTP', static: '정적 검사', none: '없음',
  simulation: '시뮬레이션', selection: '선택 및 추적', usability: '사용성',
  visual: '시각적 완성도', responsive: '반응형', delivery: '실행 및 제공',
};

function markdown(review, registry) {
  const value = item => item === null ? '미산정' : literal(item);
  const label = item => labels[item];
  const current = review.format_version === '5.0';
  const bounded = current || review.format_version === '4.0';
  const range = (lower, upper) => lower === upper ? lower.toFixed(2) : `${lower.toFixed(2)} ~ ${upper.toFixed(2)}`;
  const definitions = registry.rubrics[review.rubric_version].checks;
  const missingChecks = review.checks.filter(check => check.status === 'unverified').map(check => check.id);
  const lines = [
    '# 벤치마크 리뷰', '', '## 결론', '',
    `상태: ${label(review.status)}.`,
    `필수 요구사항 판정: ${label(review.requirements_verdict)}.`,
    current
      ? `품질 총점: ${review.total_score.toFixed(2)} / 100. ${review.total_score === review.total_score_upper_bound ? '확정 점수' : '보수적 하한'}. 범위: ${range(review.total_score, review.total_score_upper_bound)}.`
      : bounded
        ? `최종 총점: ${review.total_score.toFixed(2)} / 100. ${review.total_score === review.total_score_upper_bound ? '확정 점수' : '보수적 하한'}. 범위: ${range(review.total_score, review.total_score_upper_bound)}.`
        : `최종 총점: ${value(review.total_score)}.`,
  ];
  if (current) {
    const mandatory = status => definitions
      .filter((definition, index) => definition.mandatory && review.checks[index].status === status)
      .map(definition => definition.id);
    const failed = mandatory('fail');
    const unknown = mandatory('unverified');
    lines.push('품질 점수에는 상한 제한을 적용하지 않습니다. 필수 요구사항 판정은 품질 점수와 독립적입니다.',
      `필수 항목 실패 (${failed.length}): ${failed.length ? failed.join(', ') : '없음'}.`,
      `필수 항목 미검증 (${unknown.length}): ${unknown.length ? unknown.join(', ') : '없음'}.`);
  }
  if (bounded && !current) {
    lines.push(`상한 적용 전 총점: ${range(review.raw_total_score, review.raw_total_score_upper_bound)}.`,
      `확정된 상한: ${review.score_cap === null ? '없음' : review.score_cap}.`);
    if (review.score_cap !== null) lines.push('확정된 69점 상한은 필수 항목 실패에 적용하는 기존 리뷰 정책이며, 보정된 품질 측정값이 아닙니다.');
    else if (definitions.some((definition, index) => definition.mandatory && review.checks[index].status === 'unverified')) {
      lines.push('미검증 필수 항목은 실패할 수 있습니다. 하한에는 가능한 69점 상한을 반영했지만, 실패나 상한이 확정된 것은 아닙니다.');
    }
  }
  if (bounded) {
    const missingMetadata = [
      ...(review.reviewed_at === null ? ['reviewed_at'] : []),
      ...Object.entries(review.reviewer).filter(([, item]) => item === null).map(([field]) => `reviewer.${field}`),
    ];
    lines.push(`미검증 항목 (${missingChecks.length}): ${missingChecks.length ? missingChecks.join(', ') : '없음'}.`,
      `누락된 리뷰어 메타데이터: ${missingMetadata.length ? missingMetadata.join(', ') : '없음'}.`);
  }
  lines.push(
    `형식: ${literal(review.format_version)}. 평가 기준: ${literal(review.rubric_version)}.`,
    `리뷰 시각: ${review.reviewed_at === null ? '확인 불가' : literal(review.reviewed_at)}.`,
    `리뷰어 모델: ${review.reviewer.model === null ? '확인 불가' : literal(review.reviewer.model)}.`,
    `리뷰어 추론 수준: ${review.reviewer.effort === null ? '확인 불가' : literal(review.reviewer.effort)}.`,
    '', '## 점수', '', '| 평가 영역 | 점수 / 5 | 가중치 | 근거 |', '| --- | --- | --- | --- |',
  );
  for (const category of review.categories) {
    const ids = definitions.filter(check => check.category === category.id).map(check => check.id);
    const categoryScore = bounded ? range(category.score, category.score_upper_bound) : category.score === null ? '미검증' : category.score;
    const basis = bounded
      ? category.score === category.score_upper_bound ? '확정 점수; 전체 영역 항목을 분모로 사용' : '보수적 하한; 미검증 평점은 0~5 범위이며 전체 영역 항목을 분모로 사용'
      : category.score === null ? '근거 불충분' : '관찰된 평점의 평균';
    lines.push(`| ${label(category.id)} (${literal(category.id)}) | ${categoryScore} | ${category.weight} | ${ids.join(', ')}; ${basis} |`);
  }
  lines.push('', '## 발견 사항', '');
  for (const [status, heading] of [['fail', '확인된 실패'], ['unverified', '검증하지 못한 항목'], ['pass', '검증된 통과 항목']]) {
    lines.push(`### ${heading}`, '');
    const checks = review.checks.filter(check => check.status === status);
    if (checks.length === 0) lines.push('없음.', '');
    for (const check of checks) {
      lines.push(`#### ${literal(check.id)}`, '',
        `상태: ${label(check.status)}. 평점: ${check.rating === null ? '미검증' : check.rating}. 방법: ${label(check.method)}.`, '',
        literal(check.evidence), '');
    }
  }
  lines.push('## 검증 및 한계', '',
    `브라우저 검증 방식: ${label(review.verification.browser_method)}.`,
    `뷰포트: ${review.verification.viewports.length ? review.verification.viewports.map(literal).join(', ') : '관찰하지 않음'}.`,
    `검증 범위: 통과 ${review.coverage.passed}개, 실패 ${review.coverage.failed}개, 미검증 ${review.coverage.unverified}개.`, '');
  if (review.verification.browser_method === 'none') lines.push('브라우저를 실행하지 않았습니다.', '');
  lines.push('### 참고 사항', '');
  if (review.notes.length === 0) lines.push('없음.');
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
