(() => {
  'use strict';
  const planets = [
    { id: 'aurel', name: 'Aurel', period: 12, phase: 0, radius: 4.2, size: 0.48, color: '#df9a62', detail: 'The inner world. A quick orbit in warm amber.' },
    { id: 'nereid', name: 'Nereid', period: 24, phase: 120, radius: 7.3, size: 0.76, color: '#80bdc5', detail: 'The middle world. A cool blue-green wanderer.' },
    { id: 'vesper', name: 'Vesper', period: 48, phase: 240, radius: 10.6, size: 1.02, color: '#a38acb', detail: 'The outer world. A slower path through violet.' }
  ];
  const byId = new Map(planets.map(planet => [planet.id, planet]));
  const $ = id => document.getElementById(id);
  const state = { time: 0, speed: 1, selectedPlanet: null, viewMode: 'overview', playing: true };
  const storageKey = 'orbit-observatory-2-snapshots-v1';
  const angleAt = (planet, time) => ((planet.phase + time * 360 / planet.period) % 360 + 360) % 360;
  const numberText = value => Number(value.toFixed(4)).toString();
  const timeInput = $('time');
  const speedInput = $('speed');
  const timeReadout = $('time-readout');
  const angleReadout = $('selected-angle');
  let snapshots = [];
  let captureSerial = 0;
  let renderer, scene, camera, sun, sunGlow;
  let lastTick = performance.now();
  const cameraPose = { yaw: 0.6, pitch: 0.62, distance: 33 };
  const meshes = new Map();
  const highlights = new Map();
  const sceneHost = $('scene');
  let target;

  function notice(id, text, isError = false) {
    const element = $(id);
    element.textContent = text;
    element.classList.toggle('error', isError);
  }

  function validatePayload(text) {
    let root;
    try { root = JSON.parse(text); } catch (_) { throw new Error('Invalid JSON. Check the text and try again.'); }
    const object = value => value !== null && typeof value === 'object' && !Array.isArray(value);
    if (!object(root)) throw new Error('The root must be an object with version and snapshots.');
    if (root.version !== 1) throw new Error('The version must be the number 1.');
    if (!Array.isArray(root.snapshots) || root.snapshots.length > 3) throw new Error('Snapshots must be an array with zero to three entries.');
    const ids = new Set();
    return root.snapshots.map((entry, index) => {
      const prefix = `Observation ${index + 1}: `;
      if (!object(entry)) throw new Error(prefix + 'each entry must be an object.');
      for (const field of ['id', 'time', 'speed', 'selectedPlanet', 'viewMode']) {
        if (!Object.prototype.hasOwnProperty.call(entry, field)) throw new Error(prefix + `missing ${field}.`);
      }
      if (typeof entry.id !== 'string' || entry.id.length === 0 || ids.has(entry.id)) throw new Error(prefix + 'IDs must be nonempty, unique strings.');
      if (typeof entry.time !== 'number' || !Number.isFinite(entry.time) || entry.time < 0 || entry.time > 120) throw new Error(prefix + 'time must be a number from 0 to 120.');
      if (typeof entry.speed !== 'number' || !Number.isFinite(entry.speed) || entry.speed < 0.25 || entry.speed > 4) throw new Error(prefix + 'speed must be a number from 0.25 to 4.');
      if (entry.selectedPlanet !== null && !byId.has(entry.selectedPlanet)) throw new Error(prefix + 'selectedPlanet must be null, aurel, nereid, or vesper.');
      if (entry.viewMode !== 'overview' && entry.viewMode !== 'tracking') throw new Error(prefix + 'viewMode must be overview or tracking.');
      if (entry.viewMode === 'tracking' && entry.selectedPlanet === null) throw new Error(prefix + 'tracking requires a selected planet.');
      ids.add(entry.id);
      return { id: entry.id, time: entry.time, speed: entry.speed, selectedPlanet: entry.selectedPlanet, viewMode: entry.viewMode };
    });
  }

  const payloadText = () => JSON.stringify({ version: 1, snapshots }, null, 2);

  function saveSnapshots() {
    try {
      localStorage.setItem(storageKey, payloadText());
      notice('storage-message', 'Saved in this browser. Live playback starts fresh on reload.');
    } catch (_) {
      notice('storage-message', 'Saving unavailable. Your observations work in memory, but unsaved changes will not survive reload.', true);
    }
  }

  function loadSnapshots() {
    let stored;
    try { stored = localStorage.getItem(storageKey); }
    catch (_) {
      notice('storage-message', 'Saving unavailable. Observations are in-memory only and may not survive reload.', true);
      return;
    }
    if (stored === null) return;
    try { snapshots = validatePayload(stored); }
    catch (_) {
      snapshots = [];
      notice('storage-message', 'Recovery notice: saved observations were invalid. Started with an empty archive; live controls remain available.', true);
    }
  }

  function element(tag, className, text) {
    const node = document.createElement(tag);
    if (className) node.className = className;
    if (text !== undefined) node.textContent = text;
    return node;
  }

  function renderSnapshots() {
    const list = $('snapshots');
    list.replaceChildren();
    $('snapshot-count').textContent = `${String(snapshots.length).padStart(2, '0')} / 03`;
    $('empty-snapshots').hidden = snapshots.length !== 0;
    snapshots.forEach((entry, index) => {
      const row = element('li', 'snapshot');
      row.dataset.snapshotId = entry.id;
      const heading = element('div', 'snapshot-heading');
      heading.append(element('strong', '', `${numberText(entry.time)} s`), element('span', '', entry.viewMode));
      const idLine = element('p', 'snapshot-id', `ID: ${entry.id}`);
      const details = element('p', 'snapshot-values', `${numberText(entry.speed)}× · ${entry.selectedPlanet === null ? 'No selection' : byId.get(entry.selectedPlanet).name}`);
      const angles = element('p', 'snapshot-angles');
      planets.forEach((planet, planetIndex) => {
        if (planetIndex) angles.append(document.createElement('br'));
        angles.append(document.createTextNode(`${planet.name} / ${angleAt(planet, entry.time).toFixed(2)}°`));
      });
      const actions = element('div', 'snapshot-actions');
      const restore = element('button', '', 'Restore');
      restore.setAttribute('aria-label', `Restore observation ${index + 1}, ${numberText(entry.time)} seconds`);
      restore.addEventListener('click', () => restoreSnapshot(entry));
      const remove = element('button', '', 'Delete');
      remove.setAttribute('aria-label', `Delete observation ${index + 1}`);
      remove.addEventListener('click', () => {
        snapshots = snapshots.filter(item => item.id !== entry.id);
        renderSnapshots();
        saveSnapshots();
        notice('snapshot-message', 'Observation deleted. Remaining observations are unchanged.');
        $('capture').focus();
      });
      actions.append(restore, remove);
      row.append(heading, idLine, details, angles, actions);
      list.append(row);
    });
  }

  function captureSnapshot() {
    if (snapshots.length >= 3) {
      notice('snapshot-message', 'Archive full: three observations maximum. Delete one before capturing another.', true);
      return;
    }
    let id;
    do { id = `obs-${Date.now().toString(36)}-${(++captureSerial).toString(36)}`; }
    while (snapshots.some(entry => entry.id === id));
    snapshots.push({ id, time: state.time, speed: state.speed, selectedPlanet: state.selectedPlanet, viewMode: state.viewMode });
    renderSnapshots();
    saveSnapshots();
    notice('snapshot-message', 'Observation captured. Live state is unchanged.');
  }

  function standardCamera() {
    cameraPose.yaw = 0.6;
    cameraPose.pitch = 0.62;
    cameraPose.distance = state.viewMode === 'tracking' ? 9 : 33;
  }

  function restoreSnapshot(entry) {
    state.time = entry.time;
    state.speed = entry.speed;
    state.selectedPlanet = entry.selectedPlanet;
    state.viewMode = entry.viewMode;
    state.playing = false;
    lastTick = performance.now();
    standardCamera();
    updateScene();
    updateControls();
    notice('snapshot-message', `Restored ${numberText(entry.time)} s. Playback paused; standard ${entry.viewMode} framing.`);
  }

  function updateReadouts() {
    timeReadout.innerHTML = `${state.time.toFixed(2).padStart(6, '0')} <small>s</small>`;
    timeInput.value = state.time;
    timeInput.setAttribute('aria-valuetext', `${state.time.toFixed(2)} simulation seconds`);
    if (state.selectedPlanet) angleReadout.textContent = `${angleAt(byId.get(state.selectedPlanet), state.time).toFixed(2)}°`;
  }

  function updateControls() {
    $('play-label').textContent = state.playing ? 'Pause' : 'Resume';
    $('play-icon').textContent = state.playing ? 'Ⅱ' : '▷';
    $('play').setAttribute('aria-label', state.playing ? 'Pause simulation' : 'Resume simulation');
    $('play').disabled = !state.playing && state.time >= 120;
    $('play-status').textContent = state.time >= 120 ? 'END OF ORBIT' : state.playing ? 'RUNNING' : 'PAUSED';
    speedInput.value = state.speed;
    $('speed-readout').textContent = `${numberText(state.speed)}×`;
    speedInput.setAttribute('aria-valuetext', `${numberText(state.speed)} times normal speed`);
    $('track').disabled = !state.selectedPlanet;
    $('track').setAttribute('aria-pressed', String(state.viewMode === 'tracking'));
    $('overview').setAttribute('aria-pressed', String(state.viewMode === 'overview'));
    const selected = byId.get(state.selectedPlanet);
    $('selected-name').textContent = selected ? selected.name : 'No planet selected';
    $('selection-hint').textContent = selected ? selected.detail : 'Choose a planet in the scene or the list above.';
    $('selected-period').textContent = selected ? `${selected.period} sim seconds` : '—';
    if (!selected) angleReadout.textContent = '—';
    $('view-label').textContent = selected && state.viewMode === 'tracking' ? `TRACKING / ${selected.name.toUpperCase()}` : 'FULL SYSTEM';
    document.querySelectorAll('[data-planet]').forEach(button => button.setAttribute('aria-pressed', String(button.dataset.planet === state.selectedPlanet)));
    updateReadouts();
  }

  function selectPlanet(id) {
    if (!byId.has(id)) return;
    state.selectedPlanet = id;
    updateScene();
    updateControls();
  }

  $('play').addEventListener('click', () => {
    if (state.time >= 120) return;
    state.playing = !state.playing;
    lastTick = performance.now();
    updateControls();
  });
  timeInput.addEventListener('input', () => {
    state.time = Number(timeInput.value);
    state.playing = false;
    lastTick = performance.now();
    updateScene();
    updateControls();
  });
  speedInput.addEventListener('input', () => {
    state.speed = Number(speedInput.value);
    updateControls();
  });
  document.querySelectorAll('[data-planet]').forEach(button => button.addEventListener('click', () => selectPlanet(button.dataset.planet)));
  $('track').addEventListener('click', () => {
    if (!state.selectedPlanet) return;
    state.viewMode = 'tracking';
    standardCamera();
    updateScene();
    updateControls();
  });
  $('overview').addEventListener('click', () => {
    state.viewMode = 'overview';
    standardCamera();
    updateScene();
    updateControls();
  });
  $('capture').addEventListener('click', captureSnapshot);
  $('export').addEventListener('click', () => {
    $('snapshot-json').value = payloadText();
    notice('json-message', `Exported ${snapshots.length} observations. Archive and live scene are unchanged.`);
  });
  $('import').addEventListener('click', () => {
    let next;
    try { next = validatePayload($('snapshot-json').value); }
    catch (error) { notice('json-message', error.message + ' Nothing was changed.', true); return; }
    snapshots = next;
    renderSnapshots();
    saveSnapshots();
    notice('json-message', `Imported ${snapshots.length} observations. Live scene is unchanged.`);
  });

  function buildScene() {
    scene = new THREE.Scene();
    scene.background = new THREE.Color('#090d14');
    camera = new THREE.PerspectiveCamera(43, 1, 0.1, 220);
    target = new THREE.Vector3();
    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: false });
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
    renderer.outputEncoding = THREE.sRGBEncoding;
    renderer.domElement.setAttribute('aria-label', 'Solar system with one sun and three planets');
    renderer.domElement.setAttribute('role', 'img');
    sceneHost.append(renderer.domElement);
    scene.add(new THREE.AmbientLight(0x9eb8e8, 0.5));
    const sunLight = new THREE.PointLight(0xffddb1, 1.9, 90, 1.2);
    scene.add(sunLight);
    const rimLight = new THREE.DirectionalLight(0xb6ceff, 0.45);
    rimLight.position.set(-10, 15, 10);
    scene.add(rimLight);
    sun = new THREE.Mesh(new THREE.SphereGeometry(1.45, 48, 32), new THREE.MeshBasicMaterial({ color: 0xffd39a }));
    scene.add(sun);
    const glowCanvas = document.createElement('canvas');
    glowCanvas.width = glowCanvas.height = 128;
    const context = glowCanvas.getContext('2d');
    const glow = context.createRadialGradient(64, 64, 3, 64, 64, 64);
    glow.addColorStop(0, '#ffe0ac'); glow.addColorStop(0.23, '#ffba6270'); glow.addColorStop(0.55, '#e8762220'); glow.addColorStop(1, '#d6711500');
    context.fillStyle = glow; context.fillRect(0, 0, 128, 128);
    sunGlow = new THREE.Sprite(new THREE.SpriteMaterial({ map: new THREE.CanvasTexture(glowCanvas), transparent: true, blending: THREE.AdditiveBlending, depthWrite: false }));
    sunGlow.scale.set(9, 9, 1);
    scene.add(sunGlow);
    let seed = 8128;
    const random = () => { seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0; return seed / 4294967296; };
    const stars = new Float32Array(1650);
    for (let i = 0; i < stars.length; i += 3) {
      const azimuth = random() * Math.PI * 2;
      const vertical = random() * 2 - 1;
      const distance = 55 + random() * 45;
      const spread = Math.sqrt(1 - vertical * vertical);
      stars[i] = Math.cos(azimuth) * spread * distance;
      stars[i + 1] = vertical * distance;
      stars[i + 2] = Math.sin(azimuth) * spread * distance;
    }
    const starGeometry = new THREE.BufferGeometry();
    starGeometry.setAttribute('position', new THREE.BufferAttribute(stars, 3));
    scene.add(new THREE.Points(starGeometry, new THREE.PointsMaterial({ color: 0x9bb5d5, size: 0.085, transparent: true, opacity: 0.65, sizeAttenuation: true })));
    planets.forEach(planet => {
      const orbitPoints = [];
      for (let i = 0; i < 192; i++) {
        const angle = i / 192 * Math.PI * 2;
        orbitPoints.push(new THREE.Vector3(Math.cos(angle) * planet.radius, 0, Math.sin(angle) * planet.radius));
      }
      const orbit = new THREE.LineLoop(new THREE.BufferGeometry().setFromPoints(orbitPoints), new THREE.LineBasicMaterial({ color: 0x40506a, transparent: true, opacity: 0.54 }));
      scene.add(orbit);
      const textureCanvas = document.createElement('canvas');
      textureCanvas.width = 256; textureCanvas.height = 128;
      const paint = textureCanvas.getContext('2d');
      paint.fillStyle = planet.color; paint.fillRect(0, 0, 256, 128);
      for (let band = 0; band < 38; band++) {
        paint.fillStyle = band % 3 === 0 ? '#ffffff14' : '#0000000d';
        paint.fillRect(0, random() * 128, 256, 1 + random() * 6);
      }
      const texture = new THREE.CanvasTexture(textureCanvas);
      texture.encoding = THREE.sRGBEncoding;
      const mesh = new THREE.Mesh(new THREE.SphereGeometry(planet.size, 40, 28), new THREE.MeshStandardMaterial({ map: texture, roughness: 0.85, metalness: 0.02 }));
      mesh.userData.planetId = planet.id;
      mesh.rotation.z = 0.12;
      meshes.set(planet.id, mesh);
      scene.add(mesh);
      const marker = new THREE.Mesh(new THREE.TorusGeometry(planet.size * 1.45, 0.025, 8, 80), new THREE.MeshBasicMaterial({ color: 0xffd8a6, transparent: true, opacity: 0.95, depthTest: false }));
      marker.renderOrder = 2;
      marker.visible = false;
      highlights.set(planet.id, marker);
      scene.add(marker);
      if (planet.id === 'vesper') {
        const ring = new THREE.Mesh(new THREE.RingGeometry(1.27, 1.8, 72), new THREE.MeshStandardMaterial({ color: 0xa197b8, side: THREE.DoubleSide, transparent: true, opacity: 0.56, roughness: 1 }));
        ring.rotation.x = Math.PI / 2.5;
        mesh.add(ring);
      }
    });
    installCameraInput();
    new ResizeObserver(resizeScene).observe(sceneHost);
    resizeScene();
  }

  function resizeScene() {
    if (!renderer) return;
    const width = sceneHost.clientWidth;
    const height = sceneHost.clientHeight;
    if (!width || !height) return;
    renderer.setSize(width, height, false);
    camera.aspect = width / height;
    camera.updateProjectionMatrix();
    updateScene();
  }

  function updateScene() {
    if (!renderer) return;
    planets.forEach(planet => {
      const angle = angleAt(planet, state.time) * Math.PI / 180;
      const mesh = meshes.get(planet.id);
      mesh.position.set(Math.cos(angle) * planet.radius, 0, Math.sin(angle) * planet.radius);
      mesh.rotation.y = state.time * 0.12;
    });
    if (state.viewMode === 'tracking' && state.selectedPlanet) target.copy(meshes.get(state.selectedPlanet).position);
    else target.set(0, 0, 0);
    const horizontal = Math.cos(cameraPose.pitch) * cameraPose.distance;
    camera.position.set(target.x + Math.sin(cameraPose.yaw) * horizontal, target.y + Math.sin(cameraPose.pitch) * cameraPose.distance, target.z + Math.cos(cameraPose.yaw) * horizontal);
    camera.lookAt(target);
    highlights.forEach((marker, id) => {
      marker.visible = id === state.selectedPlanet;
      marker.position.copy(meshes.get(id).position);
      marker.quaternion.copy(camera.quaternion);
    });
    renderer.render(scene, camera);
  }

  function installCameraInput() {
    const canvas = renderer.domElement;
    const pointerPositions = new Map();
    let anchor = null;
    let moved = false;
    let pinchDistance = null;
    const raycaster = new THREE.Raycaster();
    const pointer = new THREE.Vector2();
    const planetMeshes = Array.from(meshes.values());
    const zoom = scale => { cameraPose.distance = Math.max(3.5, Math.min(75, cameraPose.distance * scale)); };
    canvas.addEventListener('pointerdown', event => {
      if (event.button !== 0 && event.pointerType === 'mouse') return;
      canvas.setPointerCapture(event.pointerId);
      pointerPositions.set(event.pointerId, { x: event.clientX, y: event.clientY });
      if (pointerPositions.size === 1) { anchor = { x: event.clientX, y: event.clientY }; moved = false; }
      else { moved = true; pinchDistance = null; }
    });
    canvas.addEventListener('pointermove', event => {
      const previous = pointerPositions.get(event.pointerId);
      if (!previous) return;
      const dx = event.clientX - previous.x;
      const dy = event.clientY - previous.y;
      pointerPositions.set(event.pointerId, { x: event.clientX, y: event.clientY });
      if (pointerPositions.size > 1) {
        const [first, second] = pointerPositions.values();
        const distance = Math.hypot(first.x - second.x, first.y - second.y);
        if (pinchDistance && distance) zoom(pinchDistance / distance);
        pinchDistance = distance;
        moved = true;
      } else if (anchor && (moved || Math.hypot(event.clientX - anchor.x, event.clientY - anchor.y) > 5)) {
        moved = true;
        cameraPose.yaw -= dx * 0.006;
        cameraPose.pitch = Math.max(-1.25, Math.min(1.4, cameraPose.pitch + dy * 0.006));
      }
      updateScene();
    });
    canvas.addEventListener('pointerup', event => {
      if (!pointerPositions.has(event.pointerId)) return;
      if (!moved && pointerPositions.size === 1 && anchor && Math.hypot(event.clientX - anchor.x, event.clientY - anchor.y) <= 5) {
        const rect = canvas.getBoundingClientRect();
        pointer.set((event.clientX - rect.left) / rect.width * 2 - 1, -(event.clientY - rect.top) / rect.height * 2 + 1);
        raycaster.setFromCamera(pointer, camera);
        const hits = raycaster.intersectObjects(planetMeshes, false);
        if (hits.length) selectPlanet(hits[0].object.userData.planetId);
      }
      pointerPositions.delete(event.pointerId);
      pinchDistance = null;
      if (!pointerPositions.size) anchor = null;
    });
    canvas.addEventListener('pointercancel', event => { pointerPositions.delete(event.pointerId); moved = true; pinchDistance = null; });
    canvas.addEventListener('wheel', event => {
      event.preventDefault();
      zoom(Math.exp(Math.max(-250, Math.min(250, event.deltaY)) * 0.0015));
      updateScene();
    }, { passive: false });
  }

  loadSnapshots();
  renderSnapshots();
  try { buildScene(); }
  catch (error) {
    $('scene-error').hidden = false;
    $('scene-error').textContent = `The 3D renderer could not start: ${error.message}. Enable WebGL in your browser to view the system.`;
    console.error(error);
  }
  updateControls();
  lastTick = performance.now();
  function animate(now) {
    const delta = Math.max(0, (now - lastTick) / 1000);
    lastTick = now;
    if (state.playing) {
      state.time = Math.min(120, state.time + delta * state.speed);
      if (state.time === 120) { state.playing = false; updateControls(); }
    }
    updateScene();
    updateReadouts();
    requestAnimationFrame(animate);
  }
  requestAnimationFrame(animate);
})();
