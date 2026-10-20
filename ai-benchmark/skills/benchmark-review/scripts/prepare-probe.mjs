#!/usr/bin/env node
import { readFile, readdir, lstat, realpath } from 'node:fs/promises';
import { createHash } from 'node:crypto';
import path from 'node:path';

const hash = value => createHash('sha256').update(value).digest('hex');
const options = {};
try {
  const args = process.argv.slice(2);
  for (let index = 0; index < args.length; index += 2) {
    const key = args[index], value = args[index + 1];
    if (!['--artifact', '--url', '--fixture', '--size', '--scenario', '--time', '--playing'].includes(key) || !value || Object.hasOwn(options, key)) {
      throw new Error('Usage: node prepare-probe.mjs --artifact OUTPUT --url LOOPBACK_URL --fixture LOOPBACK_URL --size mobile|desktop --scenario overview|persistence|invalid-storage|quota|diagnostics [--time SECONDS] [--playing true|false]');
    }
    options[key] = value;
  }
  for (const key of ['--artifact', '--url', '--fixture', '--size', '--scenario']) {
    if (!options[key]) throw new Error(`Missing ${key}`);
  }
  const artifactUrl = new URL(options['--url']);
  const fixtureUrl = new URL(options['--fixture']);
  for (const url of [artifactUrl, fixtureUrl]) {
    if (url.protocol !== 'http:' || url.hostname !== '127.0.0.1' || url.username || url.password) throw new Error('Only credential-free 127.0.0.1 HTTP URLs are permitted');
  }
  if (artifactUrl.origin === fixtureUrl.origin) throw new Error('The viewport fixture must have its own origin');
  if (!['mobile', 'desktop'].includes(options['--size'])) throw new Error('Invalid viewport size');
  if (!['overview', 'persistence', 'invalid-storage', 'quota', 'diagnostics'].includes(options['--scenario'])) throw new Error('Invalid scenario');
  const time = options['--time'] === undefined ? 5.2 : Number(options['--time']);
  if (!Number.isFinite(time) || time < 0 || time > 120) throw new Error('Time must be within 0..120');
  if (options['--playing'] !== undefined && !['true', 'false'].includes(options['--playing'])) throw new Error('Playing must be true or false');
  const root = await realpath(options['--artifact']);
  const files = [];
  async function visit(relative) {
    for (const entry of await readdir(path.join(root, relative), { withFileTypes: true })) {
      const name = path.posix.join(relative, entry.name);
      const filename = path.join(root, name);
      const info = await lstat(filename);
      if (info.isSymbolicLink()) throw new Error('Artifact symbolic links are not permitted');
      if (info.isDirectory()) await visit(name);
      else if (info.isFile()) files.push({ path: name, sha256: hash(await readFile(filename)) });
      else throw new Error('Artifact contains a non-regular file');
    }
  }
  await visit('');
  files.sort((a, b) => a.path.localeCompare(b.path, 'en'));
  if (!files.some(file => file.path === 'index.html')) throw new Error('Artifact has no index.html');
  for (const file of files) {
    const servedUrl = new URL(file.path.split('/').map(encodeURIComponent).join('/'), artifactUrl);
    const response = await fetch(servedUrl, { redirect: 'error', cache: 'no-store' });
    if (!response.ok || hash(Buffer.from(await response.arrayBuffer())) !== file.sha256) {
      throw new Error(`Served artifact differs from --artifact: ${file.path}`);
    }
  }
  const source = await readFile(new URL('./probe-observatory.js', import.meta.url), 'utf8');
  const profileSource = await readFile(new URL('../assets/probe-terra-xhigh.json', import.meta.url), 'utf8');
  const profile = JSON.parse(profileSource);
  fixtureUrl.searchParams.set('url', artifactUrl.href);
  fixtureUrl.searchParams.set('size', options['--size']);
  const provenance = { artifact_directory: root, artifact_sha256: hash(JSON.stringify(files)), files,
    runner_sha256: hash(source), profile_sha256: hash(profileSource) };
  const call = { scenario: options['--scenario'], time, playing: options['--playing'] === 'true', isolatedOrigin: true };
  const code = `var preparedProbe = ${source};
var preparedProfile = ${JSON.stringify(profile)};
var preparedTab = await openTab(${JSON.stringify(fixtureUrl.href)});
try {
  console.log((await snapshot(preparedTab, { interactive: true })).tree);
  var preparedFrame = (await preparedTab.frames()).find(frame => frame.url() === ${JSON.stringify(artifactUrl.href)});
  if (!preparedFrame) throw new Error('The expected artifact frame did not load');
  var verifyPreparedFiles = async () => preparedFrame.evaluate(async expected => {
    const actual = [];
    for (const file of expected) {
      const url = new URL(file.path.split('/').map(encodeURIComponent).join('/'), location.href);
      const response = await fetch(url, { redirect: 'error', cache: 'no-store' });
      if (!response.ok) throw new Error('Artifact response failed: ' + file.path);
      const digest = await crypto.subtle.digest('SHA-256', await response.arrayBuffer());
      const sha256 = Array.from(new Uint8Array(digest), byte => byte.toString(16).padStart(2, '0')).join('');
      if (sha256 !== file.sha256) throw new Error('Browser-served artifact mismatch: ' + file.path);
      actual.push({ path: file.path, sha256 });
    }
    return actual;
  }, ${JSON.stringify(files)});
  var preparedServedBefore = await verifyPreparedFiles();
  var preparedReport = await preparedProbe({frame:preparedFrame,tab:preparedTab,snapshot,profile:preparedProfile,...${JSON.stringify(call)}});
  preparedReport.provenance = ${JSON.stringify(provenance)};
  preparedReport.served_files = { url: preparedFrame.url(), before: preparedServedBefore, after: await verifyPreparedFiles() };
  if (preparedReport.scenario === 'overview') display(await preparedTab.screenshot());
  console.log('OBSERVATORY_PROBE_JSON=' + JSON.stringify(preparedReport));
} finally { await closeTab(preparedTab); }`;
  process.stdout.write(`${JSON.stringify({ title: `Observatory ${options['--scenario']} ${options['--size']}`, code })}\n`);
} catch (error) {
  console.error(`prepare-probe: ${error.message}`);
  process.exitCode = 1;
}
