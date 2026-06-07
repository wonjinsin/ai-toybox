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
