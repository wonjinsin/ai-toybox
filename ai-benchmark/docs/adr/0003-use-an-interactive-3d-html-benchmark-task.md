# ADR 0003: Use an Interactive 3D HTML Benchmark Task

Date: 2026-09-24

Status: Accepted

## Context

The user proposed an HTML artifact capable of rendering 3D content and selected an interactive 3D scene as the first benchmark direction. The accepted measurement priority is requirement-compliant implementation.

## Decision

Use an HTML artifact that renders an interactive 3D scene in a browser as the first benchmark task.

This decision establishes the task category. [ADR 0004](0004-use-a-small-animated-solar-system.md) defines the accepted scene theme and core interactions. [ADR 0005](0005-standardize-the-threejs-version.md) establishes a common Three.js version. Dependency delivery, artifact packaging, and the exact benchmark prompt remain undecided.

## Consequences

- Task design must define observable interaction outcomes so requirement compliance can be assessed.
- Any assessment of visual appeal must be distinguished from functional requirement compliance; evaluation criteria and weights require further agreement.
- Rendering dependencies, execution conditions, and evaluation evidence require agreement before benchmark execution. A browser used to view or evaluate an artifact is distinct from browser tool access granted to the model during generation.
