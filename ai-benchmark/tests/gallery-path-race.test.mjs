import assert from 'node:assert/strict';
import { once } from 'node:events';
import fs from 'node:fs/promises';
import { syncBuiltinESMExports } from 'node:module';
import { tmpdir } from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { createGalleryServer } from '../scripts/serve-gallery.mjs';

for (const boundary of ['lstat', 'open']) {
  test(`replacing an ancestor before final ${boundary} cannot expose an outside file`, async t => {
    const root = await fs.mkdtemp(path.join(tmpdir(), 'gallery-path-race-'));
    const assets = path.join(root, 'runs/v1/safe/output/assets');
    const parked = path.join(root, 'parked');
    const outside = path.join(root, 'outside');
    await fs.mkdir(assets, { recursive: true });
    await fs.mkdir(outside);
    await fs.writeFile(path.join(assets, 'secret.txt'), 'Public output');
    await fs.writeFile(path.join(outside, 'secret.txt'), 'PRIVATE SENTINEL');
    const target = await fs.realpath(path.join(assets, 'secret.txt'));
    const original = fs[boundary];
    let swapped = false;
    const server = createGalleryServer({ root });
    t.after(async () => {
      fs[boundary] = original;
      syncBuiltinESMExports();
      server.closeAllConnections();
      await new Promise(resolve => server.close(resolve));
      await fs.rm(root, { recursive: true, force: true });
    });

    // Swap a real parent at a filesystem boundary; file reads and HTTP responses remain real.
    // Node isolates this test file in its own process, and the two cases run sequentially.
    fs[boundary] = async (file, ...args) => {
      if (file === target && !swapped) {
        swapped = true;
        await fs.rename(assets, parked);
        await fs.symlink(outside, assets);
      }
      return original(file, ...args);
    };
    syncBuiltinESMExports();
    server.listen(0, '127.0.0.1');
    await once(server, 'listening');
    const response = await fetch(`http://127.0.0.1:${server.address().port}/runs/v1/safe/output/assets/secret.txt`);
    const body = await response.text();
    assert.equal(swapped, true, 'the concurrent path replacement must occur');
    assert.equal(response.status, 403);
    assert.equal(body.includes('PRIVATE SENTINEL'), false, 'outside bytes must not reach the HTTP response');
  });
}
