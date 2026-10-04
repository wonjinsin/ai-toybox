---
name: benchmark-record
description: "Use when the user explicitly invokes benchmark-record to run the fixed Solar System benchmark."
disable-model-invocation: true
---

Current Claude Code session: `${CLAUDE_SESSION_ID}`. Carry the substituted value
into the shared workflow; an unexpanded placeholder is not an identifier.

Read and follow [the shared skill](../../../skills/benchmark-record/SKILL.md).
Resolve its references and assets relative to the shared skill's directory.
If the shared skill is missing, report the missing path and stop.
