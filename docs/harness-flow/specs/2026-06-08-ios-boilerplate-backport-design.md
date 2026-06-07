# iOS Boilerplate Backport — Design

**Date:** 2026-06-08
**Status:** Approved (design), pending spec review
**Topic:** Backport generic improvements developed in `smoke-tap` into `ios-boilerplate`

---

## Background

`ios-boilerplate` is a React Native / Expo iOS starter extracted from common UI
patterns (5 generic tabs, mock API, Zustand + TanStack Query). `smoke-tap` was
built from this boilerplate and then evolved — gaining i18n, animation,
project-level conventions, and a native iOS widget.

The boilerplate has not received those generic improvements. This work backports
the **reusable, domain-neutral** parts of `smoke-tap` into the boilerplate.

Core dependency versions (Expo SDK 55, RN 0.83, React 19.2, NativeWind v4,
Zustand v5, TanStack Query v5) are **already identical** between the two projects.
This is therefore an **infra / convention backport**, not a version bump.

---

## Goals

Backport these six items, each converted to a **domain-neutral scaffold** (no
`smoke-tap` names, no smoking-domain content):

1. **i18n scaffold** — `expo-localization` + a small `t()` helper, with `ko` and
   `en` locales using generic example keys.
2. **Project docs / rules** — a generic root `CLAUDE.md`, a copy of the global
   `guidelines.md` under `.claude/`, and a generic `.github/CODEOWNERS`.
3. **`@expo/vector-icons`** — added as a dependency.
4. **Animation** — `react-native-reanimated` + `react-native-worklets` deps and
   the babel plugin wiring (no example component).
5. **Widget guide (docs only)** — a generic "how to add an iOS widget to this
   boilerplate" guide. **No widget code, plugins, scripts, or `ios/` directory.**
6. **Documentation language** — all in-repo docs standardized to **English**
   (the existing Korean README is translated), matching `smoke-tap`'s convention.

## Non-goals (explicitly excluded)

- Smoking-domain logic: `useTapStore`, `types/tap.ts`, count/elapsed components, etc.
- Any `SmokeTap*` naming.
- The paper-tone UI redesign (smoke-tap's specific design language).
- Native widget implementation: `expo-widgets`, `plugins/`, `scripts/`,
  `modules/`, committed `ios/`, and the `prebuild` workflow.
- Screen changes: the existing 5 tabs (`index`, `list`, `community`, `history`,
  `profile`) and `services/mockApi.ts` structure stay as-is. The boilerplate
  **remains a managed Expo workflow** (`expo start` / `expo start --ios`).

---

## Architecture / approach

**Additive backport.** New generic files are added; existing files are touched
only where needed to wire in the new scaffolds or to translate docs. No feature
removal, no workflow change.

### Dependencies & config

`ios-boilerplate/package.json` — add:

| Package | Version | Reason |
|---|---|---|
| `expo-localization` | `~55.0.9` | device locale for i18n |
| `@expo/vector-icons` | `^15.0.2` | common icon set |
| `react-native-reanimated` | `4.2.1` | animation |
| `react-native-worklets` | `0.7.2` | reanimated v4 peer |

`ios-boilerplate/babel.config.js` — add `plugins: ['react-native-reanimated/plugin']`.

`metro.config.js` and `app.json` — **unchanged** (managed workflow preserved).

### New files

| Path | Content |
|---|---|
| `ios-boilerplate/i18n/index.ts` | `t(key, params?)` helper (reused from smoke-tap), loads `ko` + `en`, falls back to device locale then `en` |
| `ios-boilerplate/i18n/locales/ko.json` | generic example keys (Korean values) |
| `ios-boilerplate/i18n/locales/en.json` | generic example keys (English values) |
| `ios-boilerplate/CLAUDE.md` | generic project guide based on the boilerplate's actual structure; English; includes the English-docs rule and a link to the widget guide |
| `ios-boilerplate/.claude/guidelines.md` | verbatim copy of the global `guidelines.md` (same as smoke-tap) |
| `ios-boilerplate/.github/CODEOWNERS` | generic: `* @wonjinsin` + context docs only (no widget paths) |
| `ios-boilerplate/docs/guides/ios-widget.md` | generic guide for adding an iOS widget, using placeholder names (`ExampleWidget`, `group.com.example.app`) |

### Touched existing files

| Path | Change |
|---|---|
| `ios-boilerplate/app/(tabs)/_layout.tsx` | wire tab labels to `t('tabs.*')` — a minimal, demonstrative i18n integration so the scaffold is not dead code |
| `ios-boilerplate/README.md` | translate to English; add i18n and widget-guide sections |
| `ios-boilerplate/package.json` | dependency additions (above) |
| `ios-boilerplate/babel.config.js` | reanimated plugin (above) |

### i18n locale shape (generic example keys)

Domain-neutral keys that demonstrate nesting and interpolation without any
app-specific content. Illustrative shape:

```json
{
  "appName": "iOS Boilerplate",
  "tabs": { "home": "...", "list": "...", "community": "...", "history": "...", "profile": "..." },
  "common": { "greeting": "Hello, {{name}}", "loading": "...", "empty": "...", "retry": "..." }
}
```

`ko.json` and `en.json` share the same key tree; only values differ.

### `t()` behavior

Reuse smoke-tap's implementation: read `getLocales()[0].languageCode`, look up
the matching dict, fall back to a default locale, dot-path key resolution, and
`{{param}}` interpolation. Return the key itself when missing. Default fallback
locale: `en`.

### Widget guide (docs only)

`docs/guides/ios-widget.md` explains, in English with placeholder names, the
approach proven in smoke-tap for adding a native iOS widget on top of this
boilerplate: the `expo-widgets` config plugin, an App Group for shared storage,
a config plugin that writes Swift, and the fixed-order post-`prebuild` patch
chain. It states clearly that adopting a widget means switching from the managed
workflow to a `prebuild` (`expo run:ios`) workflow. **No code is added to the
boilerplate** — this is reference material only.

---

## Naming-neutrality rules

- No `SmokeTap`, `smoke-tap`, `useTapStore`, `tap`, or smoking-domain strings
  anywhere in added content.
- Widget guide examples use placeholders: `ExampleWidget`,
  `group.com.example.app`, `com.example.app`.
- i18n keys are domain-agnostic.
- App name and bundle identifier stay as the existing `ios-boilerplate` /
  `com.example.iosboilerplate`.

---

## Verification

No automated test runner exists in the boilerplate. Verification criteria:

1. `npx tsc --noEmit` passes in `ios-boilerplate/` (i18n types, touched layout).
2. `grep -ri "smoke" ios-boilerplate/` (excluding `node_modules`) returns nothing
   in newly added/edited files — confirms naming neutrality.
3. All in-repo markdown under `ios-boilerplate/` is English.
4. `babel.config.js` includes the reanimated plugin; `package.json` includes the
   four new deps.
5. Tab labels render via `t()` (the i18n scaffold is actually used).
