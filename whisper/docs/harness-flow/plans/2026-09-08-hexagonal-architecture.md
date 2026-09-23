# Hexagonal Architecture Plan

Source: Inline approved design — separate pure subtitle policies, application use cases and ports, and CLI, FFmpeg, Whisper, and filesystem adapters. The user delegated all design and implementation decisions through completion.
Goal: Run the existing local transcription workflow through explicit inward dependencies, with application behavior testable without external processes or files.
Constraints: Preserve existing CLI behavior, subtitle quality rules, checkpoint filenames and JSON compatibility, publication safety, cancellation, batching, and the shared inference limit. Keep processing local. Write documentation and comments in English. Target at least 80% coverage where measurement is supported.

### Task 1: Preserve subtitle processing behind a pure domain boundary
Delivers: Existing subtitle timing, cleaning, correction, boundary reconciliation, and retry decisions remain identical while becoming independent of I/O.
Touches: internal/domain, internal/architecture, existing subtitle and retry tests.
Blocked by: none.
- [ ] Move cue, token, origin, speech interval, and chunk models with their policies and regression tests.
- [ ] Separate file reading and engine JSON schemas from domain logic.
- [ ] Enforce the domain dependency boundary with an architecture test.

### Task 2: Run transcription and resume through injected capabilities
Delivers: Single-file and directory workflows preserve current results and resume behavior using replaceable ports.
Touches: internal/application, internal/adapters/ffmpeg, internal/adapters/whisper, internal/adapters/filesystem, internal/adapters/local.
Blocked by: Task 1 contracts; independent adapter extraction can proceed concurrently.
- [ ] Define consumer-owned audio, detection, recognition, checkpoint, publication, and input discovery ports.
- [ ] Use a job session factory to bind per-file resources and cleanup outside the application; share the inference runner across sessions.
- [ ] Keep retry selection in domain and batching, stage transitions, and durable progress in application.
- [ ] Preserve raw checkpoint cues, correction reapplication, serialization version, output collision protection, and recovery paths.
- [ ] Test application success, failure, cancellation, resume, and batch behavior using in-memory ports.

### Task 3: Wire the command and enforce the completed architecture
Delivers: The existing command runs the complete refactored pipeline with unchanged flags, output, and exit codes.
Touches: cmd/whisper-local, internal/adapters/cli, integration tests, Makefile, README.md, docs/adr.
Blocked by: Task 2.
- [ ] Move flag parsing and progress presentation to CLI; compose real dependencies outside application.
- [ ] Migrate existing end-to-end fake-executable regressions and remove the former mixed app package.
- [ ] Add import checks that prohibit domain/application dependencies on adapters and operating-system I/O.
- [ ] Run formatting, race-enabled coverage, all tests, vet, and build; review the complete diff and fix blocking findings.
- [ ] Record architecture and validation commands in existing documentation locations.
