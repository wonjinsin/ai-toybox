import { mkdir, realpath, rename, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { collectRuns, collectionRevision } from '../gallery/data.mjs';

const sourceRoot = fileURLToPath(new URL('../', import.meta.url));

export async function generateGallery({ root = sourceRoot, output = path.join(root, 'gallery/data.js') } = {}) {
  const runs = await collectRuns(root);
  const snapshot = { revision: collectionRevision(runs), runs };
  const json = JSON.stringify(snapshot, null, 2)
    .replaceAll('<', '\\u003c').replaceAll('\u2028', '\\u2028').replaceAll('\u2029', '\\u2029');
  const content = `globalThis.BENCHMARK_GALLERY = ${json};\n`;
  await mkdir(path.dirname(output), { recursive: true });
  const temporary = `${output}.${process.pid}.tmp`;
  try {
    await writeFile(temporary, content, { flag: 'wx' });
    await rename(temporary, output);
  } finally {
    await rm(temporary, { force: true });
  }
  return snapshot;
}

async function main() {
  const options = {};
  const args = process.argv.slice(2);
  for (let index = 0; index < args.length; index += 2) {
    const flag = args[index];
    const value = args[index + 1];
    if (!['--root', '--output'].includes(flag) || !value || value.startsWith('--')) {
      throw new Error('Usage: node scripts/generate-gallery.mjs [--root DIRECTORY] [--output FILE]');
    }
    options[flag.slice(2)] = path.resolve(value);
  }
  const snapshot = await generateGallery(options);
  console.log(`Generated gallery/data.js with ${snapshot.runs.length} runs.`);
}

if (process.argv[1] && await realpath(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main().catch(error => {
    console.error(`Gallery generation failed: ${error.message}`);
    process.exitCode = 1;
  });
}
