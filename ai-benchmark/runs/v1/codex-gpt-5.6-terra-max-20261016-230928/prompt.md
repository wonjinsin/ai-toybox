Create an HTML-based interactive 3D solar system observation tool using Three.js.

Requirements:

1. Display one sun and exactly three planets in a 3D scene.
2. Start the planets orbiting the sun automatically at simulation time 0 and global speed 1x when the scene opens.
3. Give the three planets different sizes, colors, and orbital speeds. Use circular orbits with periods of 12, 24, and 48 simulation seconds. These are benchmark values, not real astronomical periods. Keep each planet's initial orbital phase fixed so that a given simulation time always produces the same orbital positions, regardless of prior playback or seeking.
4. Provide a pause/resume button. Pausing must freeze simulation time and all planet orbital positions. Resuming must continue from the current simulation time without a position jump. Planet selection and camera controls must remain usable while paused.
5. Provide a global speed control ranging from 0.25x to 4x that changes the rate of simulation time while preserving the planets' relative orbital speeds. Changing speed must not change the current simulation time or orbital positions. A speed change while paused must take effect when playback resumes.
6. Make `output/index.html` run when opened directly in a browser as a `file://` page, without starting a server or running a build step to view it. Use `file://`-compatible scripts and assets; do not rely on local ES module scripts/imports or `fetch()` or XHR for local files.
7. Allow users to select any planet either by clicking it in the 3D scene or by choosing it from a labeled planet list. Both methods must update the same selection, visibly highlight the selected planet, and update an information panel showing its name, orbital period in simulation seconds, and current orbital angle in degrees. The angle must agree with the rendered orbital position and update during playback and seeking.
8. Provide a tracking mode for the selected planet and a control to return to the full-system overview. While tracking, keep the selected planet near the center of the view as it moves. Selecting another planet while tracking must update the highlight, information panel, and camera target together. Returning to the overview must preserve the selected planet, simulation time, global speed, and playback state.
9. Provide a time slider for seeking within 0 to 120 simulation seconds and a visible current-time readout. Seeking must pause playback and immediately update all planet positions, the selected planet's information, and the tracking camera when active. Playback must stop at 120 seconds; it must not wrap or continue beyond the range. Seeking to an earlier time must allow playback to resume from that time.
10. Allow camera orbit and zoom around the current target in both overview and tracking modes, including while paused. A camera drag must not be interpreted as a planet-selection click. Interacting with the planet list, time slider, speed control, or buttons must not trigger scene selection or camera movement.
11. At viewport sizes of 1440x900 and 390x844 CSS pixels, keep the scene and all primary controls usable without horizontal scrolling. Information panels must not cover or block the controls needed for selection, tracking, seeking, speed changes, or pause/resume.

Acceptance scenario:

1. Open the scene and confirm that all three planets orbit automatically.
2. Select a planet in the scene and enable tracking. Confirm that its highlight, live information, and camera target agree.
3. Pause, then seek to simulation time 30 seconds. Confirm that all planets move to their positions for that time and that the selected planet's information and tracking camera update while playback stays paused.
4. Set the global speed to 2x while paused. Confirm that time and orbital positions remain unchanged.
5. Resume. Confirm that motion continues from the sought positions without a jump and simulation time advances at twice the real-time rate.
6. Select a different planet from the list while tracking. Confirm that the selection, information panel, and camera target switch together without changing time, speed, or playback state.
7. Return to the overview. Confirm that the full system is visible and selection, time, speed, and playback state are preserved.
8. Pause and seek to 30 seconds again. Confirm that all orbital positions match the earlier 30-second state, regardless of the intervening speed and selection changes.

The intended experience is an observation tool with clear controls and consistent state across selection, camera tracking, time seeking, and playback. Keep the visual design and implementation approach open while satisfying the specified behavior.
