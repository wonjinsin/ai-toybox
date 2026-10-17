#!/usr/bin/env node
import { readFile } from 'node:fs/promises';

function requireValue(condition, message) {
  if (!condition) throw new Error(message);
}

function text(value) {
  return typeof value === 'string' && value.trim().length > 0;
}

function validate(run) {
  requireValue(run !== null && typeof run === 'object' && !Array.isArray(run), 'run must be an object');
  const metadata = ['session_id', 'model', 'effort'];
  const measurements = ['duration_ms', 'input_tokens', 'output_tokens'];
  const fields = ['format_version', 'status', ...metadata, ...measurements, 'notes'];
  for (const field of fields) {
    requireValue(Object.hasOwn(run, field), `${field} is required`);
  }
  for (const field of Object.keys(run)) {
    requireValue(fields.includes(field), `unexpected field: ${field}`);
  }
  requireValue(run.format_version === '2.1', 'format_version must be 2.1');
  requireValue(['running', 'completed', 'failed', 'interrupted'].includes(run.status), 'invalid status');
  requireValue(Array.isArray(run.notes) && run.notes.every(text), 'notes must be an array of nonempty strings');

  const completed = run.status === 'completed';
  const missing = [];
  for (const field of [...metadata, ...measurements]) {
    const value = run[field];
    if (value === null) {
      requireValue(!completed, `${field} must not be null for a completed run`);
      missing.push(field);
      continue;
    }
    if (metadata.includes(field)) {
      requireValue(text(value), `${field} must be a concrete nonempty string or null`);
      requireValue(!(field !== 'effort' && /^none$/i.test(value.trim()))
        && !/^(unknown|null|undefined|unavailable|n\/a|default|auto|unspecified|unset|not available|not applicable|tbd|todo|placeholder)$/i.test(value.trim())
        && !/\$\{[^}]*\}/.test(value), `${field} must not be a placeholder`);
    } else if (field === 'duration_ms') {
      requireValue(typeof value === 'number' && Number.isFinite(value) && value >= 0, 'duration_ms must be a finite nonnegative number or null');
    } else {
      requireValue(Number.isSafeInteger(value) && value >= 0, `${field} must be a nonnegative safe integer or null`);
    }
  }
  requireValue(missing.length === 0 || run.notes.length > 0, `notes must explain missing values: ${missing.join(', ')}`);
}

try {
  requireValue(process.argv.length === 3, 'Usage: node validate-run.mjs <run.json>');
  const source = await readFile(process.argv[2], 'utf8');
  validate(JSON.parse(source));
  process.stdout.write(source);
} catch (error) {
  console.error(error.message);
  process.exitCode = 1;
}
