import http from 'node:http';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { pipeline } from 'node:stream/promises';
import { fileURLToPath } from 'node:url';
import { collectRuns, collectionRevision, openOutputFile } from '../gallery/data.mjs';

const sourceRoot = fileURLToPath(new URL('../', import.meta.url));
const appRoutes = new Map([
  ['/', ['index.html', 'text/html; charset=utf-8']],
  ['/index.html', ['index.html', 'text/html; charset=utf-8']],
  ['/gallery/app.js', ['gallery/app.js', 'text/javascript; charset=utf-8']],
  ['/gallery/styles.css', ['gallery/styles.css', 'text/css; charset=utf-8']],
]);
const types = new Map([
  ['.html', 'text/html; charset=utf-8'],
  ['.htm', 'text/html; charset=utf-8'],
  ['.xhtml', 'application/xhtml+xml; charset=utf-8'],
  ['.css', 'text/css; charset=utf-8'],
  ['.js', 'text/javascript; charset=utf-8'],
  ['.mjs', 'text/javascript; charset=utf-8'],
  ['.json', 'application/json; charset=utf-8'],
  ['.map', 'application/json; charset=utf-8'],
  ['.txt', 'text/plain; charset=utf-8'],
  ['.md', 'text/plain; charset=utf-8'],
  ['.svg', 'image/svg+xml'],
  ['.png', 'image/png'],
  ['.jpg', 'image/jpeg'],
  ['.jpeg', 'image/jpeg'],
  ['.gif', 'image/gif'],
  ['.webp', 'image/webp'],
  ['.avif', 'image/avif'],
  ['.ico', 'image/x-icon'],
  ['.woff', 'font/woff'],
  ['.woff2', 'font/woff2'],
  ['.ttf', 'font/ttf'],
  ['.otf', 'font/otf'],
  ['.wasm', 'application/wasm'],
  ['.mp3', 'audio/mpeg'],
  ['.wav', 'audio/wav'],
  ['.ogg', 'audio/ogg'],
  ['.mp4', 'video/mp4'],
  ['.webm', 'video/webm'],
]);

function send(request, response, status, body, type = 'text/plain; charset=utf-8') {
  const bytes = Buffer.from(body);
  response.writeHead(status, { 'Content-Type': type, 'Content-Length': bytes.length });
  response.end(request.method === 'HEAD' ? undefined : bytes);
}

function loopbackHost(host) {
  if (typeof host !== 'string' || !/^(?:127\.0\.0\.1|localhost|\[::1\])(?::\d{1,5})?$/i.test(host)) return false;
  try {
    new URL(`http://${host}`);
    return true;
  } catch {
    return false;
  }
}

function sameOrigin(request) {
  const origin = request.headers.origin;
  const site = request.headers['sec-fetch-site'];
  if (site && site !== 'same-origin' && site !== 'none') return false;
  return origin === undefined || origin === new URL(`http://${request.headers.host}`).origin;
}

function requestSegments(url) {
  if (!url.startsWith('/') || url.startsWith('//')) throw new Error('Invalid request path');
  const pathname = url.split('?')[0];
  if (pathname === '/') return [];
  return pathname.slice(1).split('/').map(value => {
    const segment = decodeURIComponent(value);
    if (!segment || segment === '.' || segment === '..' || /[/\\\u0000-\u001f\u007f]/.test(segment)) {
      throw new Error('Invalid request path');
    }
    return segment;
  });
}

export function createGalleryServer({ root = sourceRoot } = {}) {
  const dataRoot = path.resolve(root);
  return http.createServer(async (request, response) => {
    response.setHeader('Cache-Control', 'no-store');
    response.setHeader('X-Content-Type-Options', 'nosniff');
    response.setHeader('Referrer-Policy', 'no-referrer');
    if (!loopbackHost(request.headers.host)) {
      send(request, response, 403, 'Only loopback hosts are allowed.\n');
      return;
    }
    if (request.method !== 'GET' && request.method !== 'HEAD') {
      response.setHeader('Allow', 'GET, HEAD');
      send(request, response, 405, 'Only GET and HEAD are supported.\n');
      return;
    }

    let segments;
    try {
      segments = requestSegments(request.url ?? '');
    } catch {
      send(request, response, 400, 'Invalid request path.\n');
      return;
    }
    const route = `/${segments.join('/')}`;
    if (route === '/api/runs') {
      if (!sameOrigin(request)) {
        send(request, response, 403, 'The collection API is same-origin only.\n');
        return;
      }
      try {
        const runs = await collectRuns(dataRoot);
        send(request, response, 200, JSON.stringify({
          revision: collectionRevision(runs),
          checkedAt: new Date().toISOString(),
          runs,
        }), 'application/json; charset=utf-8');
      } catch (error) {
        send(request, response, 500, JSON.stringify({
          error: `Unable to read the run collection (${error.code ?? 'read error'}).`,
        }), 'application/json; charset=utf-8');
      }
      return;
    }

    const app = appRoutes.get(route);
    if (app) {
      try {
        const bytes = await readFile(path.join(sourceRoot, app[0]));
        send(request, response, 200, bytes, app[1]);
      } catch {
        send(request, response, 500, 'Unable to read the gallery application.\n');
      }
      return;
    }

    if (segments.length < 5 || segments[0] !== 'runs' || segments[1] !== 'v1' || segments[3] !== 'output') {
      send(request, response, 404, 'Not found.\n');
      return;
    }
    let handle;
    try {
      let info;
      ({ handle, info } = await openOutputFile(dataRoot, segments[2], segments.slice(4)));
      response.setHeader('Content-Type', types.get(path.extname(segments.at(-1)).toLowerCase()) ?? 'application/octet-stream');
      response.setHeader('Content-Length', String(info.size));
      response.setHeader('Access-Control-Allow-Origin', '*');
      response.setHeader('Cross-Origin-Resource-Policy', 'cross-origin');
      response.setHeader('Content-Security-Policy', 'sandbox allow-scripts allow-forms allow-modals allow-pointer-lock allow-popups allow-downloads');
      if (request.method === 'HEAD') {
        response.end();
      } else {
        await pipeline(handle.createReadStream({ autoClose: false }), response);
      }
    } catch (error) {
      if (!response.headersSent && !response.destroyed) {
        const status = error.code === 'ENOENT' || error.code === 'ENOTDIR' ? 404
          : error.code === 'EACCES' || error.code === 'EPERM' || error.code === 'ELOOP' ? 403 : 500;
        send(request, response, status, status === 500 ? 'Unable to read the output file.\n' : 'Not found.\n');
      } else if (!response.destroyed) {
        response.destroy();
      }
    } finally {
      await handle?.close();
    }
  });
}

function cliOptions(args) {
  let port = 4173;
  let root = sourceRoot;
  for (let index = 0; index < args.length; index += 1) {
    const flag = args[index];
    const value = args[index + 1];
    if ((flag !== '--port' && flag !== '--root') || !value || value.startsWith('--')) {
      throw new Error('Usage: node scripts/serve-gallery.mjs [--port PORT] [--root DIRECTORY]');
    }
    if (flag === '--port') {
      if (!/^\d+$/.test(value) || Number(value) > 65535) throw new Error('Port must be an integer from 0 to 65535.');
      port = Number(value);
    } else {
      root = path.resolve(value);
    }
    index += 1;
  }
  return { port, root };
}

function start() {
  try {
    const { port, root } = cliOptions(process.argv.slice(2));
    const server = createGalleryServer({ root });
    server.on('error', error => {
      const message = error.code === 'EADDRINUSE' ? `Port ${port} is already in use. Choose another with --port.`
        : `Unable to start the gallery (${error.code ?? error.message}).`;
      console.error(message);
      process.exitCode = 1;
    });
    server.listen(port, '127.0.0.1', () => {
      console.log(`Gallery: http://127.0.0.1:${server.address().port}/`);
    });
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) start();
