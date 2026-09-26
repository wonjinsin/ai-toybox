# ADR 0004: Use a Small Animated Solar System

Date: 2026-09-24

Status: Accepted

## Context

[ADR 0003](0003-use-an-interactive-3d-html-benchmark-task.md) establishes an interactive 3D HTML artifact as the first benchmark task category.

The user prefers a simple scene whose automatic motion makes the results intuitive to compare. The user accepted the proposed small solar system and its limited controls.

## Decision

Use a small animated 3D solar system with these core requirements:

- Display one sun and three planets.
- Start the planets' orbital motion automatically when the artifact opens.
- Give the planets different sizes, colors, and orbital speeds.
- Provide a pause/resume button for orbital motion.
- Provide a global speed control that changes orbital speeds while preserving their relative differences.

## Consequences

- The scene provides visible behavior immediately, without requiring initial interaction.
- The initial interaction scope is limited to pause/resume and global speed adjustment.
- Evaluation can examine observable requirements such as distinct planet appearances, ongoing orbital motion, paused motion, and consistent speed changes.
- This simple task may provide limited separation between stronger configurations. Actual benchmark results should inform any later expansion in difficulty.
- [ADR 0005](0005-standardize-the-threejs-version.md) establishes a common Three.js version. Dependency delivery, artifact packaging, precise motion constraints, and evaluation criteria require further agreement before the benchmark prompt is finalized.
