import { readFile, readdir, stat, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const sourceRoot = fileURLToPath(new URL('../', import.meta.url));
const root = path.resolve(process.argv[2] ?? sourceRoot);
const runsDir = path.join(root, 'runs', 'v1');

async function readOptional(file, issues, required = false) {
  try {
    return await readFile(file, 'utf8');
  } catch (error) {
    if (error.code !== 'ENOENT' || required) {
      issues.push(`${path.relative(root, file)}: ${error.code}`);
    }
    return null;
  }
}

async function readRecord(file, issues, required = false) {
  const text = await readOptional(file, issues, required);
  if (text === null) return null;
  try {
    const value = JSON.parse(text);
    if (!value || typeof value !== 'object' || Array.isArray(value)) {
      throw new Error('Expected a JSON object');
    }
    return value;
  } catch (error) {
    issues.push(`${path.relative(root, file)}: ${error.message}`);
    return null;
  }
}

const textOrNull = value => typeof value === 'string' && value.trim() ? value : null;
const numberOrNull = value => Number.isFinite(value) && value >= 0 ? value : null;
const strings = value => Array.isArray(value) ? value.filter(item => typeof item === 'string') : [];

async function collectRun(entry) {
  const directory = path.join(runsDir, entry.name);
  const issues = [];
  const record = await readRecord(path.join(directory, 'run.json'), issues, true);
  const review = await readRecord(path.join(directory, 'reviews', 'review.json'), issues);
  const prompt = await readOptional(path.join(directory, 'prompt.md'), issues);
  const reviewMarkdown = await readOptional(path.join(directory, 'reviews', 'review.md'), issues);
  const artifact = path.join(directory, 'output', 'index.html');
  let artifactUrl = null;
  try {
    if ((await stat(artifact)).isFile()) {
      artifactUrl = path.relative(root, artifact).split(path.sep).map(encodeURIComponent).join('/');
    }
  } catch (error) {
    if (error.code !== 'ENOENT') issues.push(`${entry.name}/output/index.html: ${error.code}`);
  }
  const date = entry.name.match(/-(\d{4})(\d{2})(\d{2})-(\d{2})(\d{2})(\d{2})(?:-\d+)?$/);
  const createdAt = date ? `${date[1]}-${date[2]}-${date[3]}T${date[4]}:${date[5]}:${date[6]}+09:00` : null;
  const agent = entry.name.startsWith('claude-code-') ? 'claude-code'
    : entry.name.startsWith('codex-') ? 'codex'
    : entry.name.startsWith('omp-') ? 'omp' : null;
  return {
    id: entry.name,
    model: textOrNull(record?.model),
    effort: textOrNull(record?.effort),
    agent,
    createdAt,
    status: textOrNull(record?.status) ?? 'unknown',
    durationMs: numberOrNull(record?.duration_ms),
    inputTokens: numberOrNull(record?.input_tokens),
    outputTokens: numberOrNull(record?.output_tokens),
    notes: strings(record?.notes),
    prompt,
    review,
    reviewMarkdown,
    artifactUrl,
    issues,
  };
}

const entries = await readdir(runsDir, { withFileTypes: true });
const runs = await Promise.all(entries.filter(entry => entry.isDirectory()).map(collectRun));
runs.sort((a, b) => (b.createdAt ?? '').localeCompare(a.createdAt ?? '') || a.id.localeCompare(b.id));
const data = JSON.stringify({ generatedAt: new Date().toISOString(), runs }).replaceAll('<', '\\u003c');
const [template, styles, app] = await Promise.all([
  readFile(path.join(sourceRoot, 'gallery', 'template.html'), 'utf8'),
  readFile(path.join(sourceRoot, 'gallery', 'styles.css'), 'utf8'),
  readFile(path.join(sourceRoot, 'gallery', 'app.js'), 'utf8'),
]);
const html = template.replace('/* GALLERY_STYLES */', () => styles)
  .replace('/* GALLERY_APP */', () => app)
  .replace('/* GALLERY_DATA */', () => data);
const output = path.join(root, 'index.html');
await writeFile(output, html);
console.log(`Gallery: ${output}\nRuns: ${runs.length}\nArtifacts: ${runs.filter(run => run.artifactUrl).length}`);
