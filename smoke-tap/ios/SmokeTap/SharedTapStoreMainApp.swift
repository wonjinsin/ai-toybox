import Foundation
import WidgetKit

struct SharedTapStoreMainApp {
    static let appGroupId = "group.com.wonjinsin.smoketap"
    static let pendingKey = "pendingTaps"
    static let baseKey   = "baseTodayCount"
    static let baseDateKey = "baseDate"
    static let lastTapKey = "lastTapTimestamp"

    static func getPendingCount() -> Int {
        UserDefaults(suiteName: appGroupId)?.integer(forKey: pendingKey) ?? 0
    }
    static func clearPending() {
        UserDefaults(suiteName: appGroupId)?.set(0, forKey: pendingKey)
    }
    // Stamp the base count with today's date so the widget can reset it once
    // the day rolls over while the app is closed. Format must match the widget's
    // SharedTapStore.todayString() ("yyyy-MM-dd").
    static func setBaseCount(_ count: Int) {
        let f = DateFormatter()
        f.dateFormat = "yyyy-MM-dd"
        let d = UserDefaults(suiteName: appGroupId)
        d?.set(count, forKey: baseKey)
        d?.set(f.string(from: Date()), forKey: baseDateKey)
        WidgetCenter.shared.reloadTimelines(ofKind: "SmokeTapWidget")
    }
    static func setLastTap(_ ts: Double) {
        UserDefaults(suiteName: appGroupId)?.set(ts, forKey: lastTapKey)
        WidgetCenter.shared.reloadTimelines(ofKind: "SmokeTapWidget")
    }
}
