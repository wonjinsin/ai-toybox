#!/usr/bin/env node
import { spawnSync } from 'node:child_process';
import { accessSync, chmodSync, constants, existsSync, mkdirSync, readFileSync, readdirSync, renameSync, statSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const hookNames = [
  'applypatch-msg', 'pre-applypatch', 'post-applypatch', 'pre-commit', 'pre-merge-commit',
  'prepare-commit-msg', 'commit-msg', 'post-commit', 'pre-rebase', 'post-checkout',
  'post-merge', 'pre-push', 'pre-receive', 'update', 'proc-receive', 'post-receive',
  'post-update', 'reference-transaction', 'push-to-checkout', 'pre-auto-gc',
  'post-rewrite', 'sendemail-validate', 'fsmonitor-watchman', 'p4-changelist',
  'p4-prepare-changelist', 'p4-post-changelist', 'p4-pre-submit', 'post-index-change',
];

function git(args, optional = false) {
  const result = spawnSync('git', args, { encoding: 'utf8' });
  if (optional && result.status === 1) return null;
  if (result.error) throw result.error;
  if (result.status !== 0) throw new Error(result.stderr.trim() || `git exited with ${result.status}`);
  return result.stdout.trimEnd();
}

function shellQuote(value) {
  return `'${value.replaceAll("'", "'\\''")}'`;
}

function replace(file, contents, executable = false) {
  const temporary = `${file}.${process.pid}.tmp`;
  writeFileSync(temporary, contents);
  if (executable) chmodSync(temporary, 0o755);
  renameSync(temporary, file);
}

function main() {
  if (git(['rev-parse', '--is-bare-repository']) === 'true') {
    throw new Error('Install gallery hooks from a non-bare working repository.');
  }
  const root = git(['rev-parse', '--show-toplevel']);
  if (git(['config', '--local', '--bool', '--get', 'extensions.worktreeConfig'], true) === 'true'
      && git(['config', '--worktree', '--get', 'core.hooksPath'], true) !== null) {
    throw new Error('Remove the worktree-specific core.hooksPath override before installing repo-local gallery hooks.');
  }
  const common = git(['rev-parse', '--path-format=absolute', '--git-common-dir']);
  const directory = path.join(common, 'benchmark-gallery-hooks');
  const hooks = path.join(directory, 'hooks');
  const stateFile = path.join(directory, 'state.json');
  const effective = git(['config', '--path', '--get', 'core.hooksPath'], true);
  const previousHooks = effective ?? path.join(common, 'hooks');
  let state;
  if (existsSync(stateFile)) {
    state = JSON.parse(readFileSync(stateFile, 'utf8'));
    if (state.version !== 1 || typeof state.previousHooks !== 'string') {
      throw new Error(`Invalid gallery hook installation state: ${stateFile}`);
    }
    if (path.resolve(root, previousHooks) !== hooks && previousHooks !== state.previousHooks) {
      throw new Error('core.hooksPath changed after installation; restore the gallery hooks path before reinstalling.');
    }
  } else {
    if (path.resolve(root, previousHooks) === hooks) {
      throw new Error(`Missing gallery hook installation state: ${stateFile}`);
    }
    state = { version: 1, previousHooks };
  }
  mkdirSync(hooks, { recursive: true });
  replace(stateFile, `${JSON.stringify(state, null, 2)}\n`);
  const installedDriver = path.join(directory, 'gallery-pre-commit.mjs');
  replace(installedDriver, readFileSync(fileURLToPath(new URL('./gallery-pre-commit.mjs', import.meta.url))));

  const names = new Set(hookNames);
  try {
    for (const name of readdirSync(path.resolve(root, state.previousHooks))) {
      try {
        if (!statSync(path.resolve(root, state.previousHooks, name)).isFile()) continue;
        accessSync(path.resolve(root, state.previousHooks, name), constants.X_OK);
        names.add(name);
      } catch (error) {
        if (error.code !== 'EACCES' && error.code !== 'ENOENT') throw error;
      }
    }
  } catch (error) {
    if (error.code !== 'ENOENT') throw error;
  }
  for (const name of names) {
    const previousPath = path.join(state.previousHooks, name);
    const previous = shellQuote(path.isAbsolute(previousPath) ? previousPath : `./${previousPath}`);
    const invocation = name === 'pre-commit'
      ? `if [ -x "$previous" ]; then\n  "$previous" "$@" || exit "$?"\nfi\nexec ${shellQuote(process.execPath)} ${shellQuote(installedDriver)} "$@"\n`
      : 'if [ -x "$previous" ]; then\n  exec "$previous" "$@"\nfi\nexit 0\n';
    replace(path.join(hooks, name), `#!/bin/sh\nprevious=${previous}\n${invocation}`, true);
  }
  git(['config', '--local', 'core.hooksPath', hooks]);
  console.log(`Installed repository-local gallery hooks in ${hooks}`);
  console.log(`Previous hooks remain chained from ${state.previousHooks}`);
}

try {
  main();
} catch (error) {
  console.error(`Gallery hook installation failed: ${error.message}`);
  process.exitCode = 1;
}
