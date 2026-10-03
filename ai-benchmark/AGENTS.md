# Benchmark Skill Selection

Use only skills explicitly invoked in the current user request. In particular,
`$benchmark-record` or `/benchmark-record` selects `benchmark-record` alone.
Do not load or invoke other skills because they seem relevant to implementation,
verification, or recording. Use ordinary host tools for those tasks.

If no skill is explicitly invoked, do not invoke `benchmark-record` or another
skill automatically. Higher-priority host instructions still apply; disclose any
skill they require beyond the user's selection instead of claiming an isolated
run.
