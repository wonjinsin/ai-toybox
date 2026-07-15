import Foundation
import WidgetKit

struct SharedTapStoreMainApp {
    static let appGroupId = "group.com.wonjinsin.smoketap"
    static let pendingKey = "pendingTaps"
    static let baseKey   = "baseTodayCount"
    static let baseDateKey = "baseDate"
    static let lastTapKey = "lastTapTimestamp"

    // Pending taps are an array of tap timestamps (epoch seconds); return them so
    // the app can replay each on its real day instead of stamping them "now".
    static func getPendingTaps() -> [Double] {
        UserDefaults(suiteName: appGroupId)?.array(forKey: pendingKey) as? [Double] ?? []
    }
    static func clearPending() {
        UserDefaults(suiteName: appGroupId)?.removeObject(forKey: pendingKey)
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
