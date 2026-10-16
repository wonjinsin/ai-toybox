(() => {
  'use strict';

  const data = JSON.parse(document.getElementById('gallery-data').textContent);
  const runs = data.runs;
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
  const reviewStatuses = { completed: '평가 완료', partial: '부분 평가', blocked: '평가 불가' };
  const checkStatuses = { pass: '통과', fail: '실패', unverified: '미검증' };
  const models = [...new Set(runs.map(run => run.model ?? ''))].sort();

  function element(tag, className, text) {
    const node = document.createElement(tag);
    if (className) node.className = className;
    if (text !== undefined && text !== null) node.textContent = text;
    return node;
  }

  function score(run) {
    const value = run.review?.total_score;
    return Number.isFinite(value) && value >= 0 && value <= 100 ? value : null;
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
    const value = score(run);
    wrapper.append(element('span', `score${value === null ? ' unscored' : ''}`, value === null ? '미채점' : number.format(value)));
    wrapper.append(element('span', 'score-caption', value === null ? reviewLabel(run) : `/ 100 · ${reviewLabel(run)}`));
    return wrapper;
  }

  function renderModels(selected) {
    $('model-count').textContent = String(models.length).padStart(2, '0');
    const buttons = [null, ...models].map(model => {
      const button = element('button', 'model-button');
      button.type = 'button';
      button.setAttribute('aria-pressed', String(model === selected));
      button.append(element('span', 'model-symbol', model === null ? '▦' : '◈'));
      const name = element('span', 'model-name', model === null ? '모든 모델' : model || '모델 미확인');
      if (model?.includes('/')) {
        name.textContent = model.split('/').at(-1);
        name.append(element('small', '', model.slice(0, model.lastIndexOf('/'))));
      }
      button.title = model === null ? '모든 모델' : model || '모델 미확인';
      button.append(name, element('span', 'model-count', model === null ? runs.length : runs.filter(run => (run.model ?? '') === model).length));
      button.addEventListener('click', () => {
        $('search').value = '';
        $('review-filter').value = 'all';
        const url = routeUrl(model);
        if (location.hash === url || (!location.hash && url === '#')) render();
        else location.hash = url;
      });
      return button;
    });
    $('models').replaceChildren(...buttons);
  }

  function renderStats(selectedRuns) {
    const scored = selectedRuns.filter(run => score(run) !== null);
    const scores = scored.map(score);
    const stats = [
      ['실행 기록', selectedRuns.length, '저장된 시도'],
      ['생성된 결과물', selectedRuns.filter(run => run.artifactUrl).length, 'HTML 결과'],
      ['최고 점수', scores.length ? number.format(Math.max(...scores)) : '—', `${scored.length}개 채점 / 100`],
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
      $('run-list').replaceChildren(element('div', 'empty', runs.length ? '조건에 맞는 실행이 없습니다. 검색어나 평가 필터를 변경해 주세요.' : '아직 저장된 실행이 없습니다. runs/v1에 실행을 추가한 뒤 갤러리를 다시 생성해 주세요.'));
      return;
    }
    $('run-list').replaceChildren(...visible.map((run, index) => {
      const card = element('a', 'run-card');
      card.href = routeUrl(model, run.id);
      card.setAttribute('aria-label', `${run.model ?? '모델 미확인'}, ${run.effort ?? 'effort 미확인'}, ${date(run.createdAt)}, ${score(run) === null ? '미채점' : `${score(run)}점`}, 상세 보기`);
      const info = element('div');
      const title = element('div', 'run-title', run.model ?? '모델 미확인');
      title.append(element('span', 'effort', run.effort ?? 'effort 미확인'));
      const meta = element('div', 'run-meta', `${run.agent ?? '호스트 미확인'} · ${date(run.createdAt)} KST · ${duration(run.durationMs)}`);
      meta.append(element('span', 'run-id', run.id));
      info.append(title, meta);
      const status = element('div', 'run-status');
      status.append(element('span', `status-pill ${Object.hasOwn(statuses, run.status) ? run.status : 'unknown'}`, statuses[run.status] ?? run.status));
      status.append(element('span', 'run-review', run.issues.length ? '기록 확인 필요' : run.artifactUrl ? '결과물 보기 가능' : '결과물 없음'));
      card.append(element('span', 'run-number', String(index + 1).padStart(2, '0')), info, status, scoreBlock(run), element('span', 'open-arrow', '↗'));
      return card;
    }));
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
      panel.append(element('div', 'empty', '아직 output/index.html 결과물이 없습니다. 생성이 끝난 뒤 갤러리를 다시 생성해 주세요.'));
      return;
    }
    const toolbar = element('div', 'preview-toolbar');
    const open = element('a', 'action-link', '결과물 새 탭 열기 ↗');
    open.href = run.artifactUrl;
    open.target = '_blank';
    open.rel = 'noopener noreferrer';
    toolbar.append(element('span', '', 'output / index.html'), open);
    const frame = element('iframe', 'preview-frame');
    frame.title = `${run.model ?? '모델 미확인'} / ${run.effort ?? '미확인'} 생성 결과물`;
    frame.setAttribute('sandbox', 'allow-scripts allow-pointer-lock');
    frame.referrerPolicy = 'no-referrer';
    frame.src = run.artifactUrl;
    panel.append(toolbar, frame, element('p', 'score-note', '결과물은 격리된 미리보기에서 실행됩니다. 비어 있거나 일부 기능이 제한되면 새 탭에서 여세요. 외부 CDN을 사용하는 결과물은 인터넷 연결이 필요합니다. 미리보기 표시는 평가 검증을 의미하지 않습니다.'));
  }

  function reviewPanel(run, panel) {
    const review = run.review;
    if (!review) {
      panel.append(element('div', 'empty', '저장된 평가가 없습니다. 이 화면은 점수를 새로 계산하거나 평가를 실행하지 않습니다.'));
      return;
    }
    const reviewer = review.reviewer && typeof review.reviewer === 'object' ? review.reviewer : {};
    panel.append(element('p', 'review-summary', `${reviewLabel(run)} · 평가 모델: ${reviewer.model ?? '미확인'} · effort: ${reviewer.effort ?? '미확인'}\n평가 기준: ${review.rubric_version ?? '미기록'} · ${date(review.reviewed_at)} KST`));
    if (score(run) === null) panel.append(element('p', 'notice', '총점이 기록되지 않았습니다. 미검증 항목이나 평가 제한을 확인하세요. 미채점 항목은 0점으로 환산하지 않습니다.'));
    if (review.verification) {
      panel.append(element('p', 'review-summary', `브라우저 검증: ${review.verification.browser_method ?? '미기록'} · 직접 파일 실행: ${review.verification.direct_file ?? '미기록'} · 요구사항 판정: ${review.requirements_verdict ?? '미기록'}`));
    }
    const categories = Array.isArray(review.categories) ? review.categories.filter(value => value && typeof value === 'object') : [];
    const grid = element('div', 'category-grid');
    for (const category of categories) {
      const valid = Number.isFinite(category.score) && category.score >= 0 && category.score <= 5;
      const card = element('div', 'category');
      card.append(element('h3', '', categoryNames[category.id] ?? category.id ?? '평가 항목'));
      const value = element('p', '', valid ? `${category.score} ` : '미채점 ');
      value.append(element('small', '', `/ 5 · 가중치 ${category.weight ?? '미기록'}`));
      card.append(value);
      if (valid) {
        const bar = element('progress');
        bar.max = 5;
        bar.value = category.score;
        bar.setAttribute('aria-label', `${categoryNames[category.id] ?? category.id} ${category.score}/5`);
        card.append(bar);
      }
      grid.append(card);
    }
    panel.append(grid);
    const checks = Array.isArray(review.checks) ? review.checks.filter(value => value && typeof value === 'object') : [];
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
      details.append(element('summary', '', '평가 원문 보기'), element('pre', 'text-panel', run.reviewMarkdown));
      panel.append(details);
    }
  }

  function renderDetail(run, selectedModel) {
    const detail = $('detail');
    const back = element('a', 'back-button', '← 실행 목록으로');
    back.href = routeUrl(selectedModel);
    const heading = element('div', 'page-heading detail-heading');
    const titleGroup = element('div');
    const title = element('h1', '', run.model ?? '모델 미확인');
    title.id = 'detail-title';
    const subtitle = element('p', 'description', `${run.agent ?? '호스트 미확인'} · ${run.effort ?? 'effort 미확인'} · ${date(run.createdAt)} KST`);
    titleGroup.append(element('p', 'eyebrow', 'RUN DETAIL / SOLAR SYSTEM'), title, subtitle, element('p', 'run-meta', run.id));
    heading.append(titleGroup, scoreBlock(run, 'detail-score'));
    const tabs = element('div', 'detail-tabs');
    tabs.setAttribute('role', 'group');
    tabs.setAttribute('aria-label', '상세 내용 선택');
    const panel = element('div');
    panel.id = 'detail-panel';
    detail.replaceChildren(back, heading, metrics(run));
    if (run.issues.length) detail.append(element('div', 'notice', `일부 기록을 읽지 못했습니다.\n${run.issues.join('\n')}`));
    detail.append(tabs, panel);
    const sections = [['preview', '결과물'], ['review', '점수 · 평가'], ['prompt', '사용한 프롬프트'], ['record', '실행 기록']];
    function showSection(key) {
      for (const button of tabs.children) button.setAttribute('aria-pressed', String(button.dataset.section === key));
      panel.replaceChildren();
      if (key === 'preview') preview(run, panel);
      if (key === 'review') reviewPanel(run, panel);
      if (key === 'prompt') panel.append(element('pre', 'text-panel', run.prompt ?? '저장된 prompt.md가 없습니다.'));
      if (key === 'record') {
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
    for (const [key, label] of sections) {
      const button = element('button', '', label);
      button.type = 'button';
      button.dataset.section = key;
      button.setAttribute('aria-controls', 'detail-panel');
      button.addEventListener('click', () => showSection(key));
      tabs.append(button);
    }
    showSection('preview');
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
      const back = element('a', 'action-link', '실행 목록으로');
      back.href = routeUrl(selected.model);
      const title = element('h1', '', '실행을 찾을 수 없습니다');
      title.id = 'detail-title';
      $('detail').replaceChildren(title, element('p', 'description', '실행 ID를 확인하거나 갤러리를 다시 생성해 주세요.'), back);
      document.title = '실행 없음 | AI Benchmark';
    } else {
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
  $('generated-at').textContent = `마지막 수집 ${date(data.generatedAt)} KST · 갱신: node scripts/build-gallery.mjs`;
  render();
})();
