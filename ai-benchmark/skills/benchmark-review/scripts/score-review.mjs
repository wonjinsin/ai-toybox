#!/usr/bin/env node
import { readFile } from 'node:fs/promises';

function requireValue(condition, message) {
  if (!condition) throw new Error(message);
}

function object(value, keys, path) {
  requireValue(value !== null && typeof value === 'object' && !Array.isArray(value), `${path} must be an object`);
  for (const key of keys) requireValue(Object.hasOwn(value, key), `${path}.${key} is required`);
  requireValue(Object.keys(value).length === keys.length, `${path} must have exactly: ${keys.join(', ')}`);
}

function text(value) {
  return typeof value === 'string' && value.trim().length > 0;
}

function metadata(value, field) {
  if (!text(value)) return false;
  const normalized = value.trim();
  return !(field === 'model' && /^none$/i.test(normalized))
    && !/^(?:unknown|null|undefined|unavailable|n\/a|default|auto|unspecified|unset|not available|not applicable|tbd|todo|placeholder)$/i.test(normalized)
    && !/\$\{[^}]*\}/.test(value);
}

function timestamp(value) {
  if (typeof value !== 'string' || !/^\d{4}-\d{2}-\d{2}T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/.test(value) || !Number.isFinite(Date.parse(value))) return false;
  const date = value.slice(0, 10);
  return new Date(`${date}T00:00:00Z`).toISOString().slice(0, 10) === date;
}

function score(value, maximum) {
  return value === null || (typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= maximum);
}

function round(value) {
  return Math.round((value + Number.EPSILON) * 100) / 100;
}

function normalize(review, registry) {
  requireValue(review !== null && typeof review === 'object' && !Array.isArray(review), 'review must be an object');
  requireValue(typeof review.rubric_version === 'string' && Object.hasOwn(registry.rubrics, review.rubric_version), 'unsupported rubric_version');
  const rubric = registry.rubrics[review.rubric_version];
  requireValue(review.format_version === rubric.format_version, 'format_version and rubric_version do not match');
  const current = review.format_version === '3.0';
  const fields = ['format_version', 'rubric_version', 'reviewed_at', 'reviewer', 'status', 'requirements_verdict', 'verification', 'categories', 'checks', 'total_score', 'coverage', 'notes'];
  if (current) fields.push('raw_total_score', 'score_cap');
  object(review, fields, 'review');
  requireValue(['complete', 'partial', 'blocked'].includes(review.status), 'invalid status');
  requireValue(['pass', 'fail', 'unverified'].includes(review.requirements_verdict), 'invalid requirements_verdict');
  requireValue(review.reviewed_at === null || timestamp(review.reviewed_at), 'reviewed_at must be a valid ISO timestamp with timezone or null');
  if (review.status === 'complete') requireValue(review.reviewed_at !== null, 'complete review requires reviewed_at to be an actual ISO timestamp with timezone');
  object(review.reviewer, ['model', 'effort'], 'reviewer');
  for (const field of ['model', 'effort']) {
    const value = review.reviewer[field];
    requireValue(value === null || metadata(value, field), `reviewer.${field} must be a concrete nonempty string or null, not a placeholder`);
    if (review.status === 'complete') requireValue(value !== null, `complete review requires reviewer.${field}`);
  }
  requireValue(Array.isArray(review.notes) && review.notes.every(text), 'notes must be an array of nonempty strings');
  if (review.reviewed_at === null || Object.values(review.reviewer).includes(null)) {
    requireValue(review.notes.length > 0, 'missing metadata requires a note');
  }
  object(review.verification, ['browser_method', 'viewports'], 'verification');
  requireValue(['http', 'none'].includes(review.verification.browser_method), 'invalid browser_method');
  const viewports = review.verification.viewports;
  requireValue(Array.isArray(viewports) && viewports.every(value => typeof value === 'string' && /^[1-9]\d*x[1-9]\d*$/.test(value)) && new Set(viewports).size === viewports.length, 'viewports must be unique CSS-pixel dimensions');
  requireValue(review.verification.browser_method !== 'none' || viewports.length === 0, 'browser_method none cannot have viewports');
  requireValue(Array.isArray(review.categories) && review.categories.length === registry.categories.length, 'categories must contain the six registered categories');
  registry.categories.forEach((definition, index) => {
    const category = review.categories[index];
    object(category, ['id', 'weight', 'score'], `categories[${index}]`);
    requireValue(category.id === definition.id && category.weight === definition.weight, `invalid category ID, order or weight at ${index}`);
    requireValue(score(category.score, 5), `invalid category score for ${category.id}`);
  });
  requireValue(score(review.total_score, 100), 'invalid total_score');
  if (current) {
    requireValue(score(review.raw_total_score, 100), 'invalid raw_total_score');
    requireValue(review.score_cap === null || review.score_cap === rubric.mandatory_failure_cap, 'invalid score_cap');
  }
  object(review.coverage, ['passed', 'failed', 'unverified'], 'coverage');
  requireValue(Object.values(review.coverage).every(value => Number.isInteger(value) && value >= 0 && value <= rubric.checks.length), 'invalid coverage counts');
  requireValue(Object.values(review.coverage).reduce((sum, count) => sum + count, 0) === rubric.checks.length, 'coverage must sum to the registered check count');
  requireValue(Array.isArray(review.checks) && review.checks.length === rubric.checks.length, `checks must contain exactly ${rubric.checks.length} registered checks`);
  const coverage = { passed: 0, failed: 0, unverified: 0 };
  const categoryRatings = new Map(registry.categories.map(category => [category.id, []]));
  let mandatoryFailure = false;
  let mandatoryUnknown = false;
  rubric.checks.forEach((definition, index) => {
    const check = review.checks[index];
    object(check, ['id', 'status', 'rating', 'method', 'evidence'], `checks[${index}]`);
    requireValue(check.id === definition.id, `invalid check ID or order at ${index}; expected ${definition.id}`);
    requireValue(['pass', 'fail', 'unverified'].includes(check.status), `invalid status for ${check.id}`);
    requireValue(['http', 'static', 'none'].includes(check.method), `invalid method for ${check.id}`);
    requireValue(text(check.evidence), `nonempty evidence is required for ${check.id}`);
    if (check.status === 'unverified') {
      requireValue(check.rating === null, `unverified ${check.id} must have a null rating`);
      coverage.unverified += 1;
    } else {
      requireValue(Number.isInteger(check.rating) && check.rating >= 0 && check.rating <= 5, `invalid rating for ${check.id}`);
      const expectedStatus = definition.qualitative ? (check.rating >= 3 ? 'pass' : 'fail') : (check.rating === 5 ? 'pass' : 'fail');
      requireValue((definition.qualitative || check.rating === 0 || check.rating === 5) && check.status === expectedStatus, `rating and status disagree for ${check.id}`);
      requireValue(check.method === definition.method, `${check.id} requires ${definition.method} evidence`);
      if (current) {
        const sections = /^Action:\s*([\s\S]+?)\s+Expected:\s*([\s\S]+?)\s+Observed:\s*([\s\S]+)$/i.exec(check.evidence.trim());
        requireValue(sections && sections.slice(1).every(text), `${check.id} evidence requires Action:, Expected:, and Observed: sections`);
        requireValue(!/^(?:not observed|unverified|unknown|todo|tbd|placeholder)\b/i.test(sections[3].trim()), `${check.id} has no observed evidence`);
      }
      coverage[check.status === 'pass' ? 'passed' : 'failed'] += 1;
    }
    if (check.method === 'http') requireValue(review.verification.browser_method === 'http', `${check.id} HTTP evidence requires browser_method http`);
    if (check.status === 'pass' && ['L1', 'L2'].includes(check.id)) {
      const viewport = check.id === 'L1' ? '1440x900' : '390x844';
      requireValue(viewports.includes(viewport), `${check.id} requires observed viewport ${viewport}`);
    }
    mandatoryFailure ||= definition.mandatory && check.status === 'fail';
    mandatoryUnknown ||= definition.mandatory && check.status === 'unverified';
    categoryRatings.get(definition.category).push(check.rating);
  });
  const status = review.status === 'blocked' ? 'blocked' : coverage.unverified > 0 ? 'partial' : 'complete';
  requireValue(review.status === status, 'status is incompatible with check coverage');
  if (status === 'blocked') {
    requireValue(coverage.unverified === rubric.checks.length && review.notes.length > 0, 'blocked reviews require all checks unverified and a blocker note');
    requireValue(review.total_score === null && review.categories.every(category => category.score === null) && review.requirements_verdict === 'unverified', 'blocked reviews cannot retain scores or a verdict');
    requireValue(review.coverage.passed === 0 && review.coverage.failed === 0 && review.coverage.unverified === rubric.checks.length, 'blocked reviews cannot retain coverage');
    if (current) requireValue(review.raw_total_score === null && review.score_cap === null, 'blocked reviews cannot retain raw scores or a cap');
  }
  const verdict = mandatoryFailure ? 'fail' : mandatoryUnknown ? 'unverified' : 'pass';
  let weightedTotal = 0;
  const categories = registry.categories.map(definition => {
    const ratings = categoryRatings.get(definition.id);
    const mean = ratings.includes(null) ? null : ratings.reduce((sum, rating) => sum + rating, 0) / ratings.length;
    if (mean !== null) weightedTotal += mean / 5 * definition.weight;
    return { ...definition, score: mean === null ? null : round(mean) };
  });
  const raw = coverage.unverified > 0 ? null : round(weightedTotal);
  const cap = mandatoryFailure ? rubric.mandatory_failure_cap : null;
  if (status === 'partial') {
    requireValue(review.total_score === null && (!current || review.raw_total_score === null), 'partial reviews cannot publish numeric totals');
    categories.forEach((category, index) => {
      requireValue(category.score !== null || review.categories[index].score === null, `unverified category ${category.id} cannot publish a score`);
    });
  }
  return {
    ...review,
    status,
    requirements_verdict: verdict,
    categories,
    ...(current ? { raw_total_score: raw, score_cap: cap } : {}),
    total_score: raw === null ? null : cap === null ? raw : Math.min(raw, cap),
    coverage,
  };
}

try {
  requireValue(process.argv.length === 3, 'usage: node skills/benchmark-review/scripts/score-review.mjs <review.json>');
  const registry = JSON.parse(await readFile(new URL('../assets/check-registry.json', import.meta.url), 'utf8'));
  const review = JSON.parse(await readFile(process.argv[2], 'utf8'));
  process.stdout.write(`${JSON.stringify(normalize(review, registry), null, 2)}\n`);
} catch (error) {
  process.stderr.write(`score-review: ${error.message}\n`);
  process.exitCode = 1;
}
