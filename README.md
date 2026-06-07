# ai-toybox

A collection of experimental iOS projects built with React Native and Expo. Each folder is an independent, standalone project sharing a common stack: **Expo SDK 55 · Expo Router · NativeWind v4 · Zustand · TanStack Query v5 · TypeScript (strict)**.

> [!NOTE]
> These projects are not linked by a monorepo tool (no workspaces). Run `npm install` and start each project individually from its own folder.

## Projects

| Folder | Description |
|--------|-------------|
| [`ios-boilerplate/`](./ios-boilerplate) | A React Native iOS boilerplate based on common UI patterns extracted from finance, education, and laundry app references. A starting point for new apps. |
| [`smoke-tap/`](./smoke-tap) | A "one tap logs one smoke" app. Logs entries via an iOS 17+ interactive home screen widget (App Intents) without opening the app. |

### `ios-boilerplate/`

A boilerplate for quickly bootstrapping React Native iOS apps. Comes pre-wired with screen routing, state management (Zustand), data fetching (TanStack Query), and NativeWind styling. See [`ios-boilerplate/README.md`](./ios-boilerplate/README.md) for details.

### `smoke-tap/`

An app that logs smoking events with minimal friction. Record an entry with a single tap on the large in-app button, or the `+` button on the home screen widget — the widget increments the count via App Intents without launching the app. All data is stored on-device in `AsyncStorage` (no server, no accounts). See [`smoke-tap/README.md`](./smoke-tap/README.md) for details.
