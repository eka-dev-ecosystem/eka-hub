import SwiftUI
import GoogleSignIn

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
        AppLogger.info("APNs registered", category: .app)
    }

    func application(
        _ application: UIApplication,
        didFailToRegisterForRemoteNotificationsWithError error: Error
    ) {
        AppLogger.warn("APNs registration skipped or failed", category: .app, metadata: ["error": error.localizedDescription])
    }
}
#endif

@main
struct EkaHubApp: App {
    #if canImport(UIKit)
    @UIApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    #endif
    @State private var authViewModel = AuthViewModel()
    @State private var showSplash = true

    var body: some Scene {
        WindowGroup {
            ZStack {
                if showSplash {
                    SplashView(onFinished: {
                        withAnimation(.easeInOut(duration: 0.45)) {
                            showSplash = false
                        }
                    })
                    .transition(.opacity)
                    .zIndex(1)
                } else {
                    Group {
                        if authViewModel.isAuthenticated {
                            MainTabView(authViewModel: authViewModel)
                        } else {
                            AuthView(viewModel: authViewModel)
                        }
                    }
                    .transition(.opacity)
                }
            }
            .preferredColorScheme(.dark)
            .environment(\.locale, Locale(identifier: I18n.shared.currentLanguage.rawValue))
            .onOpenURL { url in
                #if canImport(UIKit)
                if url.scheme != "ekahub" {
                    _ = GIDSignIn.sharedInstance.handle(url)
                }
                #endif
            }
        }
    }
}
