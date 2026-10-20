# Orrery / Orbital Observatory

Open `index.html` directly in a modern browser with WebGL enabled. All runtime assets are local; there is no build step, module import, network request, or application server dependency. Three.js r128 and its matching OrbitControls are bundled in `vendor/` under the MIT license.

Drag the scene to orbit and scroll to zoom. Select Cinder, Azure, or Ochre in the scene or planet list, then choose Track planet. Overview returns to the full system without clearing selection. Seeking pauses playback; playback stops at 120 simulation seconds.

Capture saves up to three immutable observations. Restore pauses and returns to standard camera framing. Export and import use the labeled JSON textarea; import replaces only the saved list. The stable planet IDs are `cinder`, `azure`, and `ochre`. All five snapshot fields are required: `id`, `time`, `speed`, `selectedPlanet`, and `viewMode`. The root is `{ "version": 1, "snapshots": [] }`. Fractional time and speed are supported. Unknown extra fields are ignored.

Snapshots are stored under `orrery-observations-v1` in this page's localStorage context. Browser restrictions can prevent persistence on local files; an explicit saving-unavailable notice indicates that in-memory changes may not survive reload. Reload always starts the live simulation at time 0, speed 1, overview, no selection, and automatic playback.

All DOM controls support ordinary Tab navigation and keyboard activation. Time and speed use arrow keys. The camera itself is pointer-operated; the planet list provides keyboard selection.

Verification during construction used the connected Aside browser over loopback HTTP at its available 1600×1000 viewport. Direct-file execution and persistence, and the exact 1440×900 and 390×844 viewports, were not browser-verified. The installed driver has no `setViewportSize` method. Runtime checks covered playback, deterministic seeking, selection, camera tracking/orbit/zoom, snapshot lifecycle, transactional JSON validation, reload persistence, corrupted-storage recovery, simulated write failure, and keyboard controls. See the parent `run.json` for recording limitations.

The playback clock keeps a monotonic watermark so an animation-frame timestamp older than a preceding DOM event cannot double-count elapsed time.
