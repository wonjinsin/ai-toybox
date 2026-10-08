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
artifact. Check the documented API and navigation capabilities, including `file://`;
record actual capability and verification gaps.
`Cannot navigate to a file URL without local file access.` is an access denial,
not unsupported protocol support. Stop that access and report the missing
permission; do not switch protocols, tools or settings to bypass it.

Install/register Aside outside benchmark attempts. Do not install it or add MCP
configuration during a measured run. If Aside is unavailable or cannot perform a
required check, perform the remaining permitted checks and record the reason and
verification gaps. Do not switch to another browser or install another browser
automation runtime. A security denial still stops that action, as specified below.

No new browser installation, raw CDP, insecure flags or security-setting changes.
Follow Aside's documented initialization and navigation capabilities.

## Check and record

1. Prefer the actual `output/index.html` over `file://`.
2. If direct-file navigation is unsupported but loopback HTTP is permitted, serve
   unchanged `output/` files temporarily on `127.0.0.1`. Stop the server before
   measurement ends, including on failure. HTTP success does not prove direct-file
   execution; also inspect local scripts/assets for file compatibility.
3. If no permitted method starts, perform static checks and name the actual
   startup/access failure in notes and the reply.

Use all Requirements and the Acceptance scenario in the [fixed task](../assets/solar-system-prompt.md)
as the verification checklist. Run the scenario in order, then cover remaining
requirements, including both specified viewport sizes; the scenario alone is not exhaustive.
Observe motion across frames and control actions; labels or code inspection do not prove behavior.
Check visible rendering, console/page errors, missing local assets, ES modules/imports
and local `fetch()`/XHR. Record the actual method, gaps and any unexecuted requirements/scenario steps.

An explicit security denial stops that action. Do not switch protocol, tools,
browser surface or settings to achieve it. Unavailable bindings and unsupported
protocols are capabilities, not permission to evade a denial.
