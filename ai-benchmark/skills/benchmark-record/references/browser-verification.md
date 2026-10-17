# Verify the Local HTML

Choose a permitted Aside navigation method during setup, before navigation.
The saved artifact must run directly as `file://`, without a server/build.

## Select a browser

Use only connected Aside MCP when its direct `repl` tool is available. Use the
current model to drive the installed Aside browser through that tool; never use
Aside `exec` or a natural-language CLI task, which launches another model session.
Read the tool's current documentation before navigation. Create a new tab for
the benchmark and close only tabs created for this attempt. Do not attach to
unrelated user tabs, read browsing memory, or change account/profile settings.
Use the documented snapshot, locator, screenshot and event APIs to check the
artifact. Record actual capability and verification gaps.

Install/register Aside outside benchmark attempts. Do not install it or add MCP
configuration during a measured run. Unavailable Aside or inability to perform a
required check triggers the same stop rule, not remaining checks or a static-only
completion. Do not switch to another browser or install another browser runtime.

No new browser installation, raw CDP, insecure flags or security-setting changes.
Follow Aside's documented initialization and navigation capabilities.

## Check and record

1. Use the actual `output/index.html` over `file://` only when local-file navigation
   is explicitly supported in the connected session. A generic `openTab(url)` API
   or session-scoped `fs` does not establish host-file access. If support is
   unavailable or unestablished, select documented, host-permitted loopback HTTP
   before navigation; do not probe `file://` just to rediscover the limitation.
2. If a supported file navigation fails because the browser lacks local-file
   capability, including `Cannot navigate to a file URL without local file access.`,
   record the exact operation/error and use the same HTTP path only when host
   file access and loopback HTTP are permitted. Do not infer an OS permission
   failure from that message alone. An explicit host/tool security policy denial
   is not this capability fallback: take the
   [whole-attempt stop path](../SKILL.md#stop-on-access-or-verification-blockers),
   without retrying the denied operation, changing permissions or bypassing it.
3. For HTTP, run a tracked host process serving only unchanged `output/`, not the
   repository root:

   ```sh
   python3 -m http.server 8765 --bind 127.0.0.1 --directory "<run>/output"
   ```

   Replace `<run>/output` with the actual output directory (quote paths containing
   spaces); use `--directory output` when already in the run directory. Wait for
   startup, then use Aside's direct REPL:
   `openTab("http://127.0.0.1:8765/index.html")`. If the port is occupied, choose
   another available port and use it in both places; do not stop an unrelated
   process. Retain the process handle/PID for cleanup. Never use Aside `exec` to
   start the server.
4. Check rendering, required interactions and console/page/resource errors over
   the selected transport. Record HTTP results separately from direct-file
   verification: HTTP success leaves direct-file execution unverified. Inspect
   local scripts/assets for file compatibility without claiming execution proof.
   Keep the artifact directly runnable without a server/build.
5. Close only attempt-created tabs and stop the temporary server on every exit,
   including failure, before measurement ends. If no permitted method is
   available, take the whole-attempt stop path; do not substitute static checks.

Use all Requirements and the Acceptance scenario in the [fixed task](../assets/solar-system-prompt.md)
as the verification checklist. Run the scenario in order, then cover remaining
requirements, including both specified viewport sizes; the scenario alone is not exhaustive.
Observe motion across frames and control actions; labels or code inspection do not prove behavior.
Check visible rendering, console/page errors, missing local assets, ES modules/imports
and local `fetch()`/XHR. Record the actual method, gaps and any unexecuted requirements/scenario steps.

Access and verification blockers stop the entire attempt, not only navigation.
An observed artifact defect remains a benchmark finding, not an infrastructure
blocker.
