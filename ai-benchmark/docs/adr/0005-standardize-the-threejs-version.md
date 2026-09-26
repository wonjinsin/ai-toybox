# ADR 0005: Standardize the Three.js Version

Date: 2026-09-24

Status: Accepted

## Context

The first benchmark asks for a small interactive 3D solar system. Its primary measurement target is requirement-compliant implementation.

The user approved a common Three.js version so the compared configurations share the same rendering library.

## Decision

Use Three.js as the common 3D rendering library for the first benchmark. All runs within a comparison must use the same exact version.

Select and record the exact release before finalizing the benchmark prompt and executing runs. The specific release and delivery method have not yet been selected.

## Consequences

- Rendering library and version differences are controlled within each comparison.
- Familiarity with Three.js can still affect performance and must be considered when interpreting results.
- Three.js is an artifact runtime dependency. This decision does not grant models access to generation-time skills, tools, MCP services, or other agents.
- Artifact packaging and dependency delivery require further agreement.
- This choice applies to the first benchmark artifact, not the comparison UI or the overall project's technology stack.
