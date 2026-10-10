(() => {
  "use strict";

  const TAU = Math.PI * 2;
  const MAX_TIME = 120;
  const planetSpecs = [
    { id: "aurelia", name: "Aurelia", period: 12, radius: 3.5, size: 0.43, phase: 0.45, color: 0xff8868, emissive: 0x552013 },
    { id: "vesper", name: "Vesper", period: 24, radius: 5.8, size: 0.66, phase: 2.05, color: 0x79d5f1, emissive: 0x123849 },
    { id: "noctis", name: "Noctis", period: 48, radius: 8.2, size: 0.55, phase: 4.05, color: 0xc296ff, emissive: 0x32184c }
  ];

  const element = (id) => document.getElementById(id);
  const ui = {
    scene: element("scene"),
    sceneError: element("scene-error"),
    sceneErrorCopy: element("scene-error-copy"),
    selectedName: element("selected-name"),
    selectedPeriod: element("selected-period"),
    selectedAngle: element("selected-angle"),
    timeReadout: element("time-readout"),
    timeSlider: element("time-slider"),
    speedSlider: element("speed-slider"),
    speedValue: element("speed-value"),
    pauseButton: element("pause-button"),
    playLabel: element("play-label"),
    playState: element("play-state"),
    statusPill: element("status-pill"),
    trackButton: element("track-button"),
    overviewButton: element("overview-button"),
    planetChoices: Array.from(document.querySelectorAll(".planet-choice"))
  };

  const state = {
    time: 0,
    speed: 1,
    playing: true,
    selectedId: planetSpecs[0].id,
    cameraMode: "overview"
  };

  let THREE = window.THREE;
  let scene;
  let camera;
  let renderer;
  let clock;
  let raycaster;
  let cameraTarget;
  let cameraOffset;
  let cameraSpherical;
  let cameraTheta = 0.88;
  let cameraPhi = 1.08;
  let cameraDistance = 19;
  let drag = null;
  const pointerNdc = { x: 0, y: 0 };
  const planets = new Map();
  const planetMeshes = [];

  function showSceneError(message) {
    ui.sceneErrorCopy.textContent = message;
    ui.sceneError.hidden = false;
  }

  if (!THREE) {
    showSceneError("Three.js did not load. Open this file with an internet connection so its classic script can load.");
    return;
  }

  try {
    scene = new THREE.Scene();
    scene.background = new THREE.Color(0x050812);
    camera = new THREE.PerspectiveCamera(43, window.innerWidth / window.innerHeight, 0.1, 150);
    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: false, powerPreference: "high-performance" });
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 1.75));
    renderer.setSize(window.innerWidth, window.innerHeight);
    renderer.setClearColor(0x050812, 1);
    renderer.outputEncoding = THREE.sRGBEncoding;
    renderer.domElement.setAttribute("aria-hidden", "true");
    renderer.domElement.setAttribute("tabindex", "-1");
    ui.scene.appendChild(renderer.domElement);
    cameraTarget = new THREE.Vector3(0, 0, 0);
    cameraOffset = new THREE.Vector3();
    cameraSpherical = new THREE.Spherical();
    clock = new THREE.Clock();
    raycaster = new THREE.Raycaster();
    createLighting();
    createStarfield();
    createSun();
    createPlanets();
    bindControls();
    updatePlanetPositions(state.time);
    syncCameraTarget();
    updateCamera();
    updateInterface();
    window.addEventListener("resize", onResize, { passive: true });
    window.requestAnimationFrame(animate);
  } catch (error) {
    showSceneError("The 3D scene could not start in this browser. Try a current browser with WebGL enabled.");
    if (window.console && typeof window.console.error === "function") window.console.error(error);
  }

  function createLighting() {
    scene.add(new THREE.AmbientLight(0x9aaaca, 0.48));
    const sunlight = new THREE.PointLight(0xffd29b, 2.8, 100, 1.35);
    sunlight.position.set(0, 0, 0);
    scene.add(sunlight);
    const coolFill = new THREE.DirectionalLight(0x8fb8ff, 0.28);
    coolFill.position.set(-9, 6, 4);
    scene.add(coolFill);
  }

  function createStarfield() {
    let seed = 73129;
    const random = () => {
      seed = (seed * 16807) % 2147483647;
      return (seed - 1) / 2147483646;
    };
    const count = 1050;
    const positions = new Float32Array(count * 3);
    for (let i = 0; i < count; i += 1) {
      const vertical = random() * 2 - 1;
      const azimuth = random() * TAU;
      const distance = 35 + random() * 48;
      const radial = Math.sqrt(1 - vertical * vertical) * distance;
      const offset = i * 3;
      positions[offset] = Math.cos(azimuth) * radial;
      positions[offset + 1] = vertical * distance;
      positions[offset + 2] = Math.sin(azimuth) * radial;
    }
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute("position", new THREE.BufferAttribute(positions, 3));
    const material = new THREE.PointsMaterial({ color: 0xd9e8ff, size: 0.075, sizeAttenuation: true, transparent: true, opacity: 0.74, depthWrite: false });
    scene.add(new THREE.Points(geometry, material));
  }

  function createSun() {
    const core = new THREE.Mesh(
      new THREE.SphereGeometry(1.12, 48, 40),
      new THREE.MeshStandardMaterial({ color: 0xffcc74, emissive: 0xf18a31, emissiveIntensity: 1.4, roughness: 0.72, metalness: 0 })
    );
    core.name = "sun";
    scene.add(core);
    const corona = new THREE.Mesh(
      new THREE.SphereGeometry(1.48, 36, 30),
      new THREE.MeshBasicMaterial({ color: 0xffa34d, transparent: true, opacity: 0.105, side: THREE.BackSide, depthWrite: false })
    );
    corona.name = "sun-corona";
    scene.add(corona);
  }

  function createPlanets() {
    for (const spec of planetSpecs) {
      const orbitPoints = [];
      const segments = 160;
      for (let i = 0; i < segments; i += 1) {
        const angle = (i / segments) * TAU;
        orbitPoints.push(new THREE.Vector3(Math.cos(angle) * spec.radius, 0, Math.sin(angle) * spec.radius));
      }
      const orbit = new THREE.LineLoop(
        new THREE.BufferGeometry().setFromPoints(orbitPoints),
        new THREE.LineBasicMaterial({ color: spec.color, transparent: true, opacity: 0.3, depthWrite: false })
      );
      orbit.name = `${spec.id}-orbit`;
      scene.add(orbit);

      const material = new THREE.MeshStandardMaterial({
        color: spec.color,
        emissive: spec.emissive,
        emissiveIntensity: 0.26,
        roughness: 0.42,
        metalness: 0.12
      });
      const mesh = new THREE.Mesh(new THREE.SphereGeometry(spec.size, 36, 32), material);
      mesh.name = spec.id;
      scene.add(mesh);

      const halo = new THREE.Mesh(
        new THREE.TorusGeometry(spec.size * 1.52, Math.max(0.025, spec.size * 0.055), 8, 64),
        new THREE.MeshBasicMaterial({ color: spec.color, transparent: true, opacity: 0.9, depthWrite: false })
      );
      halo.name = `${spec.id}-selection`;
      halo.rotation.x = Math.PI / 2;
      halo.visible = spec.id === state.selectedId;
      mesh.add(halo);

      const body = { ...spec, mesh, halo, baseEmissiveIntensity: material.emissiveIntensity };
      if (halo.visible) material.emissiveIntensity = 0.72;
      planets.set(spec.id, body);
      planetMeshes.push(mesh);
    }
  }

  function bindControls() {
    for (const button of ui.planetChoices) {
      button.addEventListener("click", () => selectPlanet(button.getAttribute("data-planet")));
    }
    ui.trackButton.addEventListener("click", () => {
      state.cameraMode = "tracking";
      syncCameraTarget();
      updateCamera();
      updateInterface();
    });
    ui.overviewButton.addEventListener("click", () => {
      state.cameraMode = "overview";
      cameraDistance = Math.max(cameraDistance, 17.5);
      syncCameraTarget();
      updateCamera();
      updateInterface();
    });
    ui.pauseButton.addEventListener("click", () => {
      if (state.playing) {
        state.playing = false;
      } else if (state.time < MAX_TIME) {
        state.playing = true;
      }
      updateInterface();
    });
    ui.timeSlider.addEventListener("input", () => {
      state.time = clamp(Number(ui.timeSlider.value) || 0, 0, MAX_TIME);
      state.playing = false;
      updatePlanetPositions(state.time);
      syncCameraTarget();
      updateCamera();
      updateInterface();
    });
    ui.speedSlider.addEventListener("input", () => {
      state.speed = clamp(Number(ui.speedSlider.value) || 1, 0.25, 4);
      updateInterface();
    });

    const canvas = renderer.domElement;
    canvas.addEventListener("pointerdown", onPointerDown);
    canvas.addEventListener("pointermove", onPointerMove);
    canvas.addEventListener("pointerup", onPointerUp);
    canvas.addEventListener("pointercancel", onPointerCancel);
    canvas.addEventListener("wheel", onWheel, { passive: false });
    canvas.addEventListener("contextmenu", (event) => event.preventDefault());
  }

  function onPointerDown(event) {
    if (event.button !== 0) return;
    drag = { pointerId: event.pointerId, x: event.clientX, y: event.clientY, lastX: event.clientX, lastY: event.clientY, moved: false };
    if (typeof renderer.domElement.setPointerCapture === "function") renderer.domElement.setPointerCapture(event.pointerId);
  }

  function onPointerMove(event) {
    if (!drag || drag.pointerId !== event.pointerId) return;
    const deltaX = event.clientX - drag.lastX;
    const deltaY = event.clientY - drag.lastY;
    if (Math.abs(event.clientX - drag.x) + Math.abs(event.clientY - drag.y) > 5) drag.moved = true;
    if (drag.moved) {
      cameraTheta -= deltaX * 0.006;
      cameraPhi = clamp(cameraPhi + deltaY * 0.006, 0.22, Math.PI - 0.22);
      updateCamera();
    }
    drag.lastX = event.clientX;
    drag.lastY = event.clientY;
  }

  function onPointerUp(event) {
    if (!drag || drag.pointerId !== event.pointerId) return;
    const wasClick = !drag.moved && Math.abs(event.clientX - drag.x) + Math.abs(event.clientY - drag.y) <= 5;
    drag = null;
    if (wasClick) pickPlanet(event.clientX, event.clientY);
  }

  function onPointerCancel(event) {
    if (drag && drag.pointerId === event.pointerId) drag = null;
  }

  function onWheel(event) {
    event.preventDefault();
    cameraDistance = clamp(cameraDistance * Math.exp(event.deltaY * 0.0011), 8.5, 34);
    updateCamera();
  }

  function pickPlanet(clientX, clientY) {
    const rect = renderer.domElement.getBoundingClientRect();
    if (!rect.width || !rect.height) return;
    pointerNdc.x = ((clientX - rect.left) / rect.width) * 2 - 1;
    pointerNdc.y = -((clientY - rect.top) / rect.height) * 2 + 1;
    raycaster.setFromCamera(pointerNdc, camera);
    const hits = raycaster.intersectObjects(planetMeshes, false);
    if (hits.length) selectPlanet(hits[0].object.name);
  }

  function selectPlanet(id) {
    if (!planets.has(id)) return;
    state.selectedId = id;
    for (const [planetId, planet] of planets) {
      const selected = planetId === id;
      planet.halo.visible = selected;
      planet.mesh.material.emissiveIntensity = selected ? 0.72 : planet.baseEmissiveIntensity;
    }
    updatePlanetPositions(state.time);
    syncCameraTarget();
    updateCamera();
    updateInterface();
  }

  function normalizeAngle(angle) {
    return ((angle % TAU) + TAU) % TAU;
  }

  function orbitalAngle(planet, time) {
    return normalizeAngle(planet.phase + (time * TAU) / planet.period);
  }

  function updatePlanetPositions(time) {
    for (const planet of planets.values()) {
      const angle = orbitalAngle(planet, time);
      planet.mesh.position.x = Math.cos(angle) * planet.radius;
      planet.mesh.position.y = 0;
      planet.mesh.position.z = Math.sin(angle) * planet.radius;
    }
  }

  function syncCameraTarget() {
    if (state.cameraMode === "tracking") {
      const selected = planets.get(state.selectedId);
      if (selected) cameraTarget.copy(selected.mesh.position);
    } else {
      cameraTarget.set(0, 0, 0);
    }
  }

  function updateCamera() {
    if (!camera || !cameraTarget) return;
    cameraSpherical.set(cameraDistance, cameraPhi, cameraTheta);
    cameraOffset.setFromSpherical(cameraSpherical);
    camera.position.copy(cameraTarget).add(cameraOffset);
    camera.lookAt(cameraTarget);
  }

  function updateInterface() {
    const selected = planets.get(state.selectedId);
    if (selected) {
      ui.selectedName.textContent = selected.name;
      ui.selectedPeriod.textContent = `Orbital period · ${selected.period} simulation seconds`;
      const degrees = (orbitalAngle(selected, state.time) * 180) / Math.PI;
      ui.selectedAngle.textContent = `${degrees.toFixed(1)}°`;
    }
    const displayTime = state.time.toFixed(1).padStart(5, "0");
    ui.timeReadout.innerHTML = `T+${displayTime} <small>s</small>`;
    ui.timeSlider.value = String(state.time);
    ui.timeSlider.setAttribute("aria-valuetext", `${state.time.toFixed(1)} seconds`);
    ui.speedSlider.value = String(state.speed);
    ui.speedSlider.setAttribute("aria-valuetext", `${state.speed.toFixed(2)} times speed`);
    ui.speedValue.textContent = `${state.speed.toFixed(2)}×`;
    ui.pauseButton.dataset.playing = String(state.playing);
    ui.pauseButton.setAttribute("aria-label", state.playing ? "Pause simulation" : "Resume simulation");
    ui.pauseButton.setAttribute("aria-pressed", String(!state.playing));
    ui.playLabel.textContent = state.playing ? "Pause" : "Resume";
    ui.playState.textContent = state.playing ? "PLAYING" : (state.time >= MAX_TIME ? "TIME LIMIT" : "PAUSED");
    ui.statusPill.textContent = state.playing ? "ORBITING" : (state.time >= MAX_TIME ? "TIME LIMIT" : "PAUSED");
    ui.trackButton.setAttribute("aria-pressed", String(state.cameraMode === "tracking"));
    ui.overviewButton.setAttribute("aria-pressed", String(state.cameraMode === "overview"));
    for (const button of ui.planetChoices) {
      button.setAttribute("aria-pressed", String(button.getAttribute("data-planet") === state.selectedId));
    }
  }

  function animate() {
    window.requestAnimationFrame(animate);
    const delta = clock.getDelta();
    if (state.playing && delta > 0) {
      state.time = Math.min(MAX_TIME, state.time + delta * state.speed);
      if (state.time >= MAX_TIME) {
        state.time = MAX_TIME;
        state.playing = false;
      }
      updatePlanetPositions(state.time);
      syncCameraTarget();
      updateInterface();
    } else if (state.cameraMode === "tracking") {
      syncCameraTarget();
      updateCamera();
    }
    renderer.render(scene, camera);
  }

  function onResize() {
    const width = Math.max(1, window.innerWidth);
    const height = Math.max(1, window.innerHeight);
    camera.aspect = width / height;
    camera.updateProjectionMatrix();
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 1.75));
    renderer.setSize(width, height);
  }

  function clamp(value, min, max) {
    return Math.min(max, Math.max(min, value));
  }
})();
