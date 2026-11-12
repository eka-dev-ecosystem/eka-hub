import SwiftUI

public struct MainTabView: View {
    var authViewModel: AuthViewModel
    // Default landing tab = 0 (Gemini AI Chat) per Personal AI Hub UX
    @State private var selectedTab: Int = 0
    @State private var incomingChatPrompt: String?
    @State private var mismatchState: AccountMismatchState?

    public init(authViewModel: AuthViewModel) {
        self.authViewModel = authViewModel
    }

    public var body: some View {
        TabView(selection: $selectedTab) {
            AiChatView(
                authViewModel: authViewModel,
                incomingPrompt: $incomingChatPrompt,
                onOpenSatellites: {
                    selectedTab = 2
                },
                onExecuteDeepLink: { url in
                    executeDeepLink(url)
                }
            )
            .tabItem {
                Label(L10n.Nav.chat, systemImage: "sparkles")
            }
            .tag(0)

            HubDashboardView(
                authViewModel: authViewModel,
                onNavigateToChat: { prompt in
                    if let prompt, !prompt.isEmpty {
                        incomingChatPrompt = prompt
                    }
                    selectedTab = 0
                },
                onOpenSatellites: {
                    selectedTab = 2
                },
                onExecuteDeepLink: { url in
                    executeDeepLink(url)
                }
            )
            .tabItem {
                Label(L10n.Nav.dashboard, systemImage: "rectangle.grid.2x2.fill")
            }
            .tag(1)

            ConnectedAppsView()
                .tabItem {
                    Label(L10n.Nav.apps, systemImage: "link.circle.fill")
                }
                .tag(2)

            ProfileView(authViewModel: authViewModel)
                .tabItem {
                    Label(L10n.Nav.profile, systemImage: "person.crop.circle.fill")
                }
                .tag(3)
        }
        .tint(AppColors.geminiBlue)
        .onOpenURL { url in
            handleIncomingDeepLink(url)
        }
        .sheet(item: $mismatchState) { state in
            AccountMismatchSheetView(
                state: state,
                onSwitchAccount: {
                    mismatchState = nil
                    DeepLinkManager.openURL(state.targetURL.absoluteString)
                },
                onUpdateHubConnection: {
                    mismatchState = nil
                    selectedTab = 2
                },
                onCancel: {
                    mismatchState = nil
                }
            )
        }
    }

    private func handleIncomingDeepLink(_ url: URL) {
        if url.scheme == "ekahub" && url.host == "link-callback" {
            NotificationCenter.default.post(name: .linkedAppsDidChange, object: nil)
            selectedTab = 2
            return
        }

        let currentUID = authViewModel.currentUser?.id ?? KeychainManager.shared.get(for: "user_id")
        let currentEmail = authViewModel.currentUser?.email ?? KeychainManager.shared.get(for: "user_email")

        let result = ChildAppDeepLinkGuard.shared.evaluate(
            url: url,
            currentUID: currentUID,
            currentEmail: currentEmail
        )

        switch result {
        case .allow(let targetURL):
            DeepLinkManager.openURL(targetURL.absoluteString)
        case .requireLogin:
            authViewModel.signOut()
        case .showMismatchSheet(let state):
            mismatchState = state
        }
    }

    private func executeDeepLink(_ url: URL) {
        DeepLinkManager.openURL(url.absoluteString)
    }
}
