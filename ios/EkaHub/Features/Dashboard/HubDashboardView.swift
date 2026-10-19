import SwiftUI

public struct HubDashboardView: View {
    @State private var viewModel = HubDashboardViewModel()
    public var onNavigateToChat: () -> Void = {}

    public init(onNavigateToChat: @escaping () -> Void = {}) {
        self.onNavigateToChat = onNavigateToChat
    }

    public var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 18) {
                    // Header
                    HStack {
                        VStack(alignment: .leading, spacing: 4) {
                            Text("Personal AI Hub")
                                .font(.title2)
                                .fontWeight(.bold)
                                .foregroundColor(.white)
                            Text("Life Aggregator & Intelligent Assistant")
                                .font(.caption)
                                .foregroundColor(.gray)
                        }
                        Spacer()
                        Button(action: onNavigateToChat) {
                            Image(systemName: "bubble.left.and.bubble.right.fill")
                                .foregroundColor(AppColors.primary)
                                .padding(10)
                                .background(AppColors.darkSurface)
                                .clipShape(Circle())
                        }
                    }

                    // Daily AI Briefing Banner
                    GradientAiBannerView(contentText: viewModel.briefingText)

                    // Bento Grid Satellites
                    Text(L10n.Hub.satellitesTitle)
                        .font(.headline)
                        .foregroundColor(.white)
                        .padding(.top, 4)

                    HStack(spacing: 12) {
                        ForEach(viewModel.summaries) { item in
                            Button(action: { DeepLinkManager.openURL(item.deepLink) }) {
                                BentoCardView {
                                    HStack {
                                        Image(systemName: item.appId.contains("tracker") ? "wallet.pass.fill" : "checklist")
                                            .foregroundColor(item.appId.contains("tracker") ? AppColors.emerald : AppColors.primary)
                                        Spacer()
                                        Image(systemName: "arrow.up.right")
                                            .font(.caption2)
                                            .foregroundColor(.gray)
                                    }

                                    Spacer()

                                    Text(item.appName)
                                        .font(.caption)
                                        .foregroundColor(.gray)

                                    Text(item.highlightMetric)
                                        .font(.system(size: 16, weight: .bold))
                                        .foregroundColor(.white)

                                    Text(item.highlightLabel)
                                        .font(.caption2)
                                        .foregroundColor(.secondary)

                                    Text(item.connectedEmail)
                                        .font(.system(size: 9))
                                        .padding(.horizontal, 6)
                                        .padding(.vertical, 2)
                                        .background(Color.black.opacity(0.3))
                                        .foregroundColor(.gray)
                                        .cornerRadius(4)
                                }
                                .frame(height: 150)
                            }
                        }
                    }
                }
                .padding(16)
            }
            .background(AppColors.darkBackground)
            .refreshable {
                await viewModel.refresh()
            }
        }
    }
}
