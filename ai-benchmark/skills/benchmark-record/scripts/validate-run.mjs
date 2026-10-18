#!/usr/bin/env node
import { constants } from 'node:fs';
import { lstat, open, readFile, readdir, realpath, rename, unlink } from 'node:fs/promises';
import { createHash, randomUUID } from 'node:crypto';
import path from 'node:path';

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

const digest = source => createHash('sha256').update(source).digest('hex');

async function safePath(filename, directory = false) {
  const absolute = path.resolve(filename);
  requireValue(!(await lstat(absolute)).isSymbolicLink(), `symlink is not allowed: ${absolute}`);
  const canonical = path.join(await realpath(path.dirname(absolute)), path.basename(absolute));
  const stat = await lstat(canonical);
  requireValue(!stat.isSymbolicLink(), `symlink is not allowed: ${canonical}`);
  requireValue(directory ? stat.isDirectory() : stat.isFile(), `unexpected path type: ${canonical}`);
  if (!directory) requireValue(stat.nlink === 1, `hard-linked file is not allowed: ${canonical}`);
  return canonical;
}

async function readRegular(filename) {
  const canonical = await safePath(filename);
  const handle = await open(canonical, constants.O_RDONLY | constants.O_NOFOLLOW);
  try {
    const stat = await handle.stat();
    requireValue(stat.isFile() && stat.nlink === 1, `unsafe file: ${filename}`);
    return await handle.readFile('utf8');
  } finally {
    await handle.close();
  }
}


async function publish(mode, proposalPath, runPath, expected) {
  const directory = await safePath(runPath, true);
  requireValue(directory !== path.parse(directory).root, 'run directory must not be a filesystem root');
  const proposal = await safePath(proposalPath);
  const relative = path.relative(directory, proposal);
  requireValue(relative === '..' || relative.startsWith(`..${path.sep}`) || path.isAbsolute(relative),
    'proposal must be outside the run directory');
  const source = await readRegular(proposal);
  const run = JSON.parse(source);
  validate(run);
  requireValue(mode === '--checkpoint' ? run.status === 'running' : run.status !== 'running',
    mode === '--checkpoint' ? 'checkpoint requires running status' : 'finalization requires terminal status');
  if (expected !== undefined) requireValue(/^[a-f0-9]{64}$/.test(expected), 'expected SHA-256 must be 64 lowercase hex characters');
  requireValue(mode !== '--finalize' || expected !== undefined, 'finalization requires expected SHA-256');

  const target = path.join(directory, 'run.json');
  const temporary = path.join(directory, `.run-${randomUUID()}.tmp`);
  let temporaryCreated = false;
  try {
    let original;
    if (expected !== undefined) {
      original = await readRegular(target);
      requireValue(digest(original) === expected, 'existing record changed or expected SHA-256 is incorrect');
      const previous = JSON.parse(original);
      validate(previous);
      requireValue(previous.status === 'running' || (mode === '--checkpoint' && previous.status === 'interrupted'),
        'cannot overwrite a terminal historical record');
      requireValue(previous.session_id === null || run.session_id === previous.session_id,
        'original session_id must be preserved');
    } else {
      requireValue((await readdir(directory)).length === 0,
        'run directory changed before checkpoint');
    }
    if (run.status === 'completed') {
      await safePath(path.join(directory, 'output'), true);
      requireValue((await readRegular(path.join(directory, 'prompt.md'))).trim().length > 0,
        'completed run requires prompt.md');
      requireValue((await readRegular(path.join(directory, 'output', 'index.html'))).trim().length > 0,
        'completed run requires output/index.html');
    }
    const handle = await open(temporary, 'wx', 0o600);
    temporaryCreated = true;
    try {
      await handle.writeFile(source, 'utf8');
      await handle.sync();
    } finally {
      await handle.close();
    }
    validate(JSON.parse(await readRegular(temporary)));
    await safePath(directory, true);
    if (original !== undefined) {
      requireValue(await readRegular(target) === original, 'existing record changed during publication');
    } else {
      requireValue(!(await readdir(directory)).includes('run.json'), 'record appeared during publication');
    }
    await rename(temporary, target);
    temporaryCreated = false;
    const saved = await readRegular(target);
    validate(JSON.parse(saved));
    requireValue(saved === source, 'published record read-back mismatch');
    console.error(`Published SHA-256: ${digest(saved)}`);
    return saved;
  } finally {
    if (temporaryCreated) await unlink(temporary);
  }
}

try {
  const args = process.argv.slice(2);
  let source;
  if (args.length === 1) {
    source = await readFile(args[0], 'utf8');
    validate(JSON.parse(source));
  } else {
    const [mode, proposal, directory, expected] = args;
    requireValue((mode === '--checkpoint' && (args.length === 3 || args.length === 4))
      || (mode === '--finalize' && args.length === 4),
    'Usage: node validate-run.mjs <run.json> | --checkpoint <proposal.json> <run-directory> [<expected-sha256>] | --finalize <proposal.json> <run-directory> <expected-sha256>');
    source = await publish(mode, proposal, directory, expected);
  }
  process.stdout.write(source);
} catch (error) {
  console.error(error.message);
  process.exitCode = 1;
}
