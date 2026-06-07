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

t('tabs.home');                      // "Home" / "홈"
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
