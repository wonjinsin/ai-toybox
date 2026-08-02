# ai-toybox

A collection of experimental projects. Each folder is an independent, standalone project — see each project's own README for setup and details.

> [!NOTE]
> These projects are not linked by a monorepo tool (no workspaces). Set up and run each project individually from its own folder.

## Projects

| Folder | Description |
|--------|-------------|
| [`ios-boilerplate/`](./ios-boilerplate) | A boilerplate for quickly bootstrapping React Native iOS apps. Comes pre-wired with screen routing, state management (Zustand), data fetching (TanStack Query), and NativeWind styling. |
| [`ledger/`](./ledger) | An AI-powered personal ledger CLI in Go. Imports bank/card CSVs of any format (an AI figures out the column mapping), flags questionable rows interactively, auto-categorizes merchants, and answers natural-language questions about your spending via text-to-SQL — all through a local `claude -p`/`codex exec` subprocess, no API costs. |
| [`smoke-tap/`](./smoke-tap) | An app that logs smoking events with minimal friction. Record an entry with a single tap on the large in-app button, or the `+` button on the home screen widget — the widget increments the count via App Intents without launching the app. All data is stored on-device in AsyncStorage (no server, no accounts). |
