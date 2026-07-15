internal import ExpoModulesCore

class SharedTapStoreModule: Module {
    func definition() -> ModuleDefinition {
        Name("SharedTapStore")
        AsyncFunction("getPendingTaps")  { () -> [Double] in SharedTapStoreMainApp.getPendingTaps() }
        AsyncFunction("clearPending")    { () -> Void in SharedTapStoreMainApp.clearPending() }
        AsyncFunction("setBaseCount")    { (count: Int) -> Void in SharedTapStoreMainApp.setBaseCount(count) }
        AsyncFunction("setLastTap")      { (ts: Double) -> Void in SharedTapStoreMainApp.setLastTap(ts) }
    }
}
