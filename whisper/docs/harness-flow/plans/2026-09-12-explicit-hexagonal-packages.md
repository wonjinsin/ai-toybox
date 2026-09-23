# Explicit Hexagonal Packages Plan

Source: Inline approved design — use independent domain, usecase, port/in, port/out, adapter/in, adapter/out, and bootstrap packages. Input adapters consume input ports; use cases implement input ports and consume output ports; output adapters implement output ports. Bootstrap owns concrete wiring.
Goal: Make architectural roles visible and enforceable as Go package boundaries while preserving the existing local transcription workflow.
Constraints: Preserve existing CLI behavior, subtitle quality rules, checkpoint filenames and JSON compatibility, publication safety, cancellation, batching, and the shared inference limit. Keep processing local. Write documentation and comments in English. Target at least 80% coverage where measurement is supported. Preserve user-owned changes and avoid unrelated cleanup.

### Task 1: Run the existing workflow through independent port contracts
Delivers: CLI and engine/storage adapters depend on contracts without importing the use-case implementation package.
Touches: internal/port, internal/application, internal/adapters, internal/integration, cmd/whisper-local, internal/architecture.
Blocked by: none.
- [ ] Input ports own transcription interfaces, requests, results, progress responses, and request validation.
- [ ] Output ports own session configuration, resource capabilities, recognition and checkpoint contracts, and checkpoint invariants.
- [ ] Shared output-collision errors live in the neutral port package; input and output port packages do not import each other.
- [ ] Use cases explicitly translate input requests to output session configuration; preserve all configuration values.
- [ ] Preserve existing tests and behavior; add a failing dependency test before removing adapters' implementation dependency.

### Task 2: Execute the command through the final package layout
Delivers: The existing command runs with usecase, directional adapter, and bootstrap packages, and tests reject boundary violations and legacy package reintroduction.
Touches: internal/usecase, internal/adapter, internal/bootstrap, internal/architecture, internal/integration, cmd/whisper-local, README.md, docs/adr.
Blocked by: Task 1.
- [ ] Move application orchestration to usecase, CLI to adapter/in/cli, and engine/storage/render/process implementations to adapter/out.
- [ ] Bootstrap assembles the use cases and adapters and owns per-job session wiring; main delegates execution to bootstrap.
- [ ] Remove internal/application, internal/adapters, and the former internal/app without compatibility aliases.
- [ ] Enforce the complete import direction, including adapter independence from usecase and bootstrap.
- [ ] Document the final ownership and read order in the existing architecture documentation.
- [ ] Pass race-enabled coverage, formatting, vet, build, CLI smoke checks, and independent review.
