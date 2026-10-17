#!/usr/bin/env node
import { spawnSync } from 'node:child_process';
import { closeSync, copyFileSync, existsSync, mkdirSync, mkdtempSync, openSync, readFileSync, renameSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';

const project = 'ai-benchmark';
const generated = `${project}/gallery/data.js`;
const generator = `${project}/scripts/generate-gallery.mjs`;

function run(command, args, options = {}) {
  const result = spawnSync(command, args, { encoding: 'utf8', ...options });
  if (result.error) throw result.error;
  if (result.status !== 0) {
    throw new Error(result.stderr?.trim() || `${command} exited with ${result.status ?? result.signal}`);
  }
  return result.stdout;
}

function main() {
  const initialDirectory = process.cwd();
  const root = run('git', ['rev-parse', '--show-toplevel']).trim();
  const gitDirectory = run('git', ['rev-parse', '--absolute-git-dir']).trim();
  const index = process.env.GIT_INDEX_FILE
    ? path.resolve(initialDirectory, process.env.GIT_INDEX_FILE)
    : path.resolve(initialDirectory, run('git', ['rev-parse', '--path-format=absolute', '--git-path', 'index']).trim());
  const env = { ...process.env, GIT_DIR: gitDirectory, GIT_WORK_TREE: root, GIT_INDEX_FILE: index, GIT_OPTIONAL_LOCKS: '0' };
  if (env.GIT_COMMON_DIR) env.GIT_COMMON_DIR = path.resolve(initialDirectory, env.GIT_COMMON_DIR);
  const git = (args, options = {}) => run('git', ['-c', 'core.hooksPath=/dev/null', '-c', 'core.fsmonitor=false', ...args], { cwd: root, env, ...options });
  const changes = () => git(['diff', '--cached', '--name-only', '--no-renames', '-z', '--', `${project}/`]);
  if (!changes()) return;

  const lock = `${index}.lock`;
  let ownsLock = true;
  const lockDescriptor = openSync(lock, 'wx', statSync(index).mode & 0o777);
  closeSync(lockDescriptor);
  let temporary;
  try {
    temporary = mkdtempSync(path.join(tmpdir(), 'benchmark-gallery-index-'));
    const snapshotIndex = path.join(temporary, 'index');
    copyFileSync(index, snapshotIndex);
    env.GIT_INDEX_FILE = snapshotIndex;
    if (!changes()) return;
    const records = git(['ls-files', '--stage', '--full-name', '-z', '--', `${project}/`]).split('\0').filter(Boolean);
    if (records.length === 0) return;
    const entries = records.map(record => {
      const separator = record.indexOf('\t');
      const [mode, oid, stage] = record.slice(0, separator).split(' ');
      const name = record.slice(separator + 1);
      if (stage !== '0') throw new Error(`Resolve staged conflicts before generating the gallery: ${name}`);
      if (mode !== '100644' && mode !== '100755') {
        throw new Error(`Gallery snapshots require regular files, not symlinks or submodules: ${name}`);
      }
      if (!name.startsWith(`${project}/`) || name.split('/').some(part => part === '..' || part === '.git')) {
        throw new Error(`Unsafe staged gallery path: ${name}`);
      }
      return { mode, oid, name };
    });
    if (!entries.some(entry => entry.name === generator)) {
      throw new Error(`The staged project is missing ${generator}; stage the generator or delete the entire project.`);
    }

    const snapshot = path.join(temporary, 'snapshot');
    for (const entry of entries) {
      const target = path.join(snapshot, entry.name);
      mkdirSync(path.dirname(target), { recursive: true });
      const descriptor = openSync(target, 'wx', entry.mode === '100755' ? 0o755 : 0o644);
      try {
        git(['cat-file', 'blob', entry.oid], { stdio: ['ignore', descriptor, 'pipe'] });
      } finally {
        closeSync(descriptor);
      }
    }
    const output = path.join(temporary, 'data.js');
    const generatorEnv = Object.fromEntries(Object.entries(process.env).filter(([key]) => !key.startsWith('GIT_')));
    run(process.execPath, [path.join(snapshot, generator), '--root', path.join(snapshot, project), '--output', output], {
      cwd: snapshot,
      env: generatorEnv,
    });
    if (!existsSync(output) || !statSync(output).isFile()) throw new Error('The staged generator did not produce data.js.');
    const oid = git(['hash-object', '-w', '--no-filters', '--', output]).trim();
    git(['-c', 'core.splitIndex=false', 'update-index', '--no-split-index', '--add', '--cacheinfo', `100644,${oid},${generated}`]);
    writeFileSync(lock, readFileSync(snapshotIndex));
    renameSync(lock, index);
    ownsLock = false;
  } finally {
    if (ownsLock) rmSync(lock, { force: true });
    if (temporary) rmSync(temporary, { recursive: true, force: true });
  }
}

try {
  main();
} catch (error) {
  console.error(`Gallery pre-commit failed: ${error.message}`);
  process.exitCode = 1;
}
