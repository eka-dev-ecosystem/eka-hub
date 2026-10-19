import SwiftUI

#if canImport(UIKit)
import UIKit
final class AppDelegate: NSObject, UIApplicationDelegate {
    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil
    ) -> Bool {
        Task { @MainActor in
            PushNotificationManager.shared.configure()
        }
        return true
    }

    func application(
        _ application: UIApplication,
        didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data
    ) {
        Task { @MainActor in
            PushNotificationManager.shared.handleDeviceTokenData(deviceToken)
        }
    }
}
#endif

@main
struct EkaHubApp: App {
    #if canImport(UIKit)
    @UIApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    #endif

    @State private var authViewModel = AuthViewModel()

    var body: some Scene {
        WindowGroup {
            Group {
                if authViewModel.isAuthenticated {
                    MainTabView()
                } else {
                    AuthView(onAuthSuccess: {
                        authViewModel.isAuthenticated = true
                    })
                }
            }
            .preferredColorScheme(.dark)
            .environment(\.locale, Locale(identifier: I18n.shared.currentLanguage.rawValue))
        }
    }
}
