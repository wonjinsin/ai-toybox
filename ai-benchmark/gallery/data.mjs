import { constants } from 'node:fs';
import { lstat, open, readdir, realpath } from 'node:fs/promises';
import { createHash } from 'node:crypto';
import path from 'node:path';

function fileError(code, message) {
  return Object.assign(new Error(message), { code });
}

async function dataRoot(root) {
  const directory = await realpath(path.resolve(root));
  if (!(await lstat(directory)).isDirectory()) {
    throw fileError('ENOTDIR', 'The data root is not a directory');
  }
  return directory;
}

async function boundedPath(root, segments, directory = false) {
  let target = root;
  let info;
  for (let index = 0; index < segments.length; index += 1) {
    const segment = segments[index];
    if (!segment || segment === '.' || segment === '..' || /[/\\\0]/.test(segment)) {
      throw fileError('EACCES', 'Invalid path segment');
    }
    target = path.join(target, segment);
    info = await lstat(target, { bigint: true });
    if (info.isSymbolicLink()) throw fileError('ELOOP', 'Symbolic links are not served');
    if ((index < segments.length - 1 || directory) && !info.isDirectory()) {
      throw fileError('ENOTDIR', 'Expected a directory');
    }
  }
  return { target, info };
}

async function openRegularFile(root, segments) {
  const { target, info } = await boundedPath(root, segments);
  if (!info.isFile()) throw fileError('EACCES', 'Expected a regular file');
  const handle = await open(target, constants.O_RDONLY | constants.O_NOFOLLOW | constants.O_NONBLOCK);
  try {
    const opened = await handle.stat();
    if (!opened.isFile()) throw fileError('EACCES', 'Expected a regular file');
    return { handle, info: opened };
  } catch (error) {
    await handle.close();
    throw error;
  }
}

export async function openOutputFile(root, id, segments) {
  const directory = await dataRoot(root);
  return openRegularFile(directory, ['runs', 'v1', id, 'output', ...segments]);
}

const textOrNull = value => typeof value === 'string' && value.trim() ? value : null;
const numberOrNull = value => Number.isFinite(value) && value >= 0 ? value : null;
const strings = value => Array.isArray(value) ? value.filter(item => typeof item === 'string') : [];
const hash = value => createHash('sha256').update(JSON.stringify(value)).digest('hex');

export function collectionRevision(runs) {
  return hash(runs);
}

async function readOptional(root, segments, issues, required = false) {
  let handle;
  try {
    ({ handle } = await openRegularFile(root, segments));
    return await handle.readFile('utf8');
  } catch (error) {
    if (error.code !== 'ENOENT' || required) {
      issues.push(`${segments.join('/')}: ${error.code ?? error.message}`);
    }
    return null;
  } finally {
    await handle?.close();
  }
}

async function readRecord(root, segments, issues, required = false) {
  const text = await readOptional(root, segments, issues, required);
  if (text === null) return null;
  try {
    const value = JSON.parse(text);
    if (!value || typeof value !== 'object' || Array.isArray(value)) {
      throw new Error('Expected a JSON object');
    }
    return value;
  } catch (error) {
    issues.push(`${segments.join('/')}: ${error.message}`);
    return null;
  }
}

async function collectArtifact(root, prefix, issues) {
  const files = [];
  let hasEntrypoint = false;
  const output = [...prefix, 'output'];
  try {
    await boundedPath(root, output, true);
  } catch (error) {
    if (error.code !== 'ENOENT') issues.push(`${output.join('/')}: ${error.code ?? error.message}`);
    return { artifactRevision: null, artifactUrl: null };
  }

  async function visit(relative) {
    const segments = [...output, ...relative];
    let directory;
    try {
      ({ target: directory } = await boundedPath(root, segments, true));
      const entries = await readdir(directory, { withFileTypes: true });
      entries.sort((a, b) => a.name < b.name ? -1 : a.name > b.name ? 1 : 0);
      for (const entry of entries) {
        const child = [...relative, entry.name];
        const childSegments = [...output, ...child];
        try {
          const { info } = await boundedPath(root, childSegments);
          if (info.isDirectory()) {
            await visit(child);
          } else if (info.isFile()) {
            const name = child.join('/');
            files.push([name, String(info.size), String(info.mtimeNs), String(info.ctimeNs)]);
            if (name === 'index.html') hasEntrypoint = true;
          } else {
            issues.push(`${childSegments.join('/')}: not a regular file`);
          }
        } catch (error) {
          issues.push(`${childSegments.join('/')}: ${error.code ?? error.message}`);
        }
      }
    } catch (error) {
      issues.push(`${segments.join('/')}: ${error.code ?? error.message}`);
    }
  }

  await visit([]);
  files.sort((a, b) => a[0] < b[0] ? -1 : a[0] > b[0] ? 1 : 0);
  const artifactRevision = hash(files);
  const artifactUrl = hasEntrypoint
    ? `/runs/v1/${encodeURIComponent(prefix[2])}/output/index.html?v=${artifactRevision}`
    : null;
  return { artifactRevision, artifactUrl };
}

async function collectRun(root, id) {
  const issues = [];
  const prefix = ['runs', 'v1', id];
  const record = await readRecord(root, [...prefix, 'run.json'], issues, true);
  const review = await readRecord(root, [...prefix, 'reviews', 'review.json'], issues);
  const prompt = await readOptional(root, [...prefix, 'prompt.md'], issues);
  const reviewMarkdown = await readOptional(root, [...prefix, 'reviews', 'review.md'], issues);
  const artifact = await collectArtifact(root, prefix, issues);
  const date = id.match(/-(\d{4})(\d{2})(\d{2})-(\d{2})(\d{2})(\d{2})(?:-\d+)?$/);
  const createdAt = date ? `${date[1]}-${date[2]}-${date[3]}T${date[4]}:${date[5]}:${date[6]}+09:00` : null;
  const agent = id.startsWith('claude-code-') ? 'claude-code'
    : id.startsWith('codex-') ? 'codex'
    : id.startsWith('omp-') ? 'omp' : null;
  return {
    id,
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
    ...artifact,
    issues,
  };
}

export async function collectRuns(root) {
  let directory;
  let collection;
  try {
    directory = await dataRoot(root);
    ({ target: collection } = await boundedPath(directory, ['runs', 'v1'], true));
  } catch (error) {
    if (error.code === 'ENOENT') return [];
    throw error;
  }
  const entries = await readdir(collection, { withFileTypes: true });
  const runs = [];
  for (const entry of entries) {
    if (entry.isDirectory()) runs.push(await collectRun(directory, entry.name));
  }
  runs.sort((a, b) => (b.createdAt ?? '').localeCompare(a.createdAt ?? '') || a.id.localeCompare(b.id));
  return runs;
}
