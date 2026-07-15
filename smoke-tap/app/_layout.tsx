import '../global.css';
import { useEffect, useRef } from 'react';
import { AppState } from 'react-native';
import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { useTapStore } from '../store/useTapStore';
import { getPendingTaps, clearPending, setBaseCount } from '../modules/SharedTapStore';
import { C } from '../constants/colors';

function toLocalDateString(ts: number): string {
  const d = new Date(ts);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

function getTodayCount(records: { timestamp: number }[]): number {
  const today = toLocalDateString(Date.now());
  return records.filter((r) => toLocalDateString(r.timestamp) === today).length;
}

function useWidgetSync() {
  // 위젯 → 앱: pending 탭을 원본 시각 그대로 앱 store에 반영
  async function syncPending() {
    const pending = await getPendingTaps();
    if (pending.length > 0) {
      useTapStore.getState().importWidgetTaps(pending);
      await clearPending();
    }
  }

  // 앱 → 위젯: 오늘 카운트를 App Groups에 기록
  async function syncBaseCount() {
    const count = getTodayCount(useTapStore.getState().records);
    await setBaseCount(count);
  }

  useEffect(() => {
    const runSync = () => syncPending().then(syncBaseCount);

    // Wait for persist rehydration before the first sync: importWidgetTaps must
    // not append onto the pre-hydration empty records, or the async rehydrate
    // merge would overwrite (and lose) the just-imported taps.
    let unsubHydrate: (() => void) | undefined;
    if (useTapStore.persist.hasHydrated()) runSync();
    else unsubHydrate = useTapStore.persist.onFinishHydration(runSync);

    // 앱이 포그라운드로 돌아올 때마다 동기화 (이 시점엔 이미 하이드레이션 완료)
    const sub = AppState.addEventListener('change', (state) => {
      if (state === 'active') runSync();
    });
    return () => {
      sub.remove();
      unsubHydrate?.();
    };
  }, []);

  // store records가 바뀔 때마다 위젯 기준값 갱신
  useEffect(() => {
    return useTapStore.subscribe((state) => {
      setBaseCount(getTodayCount(state.records));
    });
  }, []);
}

export default function RootLayout() {
  useWidgetSync();

  return (
    <>
      <StatusBar style="dark" />
      <Stack screenOptions={{ contentStyle: { backgroundColor: C.BG } }}>
        <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
      </Stack>
    </>
  );
}
