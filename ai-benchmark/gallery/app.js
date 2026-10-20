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
    if (!['2.0', '4.0', '5.0'].includes(review?.format_version)) return null;
    const lower = review?.[lowerKey];
    const upper = review?.format_version === '2.0' ? lower : review[upperKey];
    return Number.isFinite(lower) && Number.isFinite(upper) && lower >= 0 && lower <= upper && upper <= maximum
      ? { lower, upper } : null;
  }

  function quality(run) {
    return ['current', 'policy-only'].includes(run.policyProjection?.kind) ? run.policyProjection.quality : null;
  }

  function headlineRange(run) {
    return quality(run) ?? (run.policyProjection?.kind === 'historical' ? scoreRange(run.policyProjection.assessment) : null);
  }

  function score(run) {
    return headlineRange(run)?.lower ?? null;
  }

  function scoreGroup(run) {
    return quality(run) ? 0 : run.policyProjection?.kind === 'historical' ? 1 : 2;
  }

  function policyLabel(run) {
    if (run.policyProjection?.kind === 'policy-only') return '품질 · 새 정책 5.0 · 기존 관찰 근거';
    if (run.policyProjection?.kind === 'current') return '품질 · 현행 정책 5.0';
    if (run.policyProjection?.kind === 'historical') return '과거 과제 2.0 · 별도 점수';
    return run.review ? '정책 적용 불가 · 기록 확인 필요' : '평가 없음';
  }

  function conformanceText(run) {
    const assessment = quality(run) ?? run.policyProjection?.assessment;
    if (!assessment) return '요구사항 판정·검증 범위 확인 불가';
    const coverage = assessment.coverage;
    const mandatory = quality(run);
    return `요구사항 ${checkStatuses[assessment.requirements_verdict]} · 실패 ${coverage.failed}개 · 미검증 ${coverage.unverified}개${mandatory ? ` · 필수 실패 ${mandatory.mandatory_failed.length}개 · 필수 미검증 ${mandatory.mandatory_unverified.length}개` : ''}`;
  }

  function rangeText(range) {
    return range.lower === range.upper ? number.format(range.lower) : `${number.format(range.lower)} ~ ${number.format(range.upper)}`;
  }

  function scoreDescription(run) {
    const range = headlineRange(run);
    const headline = range ? `${number.format(range.lower)}점 / 100${range.lower !== range.upper ? ` · 가능 범위 ${rangeText(range)}` : ''}` : '미채점';
    return `${headline} · ${policyLabel(run)} · ${conformanceText(run)} · ${reviewLabel(run)}`;
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
    const range = headlineRange(run);
    wrapper.setAttribute('role', 'group');
    wrapper.setAttribute('aria-label', scoreDescription(run));
    wrapper.append(element('span', `score${range === null ? ' unscored' : ''}`, range === null ? '미채점' : number.format(range.lower)));
    wrapper.append(element('span', 'score-caption', policyLabel(run)));
    wrapper.append(element('span', 'score-caption', range === null ? reviewLabel(run) : `/ 100 · ${range.lower === range.upper ? '확정값' : '보수적 하한'} · ${reviewLabel(run)}`));
    if (range && range.lower !== range.upper) wrapper.append(element('span', 'score-caption', `가능 범위 ${rangeText(range)}`));
    const conformance = element('span', 'score-conformance', conformanceText(run));
    conformance.dataset.verdict = quality(run)?.requirements_verdict ?? run.policyProjection?.assessment?.requirements_verdict ?? 'unverified';
    wrapper.append(conformance);
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
    const scored = selectedRuns.filter(run => quality(run));
    const scores = scored.map(run => quality(run).lower);
    const stats = [
      ['실행 기록', selectedRuns.length, '저장된 시도'],
      ['생성된 결과물', selectedRuns.filter(run => run.artifactUrl).length, 'HTML 결과'],
      ['최고 현행 품질 점수', scores.length ? number.format(Math.max(...scores)) : '—', `${scored.length}개 현행 정책 점수 · 확정값 또는 보수적 하한 / 100 · 과거 과제 제외 · 요구사항 통과와 별개`],
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
      visible.sort((a, b) => scoreGroup(a) - scoreGroup(b) || (score(b) ?? -1) - (score(a) ?? -1));
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
    if ($('sort').value === 'score') {
      const grouped = [];
      let previousGroup = null;
      const labels = ['현행 품질 순위 · solar-system-5.0', '과거 과제 점수 · 별도 순위 · solar-system-2.0', '정책 점수 없음 · 순위 제외'];
      let rank = 0;
      visible.forEach((run, index) => {
        const group = scoreGroup(run);
        if (group !== previousGroup) {
          grouped.push(element('h3', 'score-group', labels[group]));
          previousGroup = group;
          rank = 0;
        }
        cards[index].firstChild.textContent = score(run) === null ? '·' : String(++rank).padStart(2, '0');
        grouped.push(cards[index]);
      });
      reconcileChildren($('run-list'), grouped);
    } else reconcileChildren($('run-list'), cards);
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
      panel.append(element('div', 'empty', '저장된 평가가 없습니다. 이 화면은 새 관찰이나 브라우저 검증을 수행하지 않습니다.'));
      return;
    }
    const reviewer = review.reviewer && typeof review.reviewer === 'object' ? review.reviewer : {};
    panel.append(element('p', 'review-summary', `${reviewLabel(run)} · 평가 모델: ${reviewer.model ?? '미확인'} · effort: ${reviewer.effort ?? '미확인'}\n평가 기준: ${review.rubric_version ?? '미기록'} · ${date(review.reviewed_at)} KST`));
    panel.append(element('p', 'review-summary', `${policyLabel(run)}\n${conformanceText(run)}`));
    const projection = run.policyProjection;
    if (projection?.kind === 'policy-only') {
      panel.append(element('p', 'notice', `${projection.source_rubric} → ${projection.target_rubric} 정책 비교입니다. 기존 관찰만 사용하며 새 브라우저 검증이나 결과물 개선을 의미하지 않습니다. 필수 실패는 품질 점수를 제한하지 않지만 요구사항 실패 판정은 유지됩니다.`));
    } else if (projection?.kind === 'historical') {
      panel.append(element('p', 'notice', '과거 과제는 요구사항과 검증 항목이 달라 현행 정책으로 환산하지 않습니다. 점수 정렬에서도 별도 순위로 표시합니다.'));
    } else if (projection?.kind === 'current') {
      panel.append(element('p', 'score-note', '현행 품질 점수는 필수 실패 상한 없이 계산합니다. 요구사항 판정과 실패·미검증 개수는 점수와 독립적입니다.'));
    }
    if (run.sourceReviewSha256) panel.append(element('p', 'source-hash', `근거 review.json 원본 바이트 SHA-256 · ${run.sourceReviewSha256}`));
    const published = scoreRange(review);
    panel.append(element('p', 'review-summary', `원본 공개 점수 · ${review.rubric_version ?? '기준 미기록'} · ${published ? `${rangeText(published)} / 100` : '미채점'}${projection ? '' : ' · 원본 수치만 표시, 검증되지 않은 점수이며 순위에서 제외'}`));
    const checks = Array.isArray(review.checks) ? review.checks.filter(value => value && typeof value === 'object') : [];
    const pending = checks.filter(check => check.status === 'unverified').map(check => check.id ?? 'ID 미기록');
    const missing = [['평가 모델', reviewer.model], ['effort', reviewer.effort], ['평가 시각', review.reviewed_at]]
      .filter(([, value]) => typeof value !== 'string' || !value.trim()).map(([label]) => label);
    if (pending.length) panel.append(element('p', 'notice', `미검증 ${pending.length}개 (${pending.join(', ')}) · 숫자 점수가 있어도 이 항목들의 검증은 완료되지 않았습니다.`));
    if (missing.length) panel.append(element('p', 'notice', `평가 정보 누락 (${missing.join(', ')}) · 점수 확정 여부와 별개로 평가 정보가 불완전합니다.`));
    if (score(run) === null) {
      panel.append(element('p', 'notice', '유효한 정책 점수가 없어 미채점으로 표시합니다. 원본의 숫자를 품질 점수로 대신 사용하지 않습니다. 미채점은 0점이 아닙니다.'));
    }
    const raw = scoreRange(review, 'raw_total_score', 'raw_total_score_upper_bound');
    if (raw) panel.append(element('p', 'review-summary', `원본 정책의 상한 적용 전 점수 ${rangeText(raw)} / 100${raw.lower !== raw.upper ? ' · 가능 범위' : ''}`));
    if (Number.isFinite(review.score_cap)) {
      panel.append(element('p', 'notice', `원본 공개 점수에는 필수 항목 실패로 ${number.format(review.score_cap)}점 상한이 적용되었습니다. 이는 과거 4.0 정책이며 현행 품질 점수에는 적용하지 않습니다.`));
    }
    if (review.format_version === '4.0') {
      panel.append(element('p', 'score-note', '원본 4.0 정책은 필수 실패 시 양쪽 끝점을 69점 이하로 제한했습니다. 필수 항목이 미검증이면 하한만 제한할 수 있었습니다. 원본 점수와 근거는 변경하지 않습니다.'));
    }
    if (quality(run)) panel.append(element('p', 'score-note', '확정값은 반올림한 가능한 점수가 같다는 뜻이며 검증 완료를 뜻하지 않습니다. 보수적 하한은 미검증 항목을 포함한 가능 범위의 최솟값입니다. 미검증은 실패나 실제 0점 판정이 아닙니다.'));
    if (review.verification) {
      panel.append(element('p', 'review-summary', `원본 브라우저 검증 기록: ${review.verification.browser_method ?? '미기록'} · 원본 요구사항 판정: ${review.requirements_verdict ?? '미기록'}`));
    }
    const categories = quality(run)?.categories ?? projection?.assessment?.categories ?? [];
    const grid = element('div', 'category-grid');
    for (const category of categories) {
      const range = scoreRange({ ...category, format_version: quality(run) ? '5.0' : review.format_version }, 'score', 'score_upper_bound', 5);
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
