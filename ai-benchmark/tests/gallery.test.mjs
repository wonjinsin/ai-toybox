import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { once } from 'node:events';
import { mkdtemp, mkdir, readFile, rm, symlink, utimes, writeFile } from 'node:fs/promises';
import http from 'node:http';
import { tmpdir } from 'node:os';
import path from 'node:path';
import test from 'node:test';
import vm from 'node:vm';
import { fileURLToPath } from 'node:url';
import { createGalleryServer } from '../scripts/serve-gallery.mjs';
import { collectRuns } from '../gallery/data.mjs';
import { generateGallery } from '../scripts/generate-gallery.mjs';

const generator = fileURLToPath(new URL('../scripts/generate-gallery.mjs', import.meta.url));

function generate(root) {
  return spawnSync(process.execPath, [generator, '--root', root], { encoding: 'utf8' });
}

test('generator produces a reusable classic-script snapshot and preserves it on failure', async t => {
  const { root, save } = await fixture(t);
  const prompt = '</script><script>throw new Error("injected")</script>\u2028한글';
  await save('runs/v1/alpha/run.json', { model: 'static' });
  await save('runs/v1/alpha/prompt.md', prompt);
  const result = generate(root);
  assert.equal(result.status, 0, result.stderr);
  const destination = path.join(root, 'gallery/data.js');
  const first = await readFile(destination, 'utf8');
  const context = vm.createContext({});
  vm.runInContext(first, context);
  assert.equal(context.BENCHMARK_GALLERY.runs[0].model, 'static');
  assert.equal(context.BENCHMARK_GALLERY.runs[0].prompt, prompt);
  assert.equal(generate(root).status, 0);
  assert.equal(await readFile(destination, 'utf8'), first);
  await rm(path.join(root, 'runs'), { recursive: true });
  await save('runs/v1', 'not a directory');
  assert.notEqual(generate(root).status, 0);
  assert.equal(await readFile(destination, 'utf8'), first);
});

test('the generator runs when its CLI entrypoint is reached through a symlink', async t => {
  const { root, save } = await fixture(t);
  await save('runs/v1/alpha/run.json', { model: 'symlink-entry' });
  const entrypoint = path.join(root, 'generate.mjs');
  await symlink(generator, entrypoint);
  const result = spawnSync(process.execPath, [entrypoint, '--root', root], { encoding: 'utf8' });
  assert.equal(result.status, 0, result.stderr);
  const context = vm.createContext({});
  vm.runInContext(await readFile(path.join(root, 'gallery/data.js'), 'utf8'), context);
  assert.equal(context.BENCHMARK_GALLERY.runs[0].model, 'symlink-entry');
});

test('static artifact links work below a Pages prefix and revisions depend only on content', async t => {
  const root = await mkdtemp(path.join(tmpdir(), 'gallery-static-'));
  t.after(() => rm(root, { recursive: true, force: true }));
  const id = 'model #1 % 한글';
  const output = path.join(root, 'runs/v1', id, 'output');
  await mkdir(output, { recursive: true });
  await writeFile(path.join(output, '../run.json'), '{"model":"static"}');
  const entrypoint = path.join(output, 'index.html');
  await writeFile(entrypoint, '<h1>Static preview</h1>');
  const [first] = await collectRuns(root);
  assert.equal(new URL(first.artifactUrl, 'https://example.org/ai-toybox/').pathname,
    `/ai-toybox/runs/v1/${encodeURIComponent(id)}/output/index.html`);
  await utimes(entrypoint, new Date(0), new Date(0));
  assert.equal((await collectRuns(root))[0].artifactRevision, first.artifactRevision);
  await writeFile(entrypoint, '<h1>Edited preview</h1>');
  assert.notEqual((await collectRuns(root))[0].artifactRevision, first.artifactRevision);
});

async function fixture(t) {
  const root = await mkdtemp(path.join(tmpdir(), 'benchmark-gallery-'));
  const server = createGalleryServer({ root });
  server.listen(0, '127.0.0.1');
  await once(server, 'listening');
  const origin = `http://127.0.0.1:${server.address().port}`;
  t.after(async () => {
    server.closeAllConnections();
    await new Promise(resolve => server.close(resolve));
    await rm(root, { recursive: true, force: true });
  });
  const runs = path.join(root, 'runs', 'v1');
  async function save(relative, value) {
    const target = path.join(root, relative);
    await mkdir(path.dirname(target), { recursive: true });
    await writeFile(target, typeof value === 'string' ? value : JSON.stringify(value));
  }
  async function snapshot() {
    return generateGallery({ root });
  }
  return { root, runs, origin, save, snapshot };
}

test('regeneration reflects run creation, updates, invalidation, repair and deletion', async t => {
  const { root, runs, save, snapshot } = await fixture(t);
  assert.deepEqual((await snapshot()).runs, []);
  await mkdir(path.join(runs, 'alpha'), { recursive: true });
  assert.deepEqual((await snapshot()).runs, []);
  await save('runs/v1/alpha/prompt.md', 'Incomplete but nonempty run');
  const incomplete = await snapshot();
  assert.deepEqual(incomplete.runs.map(run => run.id), ['alpha']);
  assert.equal(incomplete.runs[0].model, null);
  assert.equal(incomplete.runs[0].artifactUrl, null);
  assert.match(incomplete.runs[0].issues.join(' '), /run\.json/);

  await save('runs/v1/alpha/run.json', { model: 'model-a', effort: 'high', status: 'running', duration_ms: null });
  await save('runs/v1/alpha/prompt.md', 'Original prompt');
  await save('runs/v1/alpha/output/index.html', '<h1>Live artifact</h1>');
  await save('runs/v1/alpha/reviews/review.json', { status: 'partial', total_score: null });
  const first = await snapshot();
  assert.equal(first.runs[0].status, 'running');
  assert.equal(first.runs[0].review.total_score, null);
  assert.notEqual(first.revision, incomplete.revision);
  const unchanged = await snapshot();
  assert.equal(unchanged.revision, first.revision);

  await save('runs/v1/alpha/reviews/review.json', { status: 'complete', total_score: 0 });
  await save('runs/v1/alpha/run.json', { model: 'model-a', effort: 'high', status: 'completed', duration_ms: 3210 });
  const completed = await snapshot();
  assert.equal(completed.runs[0].review.total_score, 0);
  assert.equal(completed.runs[0].status, 'completed');
  assert.equal(completed.runs[0].durationMs, 3210);
  assert.notEqual(completed.revision, first.revision);
  assert.equal(completed.runs[0].artifactRevision, first.runs[0].artifactRevision);

  await save('runs/v1/beta/run.json', { model: 'model-b' });
  await save('runs/v1/alpha/run.json', '{');
  const malformed = await snapshot();
  assert.deepEqual(malformed.runs.map(run => run.id).sort(), ['alpha', 'beta']);
  const broken = malformed.runs.find(run => run.id === 'alpha');
  assert.equal(broken.model, null);
  assert.match(broken.issues.join(' '), /run\.json/);
  assert.ok(broken.artifactUrl);
  assert.equal(await readFile(path.join(runs, 'alpha', 'run.json'), 'utf8'), '{');

  await save('runs/v1/alpha/run.json', { model: 'model-repaired', status: 'completed' });
  assert.equal((await snapshot()).runs.find(run => run.id === 'alpha').model, 'model-repaired');
  await rm(path.join(runs, 'alpha'), { recursive: true });
  assert.deepEqual((await snapshot()).runs.map(run => run.id), ['beta']);
  assert.equal(await readFile(path.join(root, 'runs/v1/beta/run.json'), 'utf8'), JSON.stringify({ model: 'model-b' }));
});

test('artifact and supporting-file changes update revisions and serve fresh encoded URLs', async t => {
  const { origin, runs, save, snapshot } = await fixture(t);
  const id = 'model #1 % 한글';
  const base = `runs/v1/${id}`;
  await save(`${base}/run.json`, { model: 'model-a' });
  await save(`${base}/output/index.html`, '<link rel="stylesheet" href="assets/site.css"><h1>First</h1>');
  await save(`${base}/output/assets/site.css`, 'body { color: red; }');
  const initial = (await snapshot()).runs[0];
  const result = await fetch(new URL(initial.artifactUrl, origin));
  assert.equal(result.status, 200);
  assert.match(await result.text(), /First/);
  assert.match(result.headers.get('cache-control'), /no-store/);

  await save(`${base}/output/assets/site.css`, 'body { color: blue; }');
  const assetChanged = (await snapshot()).runs[0];
  assert.notEqual(assetChanged.artifactRevision, initial.artifactRevision);
  assert.notEqual(assetChanged.artifactUrl, initial.artifactUrl);
  const asset = await fetch(new URL('assets/site.css', new URL(assetChanged.artifactUrl, origin)));
  assert.equal(await asset.text(), 'body { color: blue; }');
  assert.equal(asset.headers.get('access-control-allow-origin'), '*');

  await save(`${base}/output/index.html`, '<h1>Second revision</h1>');
  const htmlChanged = (await snapshot()).runs[0];
  assert.notEqual(htmlChanged.artifactRevision, assetChanged.artifactRevision);
  assert.equal(await (await fetch(new URL(htmlChanged.artifactUrl, origin))).text(), '<h1>Second revision</h1>');
  await rm(path.join(runs, id, 'output/index.html'));
  assert.equal((await snapshot()).runs[0].artifactUrl, null);
  await save(`${base}/output/index.html`, '<h1>Restored</h1>');
  assert.equal(await (await fetch(new URL((await snapshot()).runs[0].artifactUrl, origin))).text(), '<h1>Restored</h1>');
});

test('static HTTP serving uses only the saved snapshot until regeneration', async t => {
  const { root, origin, save, snapshot } = await fixture(t);
  await save('runs/v1/alpha/run.json', { model: 'before' });
  await snapshot();
  const initial = await fetch(`${origin}/gallery/data.js`);
  assert.equal(initial.status, 200);
  const saved = await initial.text();
  await save('runs/v1/alpha/run.json', { model: 'after' });
  assert.equal(await (await fetch(`${origin}/gallery/data.js`)).text(), saved);
  assert.equal((await fetch(`${origin}/api/runs`)).status, 404);
  await snapshot();
  const context = vm.createContext({});
  vm.runInContext(await (await fetch(`${origin}/gallery/data.js`)).text(), context);
  assert.equal(context.BENCHMARK_GALLERY.runs[0].model, 'after');
  assert.equal(await readFile(path.join(root, 'runs/v1/alpha/run.json'), 'utf8'), '{"model":"after"}');
});

function rawRequest(origin, pathname, headers = {}, method = 'GET') {
  return new Promise((resolve, reject) => {
    const request = http.request(origin, { path: pathname, headers, method }, response => {
      const chunks = [];
      response.on('data', chunk => chunks.push(chunk));
      response.on('end', () => resolve({ status: response.statusCode, body: Buffer.concat(chunks).toString('utf8') }));
    });
    request.on('error', reject);
    request.end();
  });
}

test('HTTP routes cannot read repository secrets, traverse paths, or mutate files', async t => {
  const { root, origin, save } = await fixture(t);
  await save('private.txt', 'PRIVATE SENTINEL');
  await save('runs/v1/safe/run.json', { model: 'model-a' });
  await save('runs/v1/safe/output/index.html', '<h1>Safe</h1>');
  for (const pathname of [
    '/private.txt', '/runs/v1/safe/run.json', '/runs/v1/safe/output/',
    '/runs/v1/safe/output/../../../../private.txt',
    '/runs/v1/safe/output/%2e%2e/%2e%2e/%2e%2e/%2e%2e/private.txt',
    '/runs/v1/safe/output/%2e%2e%2f%2e%2e%2f%2e%2e%2f%2e%2e%2fprivate.txt',
  ]) {
    const response = await rawRequest(origin, pathname);
    assert.ok([400, 403, 404].includes(response.status), `${pathname}: ${response.status}`);
    assert.equal(response.body.includes('PRIVATE SENTINEL'), false);
  }
  assert.equal((await rawRequest(origin, '/api/runs', {}, 'POST')).status, 405);
  assert.equal((await rawRequest(origin, '/api/runs', { Host: 'attacker.example' })).status, 403);
  assert.equal((await rawRequest(origin, '/api/runs')).status, 404);
  assert.equal(await readFile(path.join(root, 'private.txt'), 'utf8'), 'PRIVATE SENTINEL');
});

test('symlinked metadata and output files cannot expose files outside a run', async t => {
  const { root, runs, origin, save, snapshot } = await fixture(t);
  await save('private/record.json', { model: 'PRIVATE MODEL' });
  await save('private/secret.html', '<h1>PRIVATE SENTINEL</h1>');
  await mkdir(path.join(runs, 'safe', 'output'), { recursive: true });
  await symlink(path.join(root, 'private/record.json'), path.join(runs, 'safe/run.json'));
  await symlink(path.join(root, 'private/secret.html'), path.join(runs, 'safe/output/index.html'));
  await symlink(path.join(root, 'private'), path.join(runs, 'escaped-run'));
  const data = await snapshot();
  assert.deepEqual(data.runs.map(run => run.id), ['safe']);
  assert.equal(data.runs[0].model, null);
  assert.equal(data.runs[0].artifactUrl, null);
  const response = await fetch(`${origin}/runs/v1/safe/output/index.html`);
  assert.ok([403, 404].includes(response.status));
  assert.equal((await response.text()).includes('PRIVATE SENTINEL'), false);
});

test('collection skips only literally empty directories and preserves incomplete or malformed contents', async t => {
  const { root, runs, save, snapshot } = await fixture(t);
  await mkdir(path.join(runs, 'empty'), { recursive: true });
  await mkdir(path.join(runs, 'nested', 'output'), { recursive: true });
  await save('runs/v1/hidden/.keep', '');
  await save('runs/v1/malformed/run.json', '{');
  await save('runs/v1/broken-review/reviews/review.json', '{');
  await save('runs/v1/unreviewed/run.json', { model: 'unreviewed' });
  const result = await snapshot();
  assert.deepEqual(result.runs.map(run => run.id).sort(), ['broken-review', 'hidden', 'malformed', 'nested', 'unreviewed']);
  assert.equal(result.runs.find(run => run.id === 'broken-review').review, null);
  assert.match(result.runs.find(run => run.id === 'broken-review').issues.join(' '), /review\.json/);
  assert.equal(result.runs.find(run => run.id === 'malformed').model, null);
  assert.equal(result.runs.find(run => run.id === 'nested').artifactUrl, null);
  assert.equal(await readFile(path.join(root, 'runs/v1/malformed/run.json'), 'utf8'), '{');
});

