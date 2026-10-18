(() => {
  'use strict';

  const snapshot = globalThis.BENCHMARK_GALLERY;
  const hasData = Boolean(snapshot && typeof snapshot.revision === 'string' && Array.isArray(snapshot.runs));
  const runs = hasData ? snapshot.runs : [];
  const generateGuidance = 'ai-benchmark 폴더에서 node scripts/generate-gallery.mjs 를 실행한 뒤 이 페이지를 새로고침하세요.';
  const loadError = `gallery/data.js 데이터를 불러올 수 없습니다. ${generateGuidance}`;
  let detailView = null;
  const activeSections = new Map();
  const $ = id => document.getElementById(id);
  const number = new Intl.NumberFormat('ko-KR');
  const dateFormat = new Intl.DateTimeFormat('ko-KR', {
    timeZone: 'Asia/Seoul', year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', hour12: false,
  });
  const categoryNames = {
    simulation: '시뮬레이션 · 시간', selection: '선택 · 추적', usability: '조작 · 상호작용',
    visual: '시각적 완성도', responsive: '반응형 사용성', delivery: '전달 안정성',
  };
  const statuses = { completed: '생성 완료', running: '생성 중', failed: '생성 실패', blocked: '진행 불가', unknown: '상태 미확인' };
  const reviewStatuses = { complete: '평가 완료', partial: '부분 평가', blocked: '평가 불가' };
  const checkStatuses = { pass: '통과', fail: '실패', unverified: '미검증' };
  const sections = [['preview', '결과물'], ['review', '점수 · 평가'], ['prompt', '사용한 프롬프트'], ['record', '실행 기록']];

  function element(tag, className, text) {
    const node = document.createElement(tag);
    if (className) node.className = className;
    if (text !== undefined && text !== null) node.textContent = text;
    return node;
  }

  function scoreRange(review, lowerKey = 'total_score', upperKey = 'total_score_upper_bound', maximum = 100) {
    if (!['2.0', '4.0'].includes(review?.format_version)) return null;
    const lower = review?.[lowerKey];
    const upper = review?.format_version === '4.0' ? review[upperKey] : lower;
    return Number.isFinite(lower) && Number.isFinite(upper) && lower >= 0 && lower <= upper && upper <= maximum
      ? { lower, upper } : null;
  }

  function score(run) {
    return scoreRange(run.review)?.lower ?? null;
  }

  function rangeText(range) {
    return range.lower === range.upper ? number.format(range.lower) : `${number.format(range.lower)} ~ ${number.format(range.upper)}`;
  }

  function scoreDescription(run) {
    const range = scoreRange(run.review);
    if (!range) return `미채점 · ${reviewLabel(run)}`;
    const basis = run.review.format_version === '4.0'
      ? range.lower === range.upper ? '확정 점수' : '보수적 하한'
      : '저장된 점수';
    return `${number.format(range.lower)}점 / 100 · ${basis}${range.lower !== range.upper ? ` · 가능 범위 ${rangeText(range)}` : ''} · ${reviewLabel(run)}`;
  }

  function date(value) {
    return value && Number.isFinite(Date.parse(value)) ? dateFormat.format(new Date(value)) : '시각 미확인';
  }

  function duration(value) {
    if (value === null) return '미기록';
    const seconds = Math.round(value / 1000);
    return seconds < 60 ? `${seconds}초` : `${Math.floor(seconds / 60)}분 ${seconds % 60}초`;
  }

  function route() {
    const params = new URLSearchParams(location.hash.slice(1));
    return { model: params.has('model') ? params.get('model') : null, run: params.get('run') };
  }

  function routeUrl(model, run = null) {
    const params = new URLSearchParams();
    if (model !== null) params.set('model', model);
    if (run !== null) params.set('run', run);
    return `#${params}`;
  }

  function reviewLabel(run) {
    if (!run.review) return '평가 없음';
    return reviewStatuses[run.review.status] ?? '평가 기록 있음';
  }

  function scoreBlock(run, className = 'run-score') {
    const wrapper = element('div', className);
    const range = scoreRange(run.review);
    wrapper.setAttribute('role', 'group');
    wrapper.setAttribute('aria-label', scoreDescription(run));
    wrapper.append(element('span', `score${range === null ? ' unscored' : ''}`, range === null ? '미채점' : number.format(range.lower)));
    const basis = range && run.review.format_version === '4.0'
      ? range.lower === range.upper ? '확정 점수' : '보수적 하한'
      : '저장된 점수';
    wrapper.append(element('span', 'score-caption', range === null ? reviewLabel(run) : `/ 100 · ${basis} · ${reviewLabel(run)}`));
    if (range && range.lower !== range.upper) wrapper.append(element('span', 'score-caption', `가능 범위 ${rangeText(range)}`));
    return wrapper;
  }

  function renderModels(selected) {
    const models = [...new Set(runs.map(run => run.model ?? ''))].sort();
    $('model-count').textContent = String(models.length).padStart(2, '0');
    if (selected !== null && !models.includes(selected)) models.push(selected);
    const existing = new Map([...$('models').children].map(button => [button.dataset.focusKey, button]));
    const buttons = [null, ...models].map(model => {
      const key = `model:${routeUrl(model)}`;
      let button = existing.get(key);
      if (!button) {
        button = element('button', 'model-button');
        button.type = 'button';
        button.dataset.focusKey = key;
        button.addEventListener('click', () => {
          $('search').value = '';
          $('review-filter').value = 'all';
          const url = routeUrl(model);
          if (location.hash === url || (!location.hash && url === '#')) render();
          else location.hash = url;
        });
      }
      button.setAttribute('aria-pressed', String(model === selected));
      const name = element('span', 'model-name', model === null ? '모든 모델' : model || '모델 미확인');
      if (model?.includes('/')) {
        name.textContent = model.split('/').at(-1);
        name.append(element('small', '', model.slice(0, model.lastIndexOf('/'))));
      }
      button.title = model === null ? '모든 모델' : model || '모델 미확인';
      button.replaceChildren(element('span', 'model-symbol', model === null ? '▦' : '◈'), name, element('span', 'model-count', model === null ? runs.length : runs.filter(run => (run.model ?? '') === model).length));
      return button;
    });
    reconcileChildren($('models'), buttons);
  }

  function reconcileChildren(parent, children) {
    const retained = new Set(children);
    for (const child of [...parent.children]) {
      if (!retained.has(child)) child.remove();
    }
    children.forEach((child, index) => {
      if (parent.children[index] !== child) parent.insertBefore(child, parent.children[index] ?? null);
    });
  }

  function renderStats(selectedRuns) {
    const scored = selectedRuns.filter(run => score(run) !== null);
    const scores = scored.map(score);
    const stats = [
      ['실행 기록', selectedRuns.length, '저장된 시도'],
      ['생성된 결과물', selectedRuns.filter(run => run.artifactUrl).length, 'HTML 결과'],
      ['최고 표시 점수', scores.length ? number.format(Math.max(...scores)) : '—', `${scored.length}개 숫자 점수 · 확정값 또는 보수적 하한 / 100 · 검증 완료와 별개`],
    ];
    $('stats').replaceChildren(...stats.map(([label, value, hint]) => {
      const card = element('div', 'stat');
      card.append(element('span', 'stat-label', label), element('span', 'stat-value', value), element('span', 'stat-hint', hint));
      return card;
    }));
  }

  function renderList() {
    const { model } = route();
    const selectedRuns = runs.filter(run => model === null || (run.model ?? '') === model);
    $('collection-title').replaceChildren(document.createTextNode(model === null ? '모든 실행 결과' : model || '모델 미확인'), element('span', 'title-dot', '.'));
    $('collection-description').textContent = model === null
      ? '같은 과제, 다른 모델. 결과물을 직접 살펴보세요.'
      : 'effort별 실행과 저장된 평가를 비교하고, 생성된 결과물을 확인하세요.';
    renderStats(selectedRuns);
    const query = $('search').value.trim().toLocaleLowerCase();
    const filter = $('review-filter').value;
    const visible = selectedRuns.filter(run => {
      if (filter === 'scored' && score(run) === null) return false;
      if (filter === 'unscored' && score(run) !== null) return false;
      return `${run.model ?? ''} ${run.effort ?? ''} ${run.agent ?? ''} ${run.id}`.toLocaleLowerCase().includes(query);
    });
    if ($('sort').value === 'score') {
      visible.sort((a, b) => (score(b) ?? -1) - (score(a) ?? -1));
    }
    $('result-count').textContent = `${visible.length}개`;
    if (!visible.length) {
      $('run-list').replaceChildren(element('div', 'empty', !hasData ? loadError : runs.length ? '조건에 맞는 실행이 없습니다. 검색어나 평가 필터를 변경해 주세요.' : `아직 저장된 실행이 없습니다. runs/v1에 실행을 추가하고 ${generateGuidance}`));
      return;
    }
    const existing = new Map([...$('run-list').children].map(card => [card.dataset.focusKey, card]));
    const cards = visible.map((run, index) => {
      const key = `run:${run.id}`;
      const card = existing.get(key) ?? element('a', 'run-card');
      card.dataset.focusKey = key;
      card.href = routeUrl(model, run.id);
      card.setAttribute('aria-label', `${run.model ?? '모델 미확인'}, ${run.effort ?? 'effort 미확인'}, ${date(run.createdAt)}, ${scoreDescription(run)}, 상세 보기`);
      const info = element('div');
      const title = element('div', 'run-title', run.model ?? '모델 미확인');
      title.append(element('span', 'effort', run.effort ?? 'effort 미확인'));
      const meta = element('div', 'run-meta', `${run.agent ?? '호스트 미확인'} · ${date(run.createdAt)} KST · ${duration(run.durationMs)}`);
      meta.append(element('span', 'run-id', run.id));
      info.append(title, meta);
      const status = element('div', 'run-status');
      status.append(element('span', `status-pill ${Object.hasOwn(statuses, run.status) ? run.status : 'unknown'}`, statuses[run.status] ?? run.status));
      status.append(element('span', 'run-review', run.issues.length ? '기록 확인 필요' : run.artifactUrl ? '결과물 보기 가능' : '결과물 없음'));
      card.replaceChildren(element('span', 'run-number', String(index + 1).padStart(2, '0')), info, status, scoreBlock(run), element('span', 'open-arrow', '↗'));
      return card;
    });
    reconcileChildren($('run-list'), cards);
  }

  function metrics(run) {
    const list = element('dl', 'detail-metrics');
    for (const [label, value] of [
      ['소요 시간', duration(run.durationMs)], ['입력 토큰', run.inputTokens === null ? '미기록' : number.format(run.inputTokens)],
      ['출력 토큰', run.outputTokens === null ? '미기록' : number.format(run.outputTokens)], ['실행 상태', statuses[run.status] ?? run.status],
    ]) {
      const item = element('div', 'metric');
      item.append(element('dt', '', label), element('dd', '', value));
      list.append(item);
    }
    return list;
  }

  function preview(run, panel) {
    if (!run.artifactUrl) {
      panel.append(element('div', 'empty', `이 스냅샷에는 output/index.html 결과물이 없습니다. 결과물을 저장한 뒤 ${generateGuidance}`));
      return;
    }
    const toolbar = element('div', 'preview-toolbar');
    const open = element('a', 'action-link', '결과물 새 탭 열기 ↗');
    open.href = run.artifactUrl;
    open.target = '_blank';
    open.rel = 'noopener noreferrer';
    open.dataset.focusKey = 'open-artifact';
    toolbar.append(element('span', '', 'output / index.html'), open);
    const frame = element('iframe', 'preview-frame');
    frame.title = `${run.model ?? '모델 미확인'} / ${run.effort ?? '미확인'} 생성 결과물`;
    frame.setAttribute('sandbox', 'allow-scripts allow-same-origin allow-pointer-lock');
    frame.referrerPolicy = 'no-referrer';
    frame.src = run.artifactUrl;
    panel.append(toolbar, frame, element('p', 'score-note', '신뢰한 저장소의 결과물을 원본 권한으로 실행합니다. 결과물은 갤러리와 같은 출처의 페이지와 저장소에 접근할 수 있습니다. 신뢰할 수 없는 HTML을 추가하지 마세요. 외부 CDN을 사용하는 결과물은 인터넷 연결이 필요합니다. 미리보기 표시는 평가 검증을 의미하지 않습니다.'));
  }

  function reviewPanel(run, panel) {
    const review = run.review;
    if (!review) {
      panel.append(element('div', 'empty', '저장된 평가가 없습니다. 이 화면은 점수를 새로 계산하거나 평가를 실행하지 않습니다.'));
      return;
    }
    const reviewer = review.reviewer && typeof review.reviewer === 'object' ? review.reviewer : {};
    panel.append(element('p', 'review-summary', `${reviewLabel(run)} · 평가 모델: ${reviewer.model ?? '미확인'} · effort: ${reviewer.effort ?? '미확인'}\n평가 기준: ${review.rubric_version ?? '미기록'} · ${date(review.reviewed_at)} KST`));
    const checks = Array.isArray(review.checks) ? review.checks.filter(value => value && typeof value === 'object') : [];
    const pending = checks.filter(check => check.status === 'unverified').map(check => check.id ?? 'ID 미기록');
    const missing = [['평가 모델', reviewer.model], ['effort', reviewer.effort], ['평가 시각', review.reviewed_at]]
      .filter(([, value]) => typeof value !== 'string' || !value.trim()).map(([label]) => label);
    if (pending.length) panel.append(element('p', 'notice', `미검증 ${pending.length}개 (${pending.join(', ')}) · 숫자 점수가 있어도 이 항목들의 검증은 완료되지 않았습니다.`));
    if (missing.length) panel.append(element('p', 'notice', `평가 정보 누락 (${missing.join(', ')}) · 점수 확정 여부와 별개로 평가 정보가 불완전합니다.`));
    if (score(run) === null) {
      panel.append(element('p', 'notice', '유효한 저장 점수가 없어 미채점으로 표시합니다. 미채점은 0점이 아닙니다. 이 화면은 점수를 새로 계산하지 않습니다.'));
    }
    const raw = scoreRange(review, 'raw_total_score', 'raw_total_score_upper_bound');
    if (raw) panel.append(element('p', 'review-summary', `원점수 ${rangeText(raw)} / 100${raw.lower !== raw.upper ? ' · 가능 범위' : ''}`));
    if (Number.isFinite(review.score_cap)) {
      panel.append(element('p', 'notice', `확인된 필수 항목 실패로 확정 상한 ${number.format(review.score_cap)}점이 적용됩니다. 기존 필수 항목 정책에 따른 상한이며, 모델 품질에 대한 추가 추정이 아닙니다.`));
    }
    if (review.format_version === '4.0') {
      panel.append(element('p', 'score-note', '확정 점수는 가능한 총점이 하나라는 뜻이며 검증 완료를 뜻하지 않습니다. 보수적 하한은 미검증 항목을 포함한 가능 범위의 최솟값이며, 미검증을 실패나 실제 0점으로 판정한 것이 아닙니다. 필수 항목이 미검증이면 실패 가능성을 반영해 하한만 69점 이하로 제한할 수 있으며, 이는 확정 상한이나 실패 판정이 아닙니다.'));
    }
    if (review.verification) {
      panel.append(element('p', 'review-summary', `브라우저 검증: ${review.verification.browser_method ?? '미기록'} · 요구사항 판정: ${review.requirements_verdict ?? '미기록'}`));
    }
    const categories = Array.isArray(review.categories) ? review.categories.filter(value => value && typeof value === 'object') : [];
    const grid = element('div', 'category-grid');
    for (const category of categories) {
      const range = scoreRange({ ...category, format_version: review.format_version }, 'score', 'score_upper_bound', 5);
      const name = categoryNames[category.id] ?? category.id ?? '평가 항목';
      const card = element('div', 'category');
      card.append(element('h3', '', name));
      const value = element('p', '', range ? `${rangeText(range)}${range.lower !== range.upper ? ' · 가능 범위' : ''} ` : '미채점 ');
      value.append(element('small', '', `/ 5 · 가중치 ${category.weight ?? '미기록'}`));
      card.append(value);
      if (range) {
        const bar = element('progress');
        bar.max = 5;
        bar.value = range.lower;
        bar.setAttribute('aria-label', `${name} ${rangeText(range)} / 5${range.lower !== range.upper ? ' · 가능 범위, 막대는 보수적 하한' : ''}`);
        card.append(bar);
      }
      grid.append(card);
    }
    panel.append(grid);
    if (checks.length) panel.append(element('h2', '', '검증 항목'));
    for (const check of checks) {
      const row = element('div', 'check');
      const heading = element('div', 'check-heading');
      heading.append(element('strong', '', check.id ?? '항목'), element('span', `check-status ${Object.hasOwn(checkStatuses, check.status) ? check.status : ''}`, checkStatuses[check.status] ?? check.status ?? '미기록'));
      if (check.method) heading.append(element('span', 'run-review', check.method));
      row.append(heading, element('p', '', check.evidence ?? '근거 미기록'));
      panel.append(row);
    }
    if (Array.isArray(review.notes) && review.notes.length) {
      const notes = element('ul', 'record-notes');
      notes.append(...review.notes.filter(note => typeof note === 'string').map(note => element('li', '', note)));
      panel.append(element('h2', '', '평가 참고 사항'), notes);
    }
    if (run.reviewMarkdown) {
      const details = element('details');
      details.dataset.focusKey = 'review-original';
      details.append(element('summary', '', '평가 원문 보기'), element('pre', 'text-panel', run.reviewMarkdown));
      panel.append(details);
    }
  }

  function renderDetail(run, selectedModel) {
    if (detailView?.run === run) {
      detailView.back.href = routeUrl(selectedModel);
      return;
    }
    const back = element('a', 'back-button', '← 실행 목록으로');
    back.dataset.focusKey = 'back-to-runs';
    const heading = element('div', 'page-heading detail-heading');
    const measurements = element('div');
    const notice = element('div', 'notice');
    const tabs = element('div', 'detail-tabs');
    tabs.setAttribute('role', 'group');
    tabs.setAttribute('aria-label', '상세 내용 선택');
    const panel = element('div');
    panel.id = 'detail-panel';
    detailView = { run, back, heading, measurements, notice, tabs, panel, section: activeSections.get(run.id) ?? 'preview' };
    for (const [key, label] of sections) {
      const button = element('button', '', label);
      button.type = 'button';
      button.dataset.section = key;
      button.dataset.focusKey = `section:${key}`;
      button.setAttribute('aria-controls', 'detail-panel');
      button.addEventListener('click', () => {
        if (detailView.section === key) return;
        detailView.section = key;
        activeSections.set(run.id, key);
        renderSection();
      });
      tabs.append(button);
    }
    $('detail').replaceChildren(back, heading, measurements, notice, tabs, panel);
    const view = detailView;
    view.back.href = routeUrl(selectedModel);
    const titleGroup = element('div');
    const title = element('h1', '', run.model ?? '모델 미확인');
    title.id = 'detail-title';
    titleGroup.append(element('p', 'eyebrow', 'RUN DETAIL / SOLAR SYSTEM'), title,
      element('p', 'description', `${run.agent ?? '호스트 미확인'} · ${run.effort ?? 'effort 미확인'} · ${date(run.createdAt)} KST`),
      element('p', 'run-meta', run.id));
    view.heading.replaceChildren(titleGroup, scoreBlock(run, 'detail-score'));
    view.measurements.replaceChildren(metrics(run));
    view.notice.hidden = !run.issues.length;
    view.notice.textContent = run.issues.length ? `일부 기록을 읽지 못했습니다.\n${run.issues.join('\n')}` : '';
    renderSection();
  }

  function renderSection() {
    const { run, panel, tabs, section } = detailView;
    for (const button of tabs.children) button.setAttribute('aria-pressed', String(button.dataset.section === section));
    panel.replaceChildren();
    if (section === 'preview') preview(run, panel);
    if (section === 'review') reviewPanel(run, panel);
    if (section === 'prompt') panel.append(element('pre', 'text-panel', run.prompt ?? '저장된 prompt.md가 없습니다.'));
    if (section === 'record') {
      panel.append(element('h2', '', '설정·측정 출처와 실행 참고 사항'));
      if (run.notes.length) {
        const notes = element('ul', 'record-notes');
        notes.append(...run.notes.map(note => element('li', '', note)));
        panel.append(notes);
      } else panel.append(element('p', 'review-summary', '저장된 참고 사항이 없습니다.'));
      panel.append(element('pre', 'text-panel', JSON.stringify({
        id: run.id, model: run.model, effort: run.effort, agent: run.agent, status: run.status,
        created_at: run.createdAt, duration_ms: run.durationMs,
        input_tokens: run.inputTokens, output_tokens: run.outputTokens,
      }, null, 2)));
    }
  }

  function render() {
    const selected = route();
    renderModels(selected.model);
    const run = runs.find(item => item.id === selected.run);
    $('collection').hidden = Boolean(selected.run);
    $('detail').hidden = !selected.run;
    if (run) {
      document.title = `${run.model ?? '모델 미확인'} · ${run.effort ?? '미확인'} | AI Benchmark`;
      renderDetail(run, selected.model);
    } else if (selected.run) {
      detailView = null;
      const back = element('a', 'action-link', '실행 목록으로');
      back.href = routeUrl(selected.model);
      back.dataset.focusKey = 'back-to-runs';
      const title = element('h1', '', hasData ? '실행을 찾을 수 없습니다' : '실행 기록을 불러올 수 없습니다');
      title.id = 'detail-title';
      $('detail').replaceChildren(title, element('p', 'description', hasData ? `선택한 실행이 이 스냅샷에 없습니다. 저장된 실행을 추가하거나 변경했다면 ${generateGuidance}` : loadError), back);
      document.title = '실행 없음 | AI Benchmark';
    } else {
      detailView = null;
      $('detail').replaceChildren();
      document.title = 'AI Benchmark · Run Explorer';
      renderList();
    }
  }

  function focusMain() {
    $('main').focus({ preventScroll: true });
    $('main').scrollIntoView({ block: 'start' });
  }

  document.querySelector('.skip-link').addEventListener('click', event => {
    event.preventDefault();
    focusMain();
  });
  $('search').addEventListener('input', renderList);
  $('review-filter').addEventListener('change', renderList);
  $('sort').addEventListener('change', renderList);
  window.addEventListener('hashchange', () => {
    render();
    focusMain();
  });
  render();
  $('connection').dataset.state = hasData ? 'snapshot' : 'error';
  $('connection-state').textContent = hasData ? '정적 스냅샷 · SNAPSHOT' : '데이터 없음 · NO DATA';
  $('connection-message').textContent = hasData
    ? '저장된 스냅샷을 표시합니다. 변경된 기록은 데이터 재생성 후 새로고침하면 반영됩니다.'
    : loadError;
  $('snapshot-revision').textContent = hasData ? `스냅샷 ${snapshot.revision.slice(0, 12)}` : '스냅샷 데이터 없음';
})();
