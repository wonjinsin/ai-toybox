# Benchmark Skill Selection

During benchmark execution and recording, select and apply only skills explicitly
invoked in the current user request. `$benchmark-record` or `/benchmark-record` selects
`benchmark-record` alone unless the user also selects other skills. An explicit
request to read the shared skill and run it also selects `benchmark-record`;
quoted examples or discussion of a skill do not select it.

Do not apply unselected skills, even if their instructions were already injected
by the host or hooks. Keep this restriction after resume or compaction; neither
prior activation nor automatic routing, dependencies, or follow-up workflows
authorize another skill. Do not load or invoke it merely because it seems relevant.
Use ordinary host tools and pass this restriction to delegated benchmark work.
The discovery entrypoint, shared skill, and bundled resources are the same skill.

Follow actual instruction priority; injection alone grants no higher priority.
An unselected skill may apply only when a higher-priority host instruction
requires it. Record actually applied extras and their reasons in `run.json.notes`,
distinguishing mandatory injected instructions from newly required skill loads.
Disclose unavoidable exceptions in the final reply; do not claim isolation.
Do not announce suppressed skills as active or add boilerplate notes for them.

Do not invoke `benchmark-record` automatically when it is not explicitly selected.
Outside benchmark execution and recording, this policy does not restrict normal
host skill selection for repository investigation or maintenance.
