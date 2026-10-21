(async function probeObservatory({ frame, tab, profile, scenario, time = 5.2, playing = false, isolatedOrigin = false }) {
  const scenarios = ['overview', 'persistence', 'invalid-storage', 'quota', 'diagnostics'];
  if (!scenarios.includes(scenario)) throw new Error(`Unsupported probe scenario: ${scenario}`);
  if (profile.id !== 'terra-xhigh-observatory-2') throw new Error('Unsupported control profile');
  const url = new URL(await frame.url());
  if (url.protocol !== 'http:' || url.hostname !== '127.0.0.1' || !isolatedOrigin) {
    throw new Error('Use an isolated, probe-owned loopback origin');
  }
  const report = {
    format_version: '1.0', scope: 'focused-scenarios-only', profile: profile.id,
    scenario, url: url.href, started_at: new Date().toISOString(), observations: [],
    cleanup: { storage_restored: null, filler_removed: null },
    limits: ['This is not a complete rubric review. A passing scenario does not certify its entire check.',
      'Rendered framing requires separate visual observation. Storage quota does not prove read-denial behavior.'],
  };
  const same = (a, b) => JSON.stringify(a) === JSON.stringify(b);
  const observe = (id, check, passed, action, expected, observed) => {
    report.observations.push({ scenario_id: id, check_id: check,
      status: passed === null ? 'unverified' : passed ? 'pass' : 'fail', action, expected, observed });
  };
  const refresh = async () => { console.log(await frame.snapshot()); };
  const click = async selector => {
    await frame.evaluate(selector => document.querySelector(selector).scrollIntoView({ block: 'center' }), selector);
    await frame.click(selector);
    await refresh();
  };
  const live = () => frame.evaluate(p => ({
    time: Number(document.querySelector(p.time).value), speed: Number(document.querySelector(p.speed).value),
    playback: document.querySelector(p.pause).textContent,
    selected: document.querySelector(p.selection)?.getAttribute('aria-label') ?? null,
    mode: document.querySelector(p.camera).textContent,
  }), profile);
  const notices = () => frame.evaluate(p => ({
    snapshot: document.querySelector(p.snapshot_notice).textContent,
    json: document.querySelector(p.json_notice).textContent,
  }), profile);
  const persisted = () => frame.evaluate(key => localStorage.getItem(key), profile.storage_key);
  const exportList = async () => {
    await click(profile.export);
    return JSON.parse(await frame.evaluate(selector => document.querySelector(selector).value, profile.json));
  };
  const importList = async payload => {
    await frame.fill(profile.json, JSON.stringify(payload));
    await click(profile.import);
  };
  const reload = async () => {
    await frame.reload();
    await frame.waitForSelector(profile.capture, { state: 'visible', timeout: 15000 });
    await refresh();
  };
  const preparePaused = async () => {
    await frame.fill(profile.time, '17.25');
    await frame.fill(profile.speed, '1.5');
    await click(profile.planet);
    await click(profile.tracking);
  };
  report.viewport = await frame.evaluate(() => ({ width: innerWidth, height: innerHeight, secure: isSecureContext }));
  if (!report.viewport.secure) throw new Error('The probe viewport must be a secure context');
  await refresh();

  if (scenario === 'diagnostics') {
    const wrapper = new URL(await tab.url());
    if (wrapper.hostname !== '127.0.0.1' || wrapper.origin === url.origin) {
      throw new Error('Diagnostic control requires a separate owned scratch origin');
    }
    await tab.cdp('Runtime.enable');
    await tab.events();
    const marker = `observatory-diagnostic-${Date.now()}`;
    await tab.evaluate(marker => new Promise(resolve => {
      const script = document.createElement('script');
      script.textContent = `console.error(${JSON.stringify(marker)});setTimeout(()=>{throw new Error(${JSON.stringify(marker)})},0);`;
      document.body.append(script);
      script.remove();
      setTimeout(resolve, 250);
    }), marker);
    const events = await tab.events();
    const consoleEvents = events.filter(event => event.method === 'Runtime.consoleAPICalled' && JSON.stringify(event.params).includes(marker));
    const pageErrors = events.filter(event => event.method === 'Runtime.exceptionThrown' && JSON.stringify(event.params).includes(marker));
    observe('diagnostic-positive-control', 'D3', consoleEvents.length > 0 && pageErrors.length > 0 ? true : null,
      'Enable Runtime events, then emit console.error and an asynchronous exception on the owned scratch page.',
      'Both diagnostic channels capture their deliberate signal before empty logs can be trusted.',
      JSON.stringify({ consoleEvents, pageErrors, limitation: 'This tests transport, not artifact error-free execution.' }));
    report.finished_at = new Date().toISOString();
    return report;
  }

  if (scenario === 'overview') {
    await frame.fill(profile.time, String(time));
    await frame.fill(profile.speed, '1.5');
    await click(profile.planet);
    await click(profile.tracking);
    const setup = await live();
    const setupReached = setup.time === time && setup.speed === 1.5
      && setup.selected === 'Select Cinder, ID cinder' && setup.playback === 'Resume' && /tracking/i.test(setup.mode);
    if (playing) await click(profile.pause);
    const before = await live();
    const started = Date.now();
    await click(profile.overview);
    const after = await live();
    const elapsed = (Date.now() - started) / 1000;
    const timePreserved = playing
      ? after.time >= before.time && after.time - before.time <= elapsed * before.speed + 0.25
      : after.time === before.time;
    const passed = setupReached && before.playback === (playing ? 'Pause' : 'Resume')
      && timePreserved && after.speed === before.speed && after.selected === before.selected
      && after.playback === before.playback && /overview/i.test(after.mode);
    observe(`overview-state-${time}-${playing ? 'playing' : 'paused'}`, 'T4', passed,
      'Seek, select Cinder, track, then return to overview through the controls.',
      'The requested time, speed, selection and tracking mode are reached; overview preserves those values and playback.', JSON.stringify({ setup, setupReached, before, after, elapsed }));
    await frame.evaluate(selector => document.querySelector(selector).scrollIntoView({ block: 'start' }), profile.canvas);
    observe(`overview-framing-${time}-${playing ? 'playing' : 'paused'}`, 'T4', null,
      'Prepare the actual overview scene for screenshot inspection.',
      'The sun, all three planets and their full orbits fit the rendered canvas.',
      'Visual observation is required. DOM labels and state preservation do not prove framing.');
    report.finished_at = new Date().toISOString();
    return report;
  }

  const original = await persisted();
  const fillerKey = `observatory-probe-${Date.now()}-${Math.random().toString(36).slice(2)}`;
  let fillerCreated = false;
  try {
    if (scenario === 'persistence') {
      await importList({ version: 1, snapshots: [] });
      await preparePaused();
      await click(profile.capture);
      const captured = await exportList();
      await reload();
      const captureReload = await exportList();
      const defaults = await live();
      observe('capture-persistence', 'D6', captured.snapshots.length === 1 && same(captured, captureReload),
        'Capture through the UI, export, reload on the same origin, and export again.',
        'Captured IDs, order, times, speeds, selections and modes persist exactly.', JSON.stringify({ captured, reloaded: captureReload }));
      observe('reload-live-defaults', 'D6', defaults.time >= 0 && defaults.time < 120 && defaults.speed === 1
        && defaults.playback === 'Pause' && defaults.selected === null && /overview/i.test(defaults.mode),
        'Read live controls after snapshot persistence reload.',
        'Live controls restart playing at 1x with no selection and overview; saved live state is not restored.',
        JSON.stringify(defaults));
      await importList(profile.seed);
      await reload();
      const imported = await exportList();
      observe('import-persistence', 'D6', same(imported, profile.seed),
        'Import fractional and null-selection snapshots through the textarea, then reload and export.',
        'The exact imported list persists.', JSON.stringify(imported));
      await click(profile.delete);
      const deleted = await exportList();
      await reload();
      const deleteReload = await exportList();
      observe('delete-persistence', 'D6', same(deleted, { version: 1, snapshots: [profile.seed.snapshots[1]] }) && same(deleted, deleteReload),
        'Delete the first snapshot, export, reload and export again.',
        'Only the chosen snapshot is removed and the remaining entry persists unchanged.', JSON.stringify({ deleted, reloaded: deleteReload }));
    }
    if (scenario === 'invalid-storage') {
      const invalid = [
        ['malformed', '{'], ['version', JSON.stringify({ ...profile.seed, version: 2 })],
        ['unknown-planet', JSON.stringify({ version: 1, snapshots: [{ ...profile.seed.snapshots[0], selectedPlanet: 'constructor' }] })],
        ['duplicate-id', JSON.stringify({ version: 1, snapshots: [profile.seed.snapshots[0], profile.seed.snapshots[0]] })],
      ];
      for (const [name, value] of invalid) {
        await frame.evaluate(({ key, value }) => localStorage.setItem(key, value), { key: profile.storage_key, value });
        await reload();
        const feedback = await notices();
        const list = await exportList();
        await preparePaused();
        const state = await live();
        observe(`invalid-storage-${name}`, 'D6', list.snapshots.length === 0
          && /invalid|unavailable/i.test(feedback.snapshot) && state.time === 17.25 && state.selected !== null,
          `Reload with ${name} data at only the application-owned snapshot key, then use live controls.`,
          'Invalid saved data produces an empty list, a visible recovery notice and usable controls.',
          JSON.stringify({ list, feedback, state }));
      }
    }
    if (scenario === 'quota') {
      await importList(profile.seed);
      await preparePaused();
      const before = await live();
      await frame.evaluate(key => {
        if (localStorage.getItem(key) !== null) throw new Error('Quota fixture key already exists');
        localStorage.setItem(key, '');
      }, fillerKey);
      fillerCreated = true;
      const filled = await frame.evaluate(({ key, storageKey }) => {
        let low = 0, high = 8 * 1024 * 1024, failures = 0;
        while (low < high) {
          const size = Math.ceil((low + high) / 2);
          try { localStorage.setItem(key, 'q'.repeat(size)); low = size; }
          catch (error) { if (error.name !== 'QuotaExceededError') throw error; high = size - 1; failures++; }
        }
        return { characters: low, failures, stored: localStorage.getItem(storageKey) };
      }, { key: fillerKey, storageKey: profile.storage_key });
      await click(profile.capture);
      const captureNotice = await notices();
      const afterCapture = await exportList();
      const captureStored = await persisted();
      const captureState = await live();
      const actualWriteFailure = filled.failures > 0 && afterCapture.snapshots.length === 3
        && captureStored === filled.stored && !same(afterCapture, JSON.parse(captureStored));
      observe('quota-capture-warning', 'D6', actualWriteFailure
        ? /saving.*unavailable|memory.only|page only/i.test(captureNotice.snapshot) && same(before, captureState) : null,
        'Fill one isolated key within 8MiB, then capture a third snapshot through the UI.',
        'A proven failed write leaves a visible saving-unavailable warning and unchanged live state.',
        JSON.stringify({ filled, actualWriteFailure, memory: afterCapture, notice: captureNotice, state: captureState }));
      await click(profile.delete);
      const deletionNotice = await notices();
      const deleted = await exportList();
      const deletionStored = await persisted();
      const deletionDurable = same(deleted, JSON.parse(deletionStored));
      observe('quota-delete-feedback', 'D6', deleted.snapshots.length === 2
        && (deletionDurable || /saving.*unavailable|memory.only|page only/i.test(deletionNotice.snapshot)),
        'Delete a snapshot after the proven failed capture write and compare memory with persisted data.',
        'Deletion either persists or reports memory-only operation without a false saving promise.',
        JSON.stringify({ deleted, stored: JSON.parse(deletionStored), durable: deletionDurable, notice: deletionNotice }));
      const larger = { version: 1, snapshots: [{ ...profile.seed.snapshots[0], id: 'probe-large-' + 'x'.repeat(2000) }] };
      await importList(larger);
      const importNotice = await notices();
      const imported = await exportList();
      const importStored = await persisted();
      const importDurable = same(imported, JSON.parse(importStored));
      observe('quota-import-feedback', 'D6', same(imported, larger)
        && (importDurable || /saving.*unavailable|memory.only|page only/i.test(importNotice.json)),
        'Import a larger valid payload after saving became unavailable.',
        'The JSON panel reports memory-only operation unless the exact import really persisted.',
        JSON.stringify({ imported: same(imported, larger), durable: importDurable, notice: importNotice }));
      await frame.fill(profile.time, '73');
      await frame.fill(profile.speed, '0.25');
      await click(profile.overview);
      await click(profile.pause);
      const beforeRestore = await live();
      await click(profile.restore);
      const restored = await live();
      observe('quota-memory-restore', 'D6', beforeRestore.time !== 17.25 && beforeRestore.speed === 0.25
        && beforeRestore.playback === 'Pause' && /overview/i.test(beforeRestore.mode)
        && restored.time === 17.25 && restored.speed === 1.5
        && restored.playback === 'Resume' && restored.selected === 'Select Cinder, ID cinder' && /tracking/i.test(restored.mode),
        'Change time, speed, camera mode and playback, then restore the imported in-memory snapshot after failed saving.',
        'Snapshot restoration replaces the different live state and pauses at the captured state.', JSON.stringify({ beforeRestore, restored }));
    }
  } finally {
    await frame.evaluate(({ fillerKey, key, original, fillerCreated }) => {
      if (fillerCreated) localStorage.removeItem(fillerKey);
      if (original === null) localStorage.removeItem(key);
      else localStorage.setItem(key, original);
    }, { fillerKey, key: profile.storage_key, original, fillerCreated });
    report.cleanup = await frame.evaluate(({ fillerKey, key, original, fillerCreated }) => ({
      storage_restored: localStorage.getItem(key) === original,
      filler_removed: fillerCreated ? localStorage.getItem(fillerKey) === null : null,
    }), { fillerKey, key: profile.storage_key, original, fillerCreated });
  }
  report.finished_at = new Date().toISOString();
  return report;
})
