# ai-toybox

A collection of independent experiments in AI benchmarking, audio tooling, and iOS apps. See each project's README for requirements, setup, and usage.

> [!NOTE]
> These projects are not linked by a monorepo tool (no workspaces). Set up and run each project individually from its own folder.

## Live gallery · GitHub Pages

**[Browse the AI Benchmark Gallery](https://wonjinsin.github.io/ai-toybox/)**

Explore saved runs of the Solar System / Three.js task, inspect model-generated HTML outputs, and compare review findings and quality scores. Quality scores are shown separately from requirement pass/fail results and unverified checks.

The site publishes the [`ai-benchmark/`](./ai-benchmark) gallery, not the entire project collection. No installation is required to browse it. GitHub Actions deploys it when changes to `ai-benchmark/` or the gallery workflow reach `master`, or when the workflow is run manually.

## Projects

| Folder | Description |
|--------|-------------|
| [`ai-benchmark/`](./ai-benchmark) | A fixed Solar System / Three.js task for OMP, Codex, and Claude Code, with saved run records, artifact reviews, and a static results gallery. |
| [`whisper/`](./whisper) | A local audio and video transcription CLI using `whisper.cpp` `large-v3` and Silero VAD. Generates text, SRT, or VTT subtitles without sending files to an external service. |
| [`ios-boilerplate/`](./ios-boilerplate) | A boilerplate for quickly bootstrapping React Native iOS apps. Comes pre-wired with screen routing, state management (Zustand), data fetching (TanStack Query), and NativeWind styling. |
| [`smoke-tap/`](./smoke-tap) | An app that logs smoking events with minimal friction. Record an entry with a single tap on the large in-app button, or the `+` button on the home screen widget — the widget increments the count via App Intents without launching the app. All data is stored on-device in AsyncStorage (no server, no accounts). |
| [`voice-transfer/`](./voice-transfer) | Japanese and English speech generation with Microsoft Edge TTS, followed by local RVC voice conversion. Source text is sent to Edge TTS; conversion runs locally. Natural Japanese intonation remains unverified by direct listening. |
