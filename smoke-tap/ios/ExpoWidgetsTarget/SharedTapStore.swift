import Foundation

struct SharedTapStore {
    static let appGroupId = "group.com.wonjinsin.smoketap"
    static let pendingKey = "pendingTaps"
    static let baseKey   = "baseTodayCount"
    static let baseDateKey = "baseDate"
    static let lastTapKey = "lastTapTimestamp"

    static func dayString(_ date: Date) -> String {
        let f = DateFormatter()
        f.dateFormat = "yyyy-MM-dd"
        return f.string(from: date)
    }
    static func todayString() -> String { dayString(Date()) }
    // Pending taps are stored as an array of tap timestamps (epoch seconds), not a
    // bare counter, so each tap keeps its real day. That lets the day-scoped read
    // below reset to 0 at midnight and lets the app replay taps on their true day.
    static func recordTap() {
        guard let d = UserDefaults(suiteName: appGroupId) else { return }
        var arr = d.array(forKey: pendingKey) as? [Double] ?? []
        arr.append(Date().timeIntervalSince1970)
        d.set(arr, forKey: pendingKey)
    }
    // Only pending taps made today count toward today's display; leftover taps
    // from a previous day are ignored here (the app reattributes them on sync).
    static func getPendingCount() -> Int {
        guard let d = UserDefaults(suiteName: appGroupId) else { return 0 }
        let arr = d.array(forKey: pendingKey) as? [Double] ?? []
        let today = todayString()
        return arr.filter { dayString(Date(timeIntervalSince1970: $0)) == today }.count
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
