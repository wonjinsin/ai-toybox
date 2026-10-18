#!/usr/bin/env node
import http from 'node:http';

const html = `<!doctype html>
<html lang="en"><meta charset="utf-8"><title>Review viewport</title>
<style>html,body{margin:0}iframe{display:block;border:0}</style>
<body><script>
const params = new URLSearchParams(location.search);
const sizes = new Map([['desktop', [1440, 900]], ['mobile', [390, 844]]]);
try {
  const url = new URL(params.get('url'));
  const size = sizes.get(params.get('size'));
  if (url.protocol !== 'http:' || url.hostname !== '127.0.0.1' || url.username || url.password || !size) {
    throw new Error('Use a loopback HTTP artifact URL and size=desktop or size=mobile.');
  }
  const frame = document.createElement('iframe');
  frame.id = 'review-frame';
  frame.title = 'Unchanged benchmark artifact';
  frame.width = size[0];
  frame.height = size[1];
  frame.src = url.href;
  document.body.append(frame);
} catch (error) {
  document.body.textContent = error.message;
}
</script></body></html>`;

const port = process.argv[2] ?? '0';
if (process.argv.length > 3 || !/^\d+$/.test(port) || Number(port) > 65535) {
  console.error('Usage: node serve-viewport.mjs [PORT]');
  process.exitCode = 1;
} else {
  const server = http.createServer((request, response) => {
    if (!/^127\.0\.0\.1(?::\d+)?$/.test(request.headers.host ?? '')) {
      response.writeHead(403).end();
    } else if (request.method !== 'GET' && request.method !== 'HEAD') {
      response.writeHead(405, { Allow: 'GET, HEAD' }).end();
    } else if (request.url.split('?')[0] !== '/') {
      response.writeHead(404).end();
    } else {
      response.writeHead(200, {
        'Content-Type': 'text/html; charset=utf-8',
        'Content-Length': Buffer.byteLength(html),
        'Cache-Control': 'no-store',
        'X-Content-Type-Options': 'nosniff',
        'Referrer-Policy': 'no-referrer',
      });
      response.end(request.method === 'HEAD' ? undefined : html);
    }
  });
  server.on('error', error => {
    console.error(error.message);
    process.exitCode = 1;
  });
  server.listen(Number(port), '127.0.0.1', () => {
    console.log('Viewport fixture: http://127.0.0.1:' + server.address().port + '/');
  });
}
