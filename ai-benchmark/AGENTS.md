# Benchmark Skill Selection

During benchmark execution and recording, select only skills explicitly invoked
in the current user request. `$benchmark-record` or `/benchmark-record` selects
`benchmark-record` alone unless the user also selects other skills. An explicit
request to read the shared skill and run it also selects `benchmark-record`;
quoted examples or discussion of a skill do not select it.

Continue applying instructions already injected and active through the host or
hooks, including after session resume or compaction. Their presence does not
authorize additional skills through automatic routing, dependencies, or follow-up
workflows. Do not load or invoke another skill merely because it seems relevant.
Use ordinary host tools and pass this restriction to delegated benchmark work.
The discovery entrypoint, shared skill, and bundled resources are the same skill.

Follow actual instruction priority: injection by a hook alone does not give an
instruction higher priority. Higher-priority host requirements still apply. In
`run.json.notes`, identify known active injected skill instructions separately
from additional required skills, with the reason for each required addition.
Disclose required additions in the final reply; do not claim an isolated run.

Do not invoke `benchmark-record` automatically when it is not explicitly selected.
Outside benchmark execution and recording, this policy does not restrict normal
host skill selection for repository investigation or maintenance.
