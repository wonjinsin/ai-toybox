# Repository Guidelines

## Project Structure & Module Organization

This is a Go command-line app for local audio and video transcription. `cmd/whisper-local` contains the process entry point. `internal/domain` holds pure transcription and subtitle policies; `internal/usecase` orchestrates workflows; `internal/port` defines boundaries; `internal/adapter` contains CLI, filesystem, FFmpeg, Whisper, and subtitle implementations; and `internal/bootstrap` wires them together. Architecture rules are tested in `internal/architecture`. Integration and package tests live alongside source files under `internal/`. See `docs/adr/` for architecture decisions. Runtime models belong in `models/`; large model files are not committed.

## Build, Test, and Development Commands

- `make build` builds `bin/whisper-local` from `cmd/whisper-local`.
- `make test` runs race-enabled Go tests and enforces at least 80% aggregate coverage.
- `make check` checks `gofmt` formatting and runs `go vet ./...`.
- `make clean` removes the generated `bin/` directory.

Install FFmpeg and `whisper.cpp` tools for end-to-end local transcription. Most package tests use in-memory ports and do not require model downloads.

## Coding Style & Naming Conventions

Use standard Go formatting (`gofmt`), tabs for indentation, and idiomatic Go names: exported identifiers use `MixedCaps`; unexported identifiers use `mixedCaps`. Keep package names short and lowercase. Preserve the hexagonal dependency direction: adapters may depend on ports and domain types, while use cases must not import adapters. Keep process arguments separate from shell strings and handle paths as values.

## Testing Guidelines

Place tests in `*_test.go` files next to the package they cover. Use descriptive `TestType_Behavior` names and table-driven tests for input variations. Cover boundary behavior, cancellation, path handling, checkpoint compatibility, and architecture rules where relevant. Run `make test` and `make check` before submitting changes.

## Commit & Pull Request Guidelines

Use Conventional Commit subjects, matching repository history, such as `fix: preserve subtitle symbols` or `refactor: clarify package boundaries`. Keep changes focused. Pull requests should explain behavior and motivation, list validation commands and results, link related issues when available, and include CLI examples or screenshots when output or user-facing behavior changes.

## Security & Configuration

Keep transcription media local and do not add credentials or model binaries to Git. Configure model locations with documented flags or `WHISPER_MODEL` and `WHISPER_VAD_MODEL` environment variables. Validate external inputs and preserve safe output handling, especially when overwriting existing files.
