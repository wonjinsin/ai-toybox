import Foundation

struct SharedTapStore {
    static let appGroupId = "group.com.wonjinsin.smoketap"
    static let pendingKey = "pendingTaps"
    static let baseKey   = "baseTodayCount"
    static let baseDateKey = "baseDate"
    static let lastTapKey = "lastTapTimestamp"

    static func todayString() -> String {
        let f = DateFormatter()
        f.dateFormat = "yyyy-MM-dd"
        return f.string(from: Date())
    }
    static func recordTap() {
        guard let d = UserDefaults(suiteName: appGroupId) else { return }
        d.set(d.integer(forKey: pendingKey) + 1, forKey: pendingKey)
    }
    static func getPendingCount() -> Int {
        UserDefaults(suiteName: appGroupId)?.integer(forKey: pendingKey) ?? 0
    }
    // Base count is stamped with the day it was written. If the day has rolled
    // over while the app is closed, treat it as 0 so the widget shows the new day.
    static func getBaseCount() -> Int {
        guard let d = UserDefaults(suiteName: appGroupId) else { return 0 }
        guard d.string(forKey: baseDateKey) == todayString() else { return 0 }
        return d.integer(forKey: baseKey)
    }
    static func setLastTap(_ ts: Double) {
        UserDefaults(suiteName: appGroupId)?.set(ts, forKey: lastTapKey)
    }
    static func getLastTap() -> Date? {
        let ts = UserDefaults(suiteName: appGroupId)?.double(forKey: lastTapKey) ?? 0
        return ts > 0 ? Date(timeIntervalSince1970: ts) : nil
    }
}
