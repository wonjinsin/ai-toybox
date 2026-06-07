# CLAUDE.md — iOS Boilerplate

iOS-only Expo starter extracted from common app UI patterns: tabbed navigation,
mock API layer, global state, and server-cache hooks. Use it as the base for a
new iOS app.

## Stack

Expo SDK 55 · React Native 0.83 · React 19.2 · TypeScript strict · Expo Router
(file-based) · NativeWind v4 · Zustand v5 (persist + AsyncStorage) · TanStack
Query v5 · i18n (expo-localization) · react-native-reanimated.

Platform: iOS only (`app.json` → `platforms: ["ios"]`). Managed workflow — no
native project is committed.

## Commands

| Command | Purpose |
|---|---|
| `npm run ios` | Start Expo and open the iOS simulator |
| `npm start` | Expo dev server only |
| `npm run android` | Android emulator (best-effort; project targets iOS) |
| `npx tsc --noEmit` | Type check |
| `npx expo start --clear` | Clear Metro cache |

No test runner is configured.

## Architecture

```
app/         Expo Router screens (tabs: index · list · community · history · profile)
components/  UI (common · home)
store/       Zustand global state
hooks/       TanStack Query hooks
services/    mockApi.ts — swap these fetch functions for the real API
constants/   Design tokens + mock data
types/       Shared TypeScript types
i18n/        t() helper + en/ko locales
```

## Connecting a real API

Replace the fetch functions in `services/mockApi.ts` with real network calls.
Keep the query keys and return-type interfaces unchanged so the hooks keep working.

## Editing rules

- NativeWind v4 requires the `withNativeWind` wrap in `metro.config.js` and the
  `import '../global.css'` in `app/_layout.tsx` — keep both.
- TanStack Query v5 has no `onSuccess`; sync stores via `useEffect`.
- i18n: add user-facing strings to `i18n/locales/en.json` and `ko.json` (same key
  tree) and read them with `t('section.key')`.

## Behavioral guidelines — ALWAYS APPLY

**MANDATORY: Read `.claude/guidelines.md` and follow it on every task, no exceptions.**
These rules override default behavior and apply to all code changes in this repo —
trivial fixes included. Re-read at the start of each task; do not rely on memory.

@.claude/guidelines.md

## Documentation language

**All documentation in this repo (README.md, CLAUDE.md, guidelines, in-repo
markdown) must be written in English.** This overrides any global rule that
prefers another language for documentation. Code comments and commit messages
also remain in English. (User-facing UI strings under `i18n/` are unaffected.)

## Adding a native iOS widget (optional)

This boilerplate ships no widget. To add one (App Intents, iOS 17+), follow
`docs/guides/ios-widget.md`. Note: a native widget requires switching from the
managed workflow to a `prebuild` (`expo run:ios`) workflow.
