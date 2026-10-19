import Foundation
#if canImport(UIKit)
import UIKit
#endif

public final class DeepLinkManager {
    public static func openURL(_ string: String) {
        guard let url = URL(string: string) else { return }
        #if canImport(UIKit)
        UIApplication.shared.open(url)
        #endif
    }

    public static func initiateLinkFlow(appId: String, sessionCode: String) {
        let scheme: String
        switch appId.lowercased() {
        case "ftracker": scheme = "ftracker://"
        case "taskmanagement", "taskhub": scheme = "taskhub://"
        default: scheme = "app://"
        }
        let urlString = "\(scheme)connect-hub?session_code=\(sessionCode)&callback_url=ekahub://link-callback"
        openURL(urlString)
    }
}
