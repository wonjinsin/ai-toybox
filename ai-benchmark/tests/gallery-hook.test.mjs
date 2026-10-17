import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { chmod, copyFile, mkdir, mkdtemp, readFile, realpath, rm, symlink, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const scripts = fileURLToPath(new URL('../scripts/', import.meta.url));
const driver = path.join(scripts, 'gallery-pre-commit.mjs');
const installer = path.join(scripts, 'install-gallery-hooks.mjs');
const project = 'ai-benchmark';
const generated = `${project}/gallery/data.js`;
const generator = `${project}/scripts/generate-gallery.mjs`;
const modulePath = `${project}/gallery/data.mjs`;
const source = `${project}/runs/v1/model 한글 space/run.json`;
const fixtureGenerator = `import { readFile, readdir, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { marker } from '../gallery/data.mjs';
const root = process.argv[process.argv.indexOf('--root') + 1];
const output = process.argv[process.argv.indexOf('--output') + 1];
const runs = [];
for (const name of (await readdir(path.join(root, 'runs/v1'))).sort()) {
  runs.push([name, await readFile(path.join(root, 'runs/v1', name, 'run.json'), 'utf8')]);
}
await writeFile(output, JSON.stringify({ marker, runs }));
`;

async function fixture(t, { unborn = false } = {}) {
  const base = await realpath(await mkdtemp(path.join(tmpdir(), 'gallery hook 한글 ')));
  t.after(() => rm(base, { recursive: true, force: true }));
  const root = path.join(base, 'repository with spaces');
  const previousHooks = path.join(base, 'original hooks');
  await mkdir(root);
  await mkdir(previousHooks);
  const env = Object.fromEntries(Object.entries(process.env).filter(([key]) => !key.startsWith('GIT_')));
  Object.assign(env, {
    HOME: base,
    GIT_CONFIG_GLOBAL: '/dev/null',
    GIT_CONFIG_SYSTEM: '/dev/null',
    GIT_CONFIG_NOSYSTEM: '1',
    GIT_TERMINAL_PROMPT: '0',
    GIT_AUTHOR_NAME: 'Gallery Test',
    GIT_AUTHOR_EMAIL: 'gallery@example.invalid',
    GIT_COMMITTER_NAME: 'Gallery Test',
    GIT_COMMITTER_EMAIL: 'gallery@example.invalid',
  });
  function run(command, args, options = {}) {
    return spawnSync(command, args, { cwd: root, env, encoding: 'utf8', ...options });
  }
  function git(...args) {
    const result = run('git', args);
    assert.equal(result.status, 0, result.stderr || result.error?.message);
    return result.stdout.trimEnd();
  }
  async function save(relative, contents) {
    const target = path.join(root, relative);
    await mkdir(path.dirname(target), { recursive: true });
    await writeFile(target, contents);
  }
  git('init', '--template=', '--initial-branch=master');
  git('config', '--local', 'core.hooksPath', previousHooks);
  await save(generator, fixtureGenerator);
  await save(modulePath, "export const marker = 'staged module';\n");
  await save(source, 'initial source');
  await save(generated, 'original generated output');
  await save('unrelated.txt', 'initial unrelated');
  git('add', '--', '.');
  if (!unborn) {
    const tree = git('write-tree');
    const commit = git('commit-tree', tree, '-m', 'Fixture baseline');
    git('update-ref', 'refs/heads/master', commit);
  }
  function execute(options = {}) {
    return run(process.execPath, [driver], options);
  }
  function install(options = {}) {
    return run(process.execPath, [installer], options);
  }
  function entries() {
    return git('ls-files', '--stage', '-z').split('\0').filter(Boolean);
  }
  async function executable(file, contents) {
    await writeFile(file, contents);
    await chmod(file, 0o755);
  }
  function stagedData() {
    return JSON.parse(git('show', `:${generated}`));
  }
  return { base, root, previousHooks, env, run, git, save, execute, install, entries, executable, stagedData };
}

function succeeded(result) {
  assert.equal(result.status, 0, result.stderr || result.error?.message);
}

function withoutGenerated(entries) {
  return entries.filter(entry => !entry.endsWith(`\t${generated}`));
}

test('partial staging uses staged sources, generator and imported module without changing the worktree', async t => {
  const f = await fixture(t);
  await f.save(source, 'staged source');
  f.git('add', '--', source);
  await f.save(source, 'unstaged source');
  await f.save(generator, "throw new Error('unstaged generator must never execute');\n");
  await f.save(modulePath, "throw new Error('unstaged module must never execute');\n");
  const before = f.entries();
  succeeded(f.execute({ cwd: path.join(f.root, project, 'gallery') }));
  assert.deepEqual(f.stagedData(), { marker: 'staged module', runs: [['model 한글 space', 'staged source']] });
  assert.deepEqual(withoutGenerated(f.entries()), withoutGenerated(before));
  assert.equal(await readFile(path.join(f.root, source), 'utf8'), 'unstaged source');
  assert.equal(await readFile(path.join(f.root, generated), 'utf8'), 'original generated output');
  assert.match(await readFile(path.join(f.root, generator), 'utf8'), /unstaged generator/);
  assert.match(await readFile(path.join(f.root, modulePath), 'utf8'), /unstaged module/);
});

test('Git includes the generated snapshot in the same commit without committing unstaged edits', async t => {
  const f = await fixture(t);
  succeeded(f.install());
  await f.save(source, 'staged for commit');
  f.git('add', '--', source);
  await f.save(source, 'remaining unstaged');
  f.git('commit', '-m', 'Verify generated snapshot');
  const committed = JSON.parse(f.git('show', `HEAD:${generated}`));
  assert.deepEqual(committed.runs, [['model 한글 space', 'staged for commit']]);
  assert.equal(f.git('show', `HEAD:${source}`), 'staged for commit');
  assert.equal(await readFile(path.join(f.root, source), 'utf8'), 'remaining unstaged');
});

test('staged generator and module edits take effect before they are committed', async t => {
  const f = await fixture(t);
  await f.save(generator, fixtureGenerator.replace('{ marker, runs }', '{ marker, runs, stagedGenerator: true }'));
  await f.save(modulePath, "export const marker = 'new staged module';\n");
  f.git('add', '--', generator, modulePath);
  await f.save(generator, 'invalid unstaged javascript !');
  await f.save(modulePath, 'invalid unstaged javascript !');
  succeeded(f.execute());
  assert.equal(f.stagedData().stagedGenerator, true);
  assert.equal(f.stagedData().marker, 'new staged module');
});

test('staged deletions are excluded even when restored in the worktree and untracked files stay excluded', async t => {
  const f = await fixture(t);
  f.git('rm', '--cached', '--', source);
  await f.save(`${project}/runs/v1/kept/run.json`, 'kept staged source');
  f.git('add', '--', `${project}/runs/v1/kept/run.json`);
  await f.save(`${project}/runs/v1/untracked/run.json`, 'untracked source');
  const before = f.entries();
  succeeded(f.execute());
  assert.deepEqual(f.stagedData().runs, [['kept', 'kept staged source']]);
  assert.deepEqual(withoutGenerated(f.entries()), withoutGenerated(before));
  assert.equal(await readFile(path.join(f.root, source), 'utf8'), 'initial source');
});

test('a failed staged generator preserves the exact index and generated worktree bytes', async t => {
  const f = await fixture(t);
  await f.save(generator, "import { writeFile } from 'node:fs/promises';\nawait writeFile(process.argv[process.argv.indexOf('--output') + 1], 'partial output');\nthrow new Error('intentional generator failure');\n");
  f.git('add', '--', generator);
  const index = path.join(f.root, '.git/index');
  const before = await readFile(index);
  const result = f.execute();
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /intentional generator failure/);
  assert.deepEqual(await readFile(index), before);
  assert.equal(await readFile(path.join(f.root, generated), 'utf8'), 'original generated output');
  await assert.rejects(readFile(`${index}.lock`), { code: 'ENOENT' });
});

test('unrelated staged changes skip an invalid project generator', async t => {
  const f = await fixture(t);
  await f.save(generator, 'invalid javascript !');
  f.git('add', '--', generator);
  const tree = f.git('write-tree');
  const commit = f.git('commit-tree', tree, '-m', 'Invalid baseline');
  f.git('update-ref', 'refs/heads/master', commit);
  await f.save('unrelated.txt', 'changed unrelated');
  f.git('add', '--', 'unrelated.txt');
  const index = path.join(f.root, '.git/index');
  const before = await readFile(index);
  succeeded(f.execute());
  assert.deepEqual(await readFile(index), before);
});

test('the alternate index is authoritative and the default index is untouched', async t => {
  const f = await fixture(t);
  const defaultIndex = path.join(f.root, '.git/index');
  const alternate = path.join(f.base, 'alternate index 한글');
  await copyFile(defaultIndex, alternate);
  const defaultBefore = await readFile(defaultIndex);
  f.env.GIT_INDEX_FILE = alternate;
  await f.save(source, 'alternate staged source');
  f.git('add', '--', source);
  const before = f.entries();
  succeeded(f.execute());
  assert.equal(f.stagedData().runs[0][1], 'alternate staged source');
  assert.deepEqual(withoutGenerated(f.entries()), withoutGenerated(before));
  assert.deepEqual(await readFile(defaultIndex), defaultBefore);
});

test('a relative alternate index is resolved before entering a subdirectory snapshot', async t => {
  const f = await fixture(t);
  await copyFile(path.join(f.root, '.git/index'), path.join(f.root, 'alternate index'));
  f.env.GIT_INDEX_FILE = 'alternate index';
  await f.save(source, 'relative alternate staged source');
  f.git('add', '--', source);
  succeeded(f.execute());
  assert.equal(f.stagedData().runs[0][1], 'relative alternate staged source');
});

test('the initial commit can generate its staged gallery', async t => {
  const f = await fixture(t, { unborn: true });
  succeeded(f.execute());
  assert.equal(f.stagedData().marker, 'staged module');
});

test('deleting the entire project does not resurrect generated output', async t => {
  const f = await fixture(t);
  f.git('rm', '-r', '--cached', '--', project);
  const before = await readFile(path.join(f.root, '.git/index'));
  succeeded(f.execute());
  assert.deepEqual(await readFile(path.join(f.root, '.git/index')), before);
  assert.ok(!f.entries().some(entry => entry.includes(`\t${project}/`)));
});

test('a staged generator deletion fails instead of executing its worktree copy', async t => {
  const f = await fixture(t);
  f.git('rm', '--cached', '--', generator);
  const before = await readFile(path.join(f.root, '.git/index'));
  assert.notEqual(f.execute().status, 0);
  assert.deepEqual(await readFile(path.join(f.root, '.git/index')), before);
});

test('an existing index lock is respected without removing it', async t => {
  const f = await fixture(t, { unborn: true });
  const lock = path.join(f.root, '.git/index.lock');
  await writeFile(lock, 'another git process');
  assert.notEqual(f.execute().status, 0);
  assert.equal(await readFile(lock, 'utf8'), 'another git process');
});

test('installer chains the original pre-commit and preserves post-commit execution and arguments', async t => {
  const f = await fixture(t, { unborn: true });
  await f.executable(path.join(f.previousHooks, 'pre-commit'), '#!/bin/sh\nprintf "previous\\n" >> "$HOOK_LOG"\n');
  await f.executable(path.join(f.previousHooks, 'post-commit'), '#!/bin/sh\nexec .git/gotags "$@"\n');
  await f.executable(path.join(f.root, '.git/gotags'), '#!/bin/sh\nprintf "gotags:%s:%s\\n" "$PWD" "$1" >> "$HOOK_LOG"\n');
  f.env.HOOK_LOG = path.join(f.base, 'hooks.log');
  const original = await readFile(path.join(f.previousHooks, 'post-commit'));
  succeeded(f.install({ cwd: path.join(f.root, project) }));
  const hooks = f.git('config', '--local', '--get', 'core.hooksPath');
  assert.ok(hooks.startsWith(path.join(f.root, '.git') + path.sep));
  succeeded(f.run(path.join(hooks, 'pre-commit'), []));
  assert.equal(f.stagedData().marker, 'staged module');
  succeeded(f.run(path.join(hooks, 'post-commit'), ['argument with spaces']));
  assert.equal(await readFile(f.env.HOOK_LOG, 'utf8'), `previous\ngotags:${f.root}:argument with spaces\n`);
  assert.deepEqual(await readFile(path.join(f.previousHooks, 'post-commit')), original);
  await f.executable(path.join(f.previousHooks, 'post-commit'), '#!/bin/sh\nprintf "updated\\n" >> "$HOOK_LOG"\n');
  succeeded(f.run(path.join(hooks, 'post-commit'), []));
  assert.match(await readFile(f.env.HOOK_LOG, 'utf8'), /updated\n$/);
});

test('a failing previous pre-commit blocks generation and returns its failure', async t => {
  const f = await fixture(t, { unborn: true });
  await f.executable(path.join(f.previousHooks, 'pre-commit'), '#!/bin/sh\nexit 23\n');
  succeeded(f.install());
  const before = await readFile(path.join(f.root, '.git/index'));
  const hooks = f.git('config', '--get', 'core.hooksPath');
  assert.equal(f.run(path.join(hooks, 'pre-commit'), []).status, 23);
  assert.deepEqual(await readFile(path.join(f.root, '.git/index')), before);
});

test('reinstall preserves the original chain and forwards hooks added afterward', async t => {
  const f = await fixture(t, { unborn: true });
  f.env.HOOK_LOG = path.join(f.base, 'hooks.log');
  await f.executable(path.join(f.previousHooks, 'pre-commit'), '#!/bin/sh\nprintf "once\\n" >> "$HOOK_LOG"\n');
  succeeded(f.install());
  const hooks = f.git('config', '--get', 'core.hooksPath');
  succeeded(f.install());
  assert.equal(f.git('config', '--get', 'core.hooksPath'), hooks);
  succeeded(f.run(path.join(hooks, 'pre-commit'), []));
  assert.equal(await readFile(f.env.HOOK_LOG, 'utf8'), 'once\n');
  await f.executable(path.join(f.previousHooks, 'commit-msg'), '#!/bin/sh\nprintf "message:%s\\n" "$1" >> "$HOOK_LOG"\n');
  succeeded(f.run(path.join(hooks, 'commit-msg'), ['commit message file']));
  assert.match(await readFile(f.env.HOOK_LOG, 'utf8'), /message:commit message file\n$/);
});

test('installed driver remains usable when the entire project is deleted', async t => {
  const f = await fixture(t);
  succeeded(f.install());
  f.git('rm', '-r', '--', project);
  const hooks = f.git('config', '--get', 'core.hooksPath');
  const before = await readFile(path.join(f.root, '.git/index'));
  succeeded(f.run(path.join(hooks, 'pre-commit'), []));
  assert.deepEqual(await readFile(path.join(f.root, '.git/index')), before);
});

test('the real staged gallery generator can run with no worktree metadata access', async t => {
  const f = await fixture(t);
  await f.save(generator, await readFile(path.join(scripts, 'generate-gallery.mjs')));
  await f.save(modulePath, await readFile(fileURLToPath(new URL('../gallery/data.mjs', import.meta.url))));
  await f.save(source, JSON.stringify({ model: 'staged real model', status: 'completed' }));
  await f.save(`${project}/runs/v1/model 한글 space/output/index.html`, '<h1>Staged output</h1>');
  f.git('add', '--', project);
  await f.save(source, JSON.stringify({ model: 'unstaged real model' }));
  succeeded(f.execute());
  const javascript = f.git('show', `:${generated}`);
  assert.match(javascript, /staged real model/);
  assert.doesNotMatch(javascript, /unstaged real model/);
});

test('symlinks in the staged project fail without changing the index', async t => {
  const f = await fixture(t);
  const link = `${project}/runs/v1/linked`;
  await symlink(path.join(f.root, project, 'runs/v1/model 한글 space'), path.join(f.root, link));
  f.git('add', '--', link);
  const before = await readFile(path.join(f.root, '.git/index'));
  const result = f.execute();
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /symlinks or submodules/);
  assert.deepEqual(await readFile(path.join(f.root, '.git/index')), before);
});

test('relative previous hook paths retain their original working-directory semantics', async t => {
  const f = await fixture(t, { unborn: true });
  f.git('config', '--local', 'core.hooksPath', '.');
  f.env.HOOK_LOG = path.join(f.base, 'relative-hooks.log');
  await f.executable(path.join(f.root, 'pre-commit'), '#!/bin/sh\nprintf "relative\\n" >> "$HOOK_LOG"\n');
  succeeded(f.install());
  const hooks = f.git('config', '--get', 'core.hooksPath');
  succeeded(f.run(path.join(hooks, 'pre-commit'), []));
  assert.equal(await readFile(f.env.HOOK_LOG, 'utf8'), 'relative\n');
});

test('worktree-specific hook overrides fail installation without replacing the repo configuration', async t => {
  const f = await fixture(t);
  f.git('config', '--local', 'extensions.worktreeConfig', 'true');
  f.git('config', '--worktree', 'core.hooksPath', 'worktree-hooks');
  const result = f.install();
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /worktree-specific/);
  assert.equal(f.git('config', '--local', '--get', 'core.hooksPath'), f.previousHooks);
  assert.equal(f.git('config', '--worktree', '--get', 'core.hooksPath'), 'worktree-hooks');
});
