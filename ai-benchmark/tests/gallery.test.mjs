import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtemp, mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';

const builder = fileURLToPath(new URL('../scripts/build-gallery.mjs', import.meta.url));

async function fixture(t) {
  const root = await mkdtemp(path.join(tmpdir(), 'benchmark-gallery-'));
  t.after(() => rm(root, { recursive: true, force: true }));
  const runs = path.join(root, 'runs', 'v1');
  await mkdir(runs, { recursive: true });
  return { root, runs };
}

async function generate(root) {
  const result = spawnSync(process.execPath, [builder, root], { encoding: 'utf8' });
  const html = await readFile(path.join(root, 'index.html'), 'utf8').catch(error => {
    if (error.code === 'ENOENT') return '';
    throw error;
  });
  const embedded = html.match(/<script id="gallery-data" type="application\/json">([\s\S]*?)<\/script>/);
  return { result, html, data: embedded ? JSON.parse(embedded[1]) : { runs: [] } };
}

test('an attempt without metadata remains visible instead of disappearing', async t => {
  const { root, runs } = await fixture(t);
  const id = 'codex-model-high-20261016-230821';
  await mkdir(path.join(runs, id));
  const { data, result } = await generate(root);
  assert.deepEqual(data.runs.map(run => run.id), [id]);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(data.runs[0].status, 'unknown');
  assert.equal(data.runs[0].model, null);
  assert.equal(data.runs[0].artifactUrl, null);
  assert.equal(data.runs[0].review, null);
  assert.match(data.runs[0].issues.join(' '), /run\.json/);
});

test('record text cannot escape the data block and execute as gallery markup', async t => {
  const { root, runs } = await fixture(t);
  const hostile = '</script><script>globalThis.compromised = true</script>';
  const directory = path.join(runs, 'hostile');
  await mkdir(directory);
  await writeFile(path.join(directory, 'run.json'), JSON.stringify({ model: hostile, notes: [hostile] }));
  await writeFile(path.join(directory, 'prompt.md'), hostile);
  const { data, html, result } = await generate(root);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(data.runs[0].model, hostile);
  assert.equal(data.runs[0].prompt, hostile);
  assert.equal(html.includes(hostile), false);
});

test('a malformed record does not hide its artifact or other runs', async t => {
  const { root, runs } = await fixture(t);
  const broken = path.join(runs, 'broken');
  await mkdir(path.join(broken, 'output'), { recursive: true });
  await writeFile(path.join(broken, 'run.json'), '{');
  await writeFile(path.join(broken, 'output', 'index.html'), '<h1>Original result</h1>');
  await mkdir(path.join(runs, 'healthy'));
  await writeFile(path.join(runs, 'healthy', 'run.json'), JSON.stringify({ model: 'model-b' }));
  const { data, result } = await generate(root);
  assert.equal(result.status, 0, result.stderr);
  assert.deepEqual(data.runs.map(run => run.id).sort(), ['broken', 'healthy']);
  const run = data.runs.find(run => run.id === 'broken');
  assert.equal(run.model, null);
  assert.equal(run.artifactUrl, 'runs/v1/broken/output/index.html');
  assert.match(run.issues.join(' '), /run\.json/);
  assert.equal(await readFile(path.join(broken, 'run.json'), 'utf8'), '{');
  assert.equal(await readFile(path.join(broken, 'output', 'index.html'), 'utf8'), '<h1>Original result</h1>');
});

test('artifact links preserve special characters and resolve from the generated entrypoint', async t => {
  const { root, runs } = await fixture(t);
  const id = 'model #1 % 한글';
  const directory = path.join(runs, id);
  await mkdir(path.join(directory, 'output'), { recursive: true });
  await writeFile(path.join(directory, 'run.json'), JSON.stringify({ model: 'model-a' }));
  await writeFile(path.join(directory, 'output', 'index.html'), '<h1>Result</h1>');
  const { data, result } = await generate(root);
  assert.equal(result.status, 0, result.stderr);
  const url = new URL(data.runs[0].artifactUrl, `file://${root}/index.html`);
  assert.equal(await readFile(fileURLToPath(url), 'utf8'), '<h1>Result</h1>');
});
