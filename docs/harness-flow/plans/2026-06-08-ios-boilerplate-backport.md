# iOS Boilerplate Backport Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use harness-flow:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Backport the domain-neutral improvements from `smoke-tap` (i18n, project docs/rules, vector-icons, reanimated, widget guide, English docs) into `ios-boilerplate` without any smoke-tap-specific names or domain content.

**Architecture:** Additive backport. New generic files are added; existing files are touched only to wire scaffolds or translate docs. The boilerplate stays a managed Expo workflow — no `prebuild`, no native widget code.

**Tech Stack:** Expo SDK 55 · React Native 0.83 · React 19.2 · TypeScript strict · Expo Router · NativeWind v4 · expo-localization · react-native-reanimated · @expo/vector-icons

> **Note on TDD:** the boilerplate has no test runner configured. Verification uses `npx tsc --noEmit`, `grep`, and config inspection in place of unit tests. All commands run from `ios-boilerplate/` unless stated.

> **Path note:** repo root is the worktree root. The target subproject is `ios-boilerplate/`. All file paths below are relative to the repo root.

---

### Task 1: Dependencies & babel config

**Files:**

- Modify: `ios-boilerplate/package.json`
- Modify: `ios-boilerplate/babel.config.js`

`@expo/vector-icons` is already imported by `app/(tabs)/_layout.tsx` but missing from
`package.json` — this task also fixes that undeclared dependency.

- [ ] **Step 1: Add dependencies to `ios-boilerplate/package.json`**

Add these four entries to `dependencies` (keep alphabetical order matching the existing file):

```jsonc
"@expo/vector-icons": "^15.0.2",
"expo-localization": "~55.0.9",
"react-native-reanimated": "4.2.1",
"react-native-worklets": "0.7.2",
```

Resulting `dependencies` block:

```json
  "dependencies": {
    "@expo/vector-icons": "^15.0.2",
    "@react-native-async-storage/async-storage": "^2.2.0",
    "@tanstack/react-query": "^5.91.3",
    "clsx": "^2.1.1",
    "expo": "^55.0.8",
    "expo-constants": "~55.0.9",
    "expo-linking": "~55.0.8",
    "expo-localization": "~55.0.9",
    "expo-router": "~55.0.7",
    "expo-status-bar": "~55.0.4",
    "nativewind": "^4.2.3",
    "react": "19.2.0",
    "react-native": "^0.83.2",
    "react-native-reanimated": "4.2.1",
    "react-native-safe-area-context": "~5.6.2",
    "react-native-screens": "~4.23.0",
    "react-native-worklets": "0.7.2",
    "tailwind-merge": "^3.5.0",
    "zustand": "^5.0.12"
  },
```

- [ ] **Step 2: Add reanimated babel plugin to `ios-boilerplate/babel.config.js`**

Replace the file contents with:

```js
module.exports = function (api) {
  api.cache(true);
  return {
    presets: [
      ['babel-preset-expo', { jsxImportSource: 'nativewind' }],
    ],
    plugins: ['react-native-reanimated/plugin'],
  };
};
```

- [ ] **Step 3: Install and verify**

Run (from `ios-boilerplate/`): `npm install`
Expected: completes without peer-dependency errors; the four packages appear under `node_modules/`.

Verify declarations: `grep -E '"(@expo/vector-icons|expo-localization|react-native-reanimated|react-native-worklets)"' package.json`
Expected: all four lines printed.

- [ ] **Step 4: Commit**

```bash
git add ios-boilerplate/package.json ios-boilerplate/package-lock.json ios-boilerplate/babel.config.js
git commit -m "feat(boilerplate): add i18n, animation, and icon dependencies"
```

---

### Task 2: i18n scaffold

**Files:**

- Create: `ios-boilerplate/i18n/index.ts`
- Create: `ios-boilerplate/i18n/locales/en.json`
- Create: `ios-boilerplate/i18n/locales/ko.json`

- [ ] **Step 1: Create `ios-boilerplate/i18n/locales/en.json`**

```json
{
  "appName": "iOS Boilerplate",
  "tabs": {
    "home": "Home",
    "list": "List",
    "community": "Community",
    "history": "History",
    "profile": "Profile"
  },
  "common": {
    "greeting": "Hello, {{name}}",
    "loading": "Loading…",
    "empty": "Nothing here yet",
    "retry": "Retry"
  }
}
```

- [ ] **Step 2: Create `ios-boilerplate/i18n/locales/ko.json`**

Same key tree as `en.json`; only values differ.

```json
{
  "appName": "iOS 보일러플레이트",
  "tabs": {
    "home": "홈",
    "list": "목록",
    "community": "커뮤니티",
    "history": "이용내역",
    "profile": "내 정보"
  },
  "common": {
    "greeting": "안녕하세요, {{name}}",
    "loading": "불러오는 중…",
    "empty": "아직 항목이 없습니다",
    "retry": "다시 시도"
  }
}
```

- [ ] **Step 3: Create `ios-boilerplate/i18n/index.ts`**

`LocaleDict` is derived from `en.json`, so `ko.json` must keep the same shape.
Default fallback locale is `en`.

```ts
import { getLocales } from 'expo-localization';
import en from './locales/en.json';
import ko from './locales/ko.json';

type LocaleDict = typeof en;

const DEFAULT_LOCALE = 'en';
const locales: Record<string, LocaleDict> = { en, ko };

function getDict(): LocaleDict {
  const lang = getLocales()[0]?.languageCode ?? DEFAULT_LOCALE;
  return locales[lang] ?? locales[DEFAULT_LOCALE];
}

function getNestedValue(
  obj: Record<string, unknown>,
  keys: string[]
): string {
  let current: unknown = obj;
  for (const key of keys) {
    if (typeof current !== 'object' || current === null) return '';
    current = (current as Record<string, unknown>)[key];
  }
  return typeof current === 'string' ? current : '';
}

export function t(
  key: string,
  params?: Record<string, string | number>
): string {
  const dict = getDict();
  let str = getNestedValue(
    dict as unknown as Record<string, unknown>,
    key.split('.')
  );
  if (!str) return key;
  if (params) {
    Object.entries(params).forEach(([k, v]) => {
      str = str.replace(`{{${k}}}`, String(v));
    });
  }
  return str;
}
```

- [ ] **Step 4: Verify types**

Run: `npx tsc --noEmit`
Expected: PASS (no errors). Requires `resolveJsonModule` in `tsconfig.json` — confirm it is set; the project already imports JSON (`constants/mockData.ts`) so this should already be enabled. If `tsc` errors on the JSON import, add `"resolveJsonModule": true` to `compilerOptions` and re-run.

- [ ] **Step 5: Commit**

```bash
git add ios-boilerplate/i18n
git commit -m "feat(boilerplate): add i18n scaffold with en and ko locales"
```

---

### Task 3: Wire i18n into tab labels

**Files:**

- Modify: `ios-boilerplate/app/(tabs)/_layout.tsx`

Demonstrates real use of the i18n scaffold so it is not dead code. Only the
`tabBarLabel` strings change; icons and styling stay as-is.

- [ ] **Step 1: Add the import**

Add after the existing imports at the top of `app/(tabs)/_layout.tsx`:

```ts
import { t } from '../../i18n';
```

- [ ] **Step 2: Replace the five hardcoded labels**

| Screen | Before | After |
|---|---|---|
| `index` | `tabBarLabel: '홈',` | `tabBarLabel: t('tabs.home'),` |
| `list` | `tabBarLabel: '목록',` | `tabBarLabel: t('tabs.list'),` |
| `community` | `tabBarLabel: '커뮤니티',` | `tabBarLabel: t('tabs.community'),` |
| `history` | `tabBarLabel: '이용내역',` | `tabBarLabel: t('tabs.history'),` |
| `profile` | `tabBarLabel: '내 정보',` | `tabBarLabel: t('tabs.profile'),` |

- [ ] **Step 3: Verify**

Run: `npx tsc --noEmit`
Expected: PASS.

Run: `grep -n "tabBarLabel" "app/(tabs)/_layout.tsx"`
Expected: all five lines use `t('tabs.*')`; no Korean literals remain.

- [ ] **Step 4: Commit**

```bash
git add "ios-boilerplate/app/(tabs)/_layout.tsx"
git commit -m "feat(boilerplate): localize tab labels via i18n"
```

---

### Task 4: Project docs & rules

**Files:**

- Create: `ios-boilerplate/CLAUDE.md`
- Create: `ios-boilerplate/.claude/guidelines.md`
- Create: `ios-boilerplate/.github/CODEOWNERS`

- [ ] **Step 1: Create `ios-boilerplate/.claude/guidelines.md`**

Verbatim copy of the global guidelines (same file `smoke-tap` ships).

````markdown
# guidelines.md

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:

- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:

- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:

- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:

- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:

```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.
````

- [ ] **Step 2: Create `ios-boilerplate/CLAUDE.md`**

Generic, English, based on the boilerplate's actual structure. No widget/prebuild
content; links to the widget guide as optional reference.

````markdown
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
````

- [ ] **Step 3: Create `ios-boilerplate/.github/CODEOWNERS`**

Generic — no widget paths (the boilerplate has none).

```
# Default owner for everything in the repo.
*                                   @wonjinsin

# Context docs — keep them honest with the code.
CLAUDE.md                           @wonjinsin
README.md                           @wonjinsin
**/CLAUDE.md                        @wonjinsin
```

- [ ] **Step 4: Verify naming neutrality**

Run (from `ios-boilerplate/`): `grep -rin "smoke" CLAUDE.md .claude .github`
Expected: no matches.

- [ ] **Step 5: Commit**

```bash
git add ios-boilerplate/CLAUDE.md ios-boilerplate/.claude/guidelines.md ios-boilerplate/.github/CODEOWNERS
git commit -m "docs(boilerplate): add CLAUDE.md, guidelines, and CODEOWNERS"
```

---

### Task 5: Widget guide (docs only)

**Files:**

- Create: `ios-boilerplate/docs/guides/ios-widget.md`

Generic reference using placeholder names only. No code is added to the project.

- [ ] **Step 1: Create `ios-boilerplate/docs/guides/ios-widget.md`**

````markdown
# Adding a Native iOS Widget

This boilerplate ships **no** widget — it runs as a managed Expo app. This guide
describes a proven approach for adding a home-screen widget that records data with
one tap (App Intents, iOS 17+). All names below are placeholders; replace them
with your app's own.

> **Trade-off:** a native widget requires committing a native iOS project and
> switching from the managed workflow (`expo start`) to a `prebuild` workflow
> (`expo run:ios`). Adopt this only when you actually need a widget.

## Pieces involved

1. **`expo-widgets` config plugin** — declares the widget target in `app.json`:

   ```jsonc
   [
     "expo-widgets",
     {
       "enablePushNotifications": false,
       "widgets": [
         {
           "name": "ExampleWidget",
           "displayName": "Example",
           "description": "Your widget description",
           "supportedFamilies": ["systemSmall"]
         }
       ]
     }
   ]
   ```

2. **App Group** for sharing data between the app and the widget, e.g.
   `group.com.example.app`. Used by both the native shared-store code and the
   widget. Keep the ID identical everywhere it appears.

3. **A config plugin that writes Swift** (e.g. `plugins/withSharedStore.js`) so
   the native shared store is regenerated on every prebuild rather than hand-edited.

4. **Post-prebuild patch scripts**, run in a fixed order after
   `expo prebuild --clean`:
   1. overwrite the widget Swift,
   2. move the patch Run Script to fire **after** `[Expo] Configure project`,
   3. regenerate `ExpoModulesProvider.swift`.

   Wire them into a single script, e.g.:

   ```jsonc
   "prebuild:ios": "expo prebuild --platform ios --clean && node scripts/patch-widget.js && node scripts/fix-build-phase-order.js && node scripts/patch-expo-modules-provider.js"
   ```

## Gotchas

- Do not run `expo prebuild` alone — the build will use stale Swift or fail.
  Always run the full `prebuild:ios` chain.
- The App Group ID is duplicated across the plugin and the patch scripts. Keep
  them in sync. Only the `bundleIdentifier` lives in `app.json`.
- `expo-widgets` writes **absolute paths** into the Xcode project. If you prebuild
  inside a git worktree and later move/remove it, the build fails with
  "Build input files cannot be found" — re-run the prebuild chain from the
  canonical repo path.
- Do not edit files under `ios/` or `ios/Pods/` by hand — they are regenerated.

## Reference implementation

A complete working example of this pattern lives in the sibling `smoke-tap`
project (its `plugins/`, `scripts/`, `modules/`, and `ios/ExpoWidgetsTarget/`),
if available in your workspace.
````

- [ ] **Step 2: Verify naming neutrality of widget examples**

Run (from `ios-boilerplate/`): `grep -in "SmokeTap\|smoke-tap\|useTapStore" docs/guides/ios-widget.md`
Expected: no matches. (The single literal reference to the sibling `smoke-tap`
project in "Reference implementation" is intentional and is not a placeholder name —
it is fine for `grep -in "smoke"` to match only that line.)

- [ ] **Step 3: Commit**

```bash
git add ios-boilerplate/docs/guides/ios-widget.md
git commit -m "docs(boilerplate): add optional iOS widget guide"
```

---

### Task 6: Translate README to English

**Files:**

- Modify: `ios-boilerplate/README.md`

Full English rewrite of the existing Korean README, updated for the new i18n,
dependencies, and widget guide.

- [ ] **Step 1: Replace `ios-boilerplate/README.md` with:**

````markdown
# iOS Boilerplate

A React Native iOS boilerplate built from common UI patterns extracted from
finance / education / laundry app references.

**Tech Stack:** Expo · Expo Router · NativeWind v4 · Zustand · TanStack Query v5 ·
i18n (expo-localization) · react-native-reanimated · TypeScript strict

---

## Getting started

### 1. Install dependencies

```bash
npm install
```

### 2. Run the iOS simulator

```bash
npm run ios
# or
npx expo start --ios
```

### 3. Dev server only (QR code / Expo Go)

```bash
npm start
# or
npx expo start
```

---

## Commands

| Command | Description |
|--------|-------------|
| `npm run ios` | Run the iOS simulator |
| `npm run android` | Run the Android emulator |
| `npm start` | Start the Expo dev server |
| `npx tsc --noEmit` | Type-check |
| `npx expo start --clear` | Clear the Metro cache and run |

---

## Project structure

```
ios-boilerplate/
├── app/
│   ├── _layout.tsx              # QueryClientProvider + Stack root
│   └── (tabs)/
│       ├── _layout.tsx          # 5 tabs (labels via i18n)
│       ├── index.tsx            # Home tab
│       ├── list.tsx             # List tab
│       ├── community.tsx        # Community tab (FAB)
│       ├── history.tsx          # History tab
│       └── profile.tsx          # Profile tab
├── components/
│   ├── common/                  # Shared components (SkeletonBox, FAB, …)
│   └── home/                    # Home-only components
├── store/                       # Zustand stores
├── hooks/                       # TanStack Query hooks
├── services/mockApi.ts          # Mock API (real-server swap point)
├── constants/                   # Colors, typography, spacing, mock data
├── i18n/                        # t() helper + en/ko locales
└── types/                       # TypeScript types
```

---

## Internationalization (i18n)

User-facing strings live in `i18n/locales/en.json` and `i18n/locales/ko.json`,
which share the same key tree. Read them with `t()`:

```ts
import { t } from './i18n';

t('tabs.home');                     // "Home" / "홈"
t('common.greeting', { name: 'A' }); // "Hello, A" / "안녕하세요, A"
```

The active locale follows the device language and falls back to English.

---

## Connecting a real API

Swap the fetch functions in `services/mockApi.ts` for real API calls. The query
keys and return-type interfaces stay the same.

```ts
// before
export async function fetchUserProfile(): Promise<UserProfile> {
  await delay(rand(300, 600));
  return MOCK_USER;
}

// after
export async function fetchUserProfile(): Promise<UserProfile> {
  const res = await fetch('https://api.example.com/user');
  return res.json();
}
```

---

## Notes

- **NativeWind v4**: the `withNativeWind` wrap in `metro.config.js` and
  `import '../global.css'` in `app/_layout.tsx` are required.
- **TanStack Query v5**: there is no `onSuccess` — sync stores via `useEffect`.
- **Zustand persist**: banner dismiss state is persisted to AsyncStorage.
- **Shared segment tabs**: `list.tsx` and `community.tsx` share
  `useAppStore.activeSegmentTab`. For independent tab state, use a local
  `useState` like `history.tsx`.

---

## Adding a native iOS widget

This boilerplate ships no widget. See `docs/guides/ios-widget.md` for how to add
one (note: it requires switching to a `prebuild` workflow).
````

- [ ] **Step 2: Verify English + neutrality**

Run (from `ios-boilerplate/`): `grep -n "[가-힣]" README.md`
Expected: no matches (no Korean remaining).

Run: `grep -in "SmokeTap\|smoke-tap\|useTapStore" README.md`
Expected: no matches.

- [ ] **Step 3: Commit**

```bash
git add ios-boilerplate/README.md
git commit -m "docs(boilerplate): translate README to English and document i18n + widget guide"
```

---

### Task 7: Final verification

**Files:** none (verification only)

- [ ] **Step 1: Type check**

Run (from `ios-boilerplate/`): `npx tsc --noEmit`
Expected: PASS.

- [ ] **Step 2: Naming-neutrality sweep**

Run (from `ios-boilerplate/`):
`grep -rin "SmokeTap\|useTapStore\|흡연" . --exclude-dir=node_modules`
Expected: no matches.

`grep -rin "smoke-tap" . --exclude-dir=node_modules`
Expected: at most the single intentional reference line in `docs/guides/ios-widget.md`.

- [ ] **Step 3: Confirm config**

Run (from `ios-boilerplate/`):
`grep -q "react-native-reanimated/plugin" babel.config.js && echo OK-babel`
`grep -q "expo-localization" package.json && echo OK-deps`
Expected: both print their `OK-*` line.

- [ ] **Step 4: Report**

Summarize: deps added, i18n scaffold + tab wiring, docs/rules added, README in
English, widget guide added, all verifications green.

---

## Self-review

**Spec coverage:**

- i18n scaffold (ko+en, neutral keys) → Task 2, wired in Task 3. ✓
- Project docs/rules (CLAUDE.md generic, guidelines, CODEOWNERS generic) → Task 4. ✓
- `@expo/vector-icons` → Task 1. ✓
- reanimated (deps + babel only) → Task 1. ✓
- Widget guide docs-only → Task 5. ✓
- Docs language English (README translated, CLAUDE.md English-docs rule) → Tasks 2/4/6. ✓
- Non-goals respected: no screen changes, no native widget code, managed workflow kept. ✓
- Naming neutrality verified → Tasks 4/5/6/7. ✓

**Placeholder scan:** no TBD/TODO; every file step contains full content. ✓

**Type consistency:** `t(key, params?)` signature used identically in Tasks 2, 3, 6.
`en.json`/`ko.json` share one key tree; `LocaleDict = typeof en` enforces it. Tab
keys `tabs.{home,list,community,history,profile}` match between locale files and
the layout wiring. ✓
