import SwiftUI

public struct MainTabView: View {
    @State private var selectedTab: Int = 0
    @State private var mismatchState: AccountMismatchState?

    public init() {}

    public var body: some View {
        TabView(selection: $selectedTab) {
            HubDashboardView(onNavigateToChat: {
                selectedTab = 1
            })
            .tabItem {
                Label(L10n.Nav.dashboard, systemImage: "square.grid.2x2.fill")
            }
            .tag(0)

            AiChatView()
            .tabItem {
                Label(L10n.Nav.chat, systemImage: "bubble.left.and.bubble.right.fill")
            }
            .tag(1)

            ConnectedAppsView()
            .tabItem {
                Label(L10n.Nav.apps, systemImage: "link.circle.fill")
            }
            .tag(2)

            ProfileView(onSignOut: {
                // handle signout
            })
            .tabItem {
                Label(L10n.Nav.profile, systemImage: "person.crop.circle.fill")
            }
            .tag(3)
        }
        .tint(AppColors.primary)
        .sheet(item: $mismatchState) { state in
            AccountMismatchSheetView(
                state: state,
                onSwitchAccount: {
                    mismatchState = nil
                    AppLogger.info("Switch account initiated from mismatch sheet", category: .deeplink)
                },
                onUpdateHubConnection: {
                    mismatchState = nil
                    AppLogger.info("Update connection initiated from mismatch sheet", category: .deeplink)
                },
                onCancel: {
                    mismatchState = nil
                }
            )
        }
        .onOpenURL { url in
            handleIncomingURL(url)
        }
    }

    private func handleIncomingURL(_ url: URL) {
        AppLogger.info("Received incoming URL: \(url.absoluteString)", category: .deeplink)
        let guardResult = ChildAppDeepLinkGuard.shared.evaluate(
            url: url,
            currentUID: KeychainManager.shared.get(for: "user_id") ?? "usr_hub_current",
            currentEmail: "eka@eka-dev.cloud"
        )

        switch guardResult {
        case .allow(let target):
            AppLogger.success("Deep link allowed: \(target)", category: .deeplink)
        case .requireLogin(let pending, let hint):
            AppLogger.warn("Deep link requires login: \(pending) (\(hint))", category: .deeplink)
        case .showMismatchSheet(let state):
            AppLogger.warn("Account mismatch detected: expected \(state.expectedUID)", category: .deeplink)
            mismatchState = state
        }
    }
}
