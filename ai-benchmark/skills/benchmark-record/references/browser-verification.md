# Verify the Local HTML

Choose a permitted method during setup, before navigation, using ordinary host
tools. The saved artifact must run directly as `file://`, without a server/build.

## Select a browser

On shell-capable hosts with installed Chrome, prefer shell Node and Playwright
before connected browser tools. Node REPL import errors do not prove shell packages unavailable;
empty inventory or PATH lookup alone does not prove Chrome absent. On macOS, check
`/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`;
`channel: 'chrome'` also locates standard platform installations.
Check documented protocol support before choosing navigation.

If shell Playwright is missing/incompatible and installation is permitted, prepare
a temporary runtime outside `output/` during setup:

```sh
benchmark_runtime=$(mktemp -d)
npm install --prefix "$benchmark_runtime" --ignore-scripts --no-audit --no-fund playwright-core@1.63.0
```

Use shell Node `require()` on its absolute `node_modules/playwright-core` path,
not virtual Node REPL imports. Launch installed Chrome:

```js
const browser = await chromium.launch({ channel: 'chrome', headless: true, chromiumSandbox: true });
try {
  const page = await browser.newPage();
  await page.goto(pathToFileURL(absoluteIndexPath).href);
  // Check the actual scene and controls here.
} finally { await browser.close(); }
```

No new browser installation, existing user profile, raw CDP, insecure flags or
security-setting changes. Follow connected tools' documented initialization and
navigation capabilities.

## Check and record

1. Prefer the actual `output/index.html` over `file://`.
2. If direct-file navigation is unsupported but loopback HTTP is permitted, serve
   unchanged `output/` files temporarily on `127.0.0.1`. Stop the server before
   measurement ends, including on failure. HTTP success does not prove direct-file
   execution; also inspect local scripts/assets for file compatibility.
3. If no permitted method starts, perform static checks and name the actual
   startup/access failure in notes and the reply.

Check visible rendering, exactly three planets, automatic motion, pause/resume,
global speed changes preserving relative speeds, and console/page errors.
Observe motion across frames and control actions; labels or code inspection do not prove behavior.
Check missing local assets, ES modules/imports and local `fetch()`/XHR.
Record actual method and gaps.

An explicit security denial stops that action. Do not switch protocol, tools,
browser surface or settings to achieve it. Unavailable bindings and unsupported
protocols are capabilities, not permission to evade a denial.
