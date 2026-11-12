import SwiftUI

public struct HubDashboardView: View {
    @State private var viewModel = HubDashboardViewModel()
    var authViewModel: AuthViewModel
    var onNavigateToChat: (String?) -> Void
    var onOpenSatellites: () -> Void
    var onExecuteDeepLink: (URL) -> Void

    public init(
        authViewModel: AuthViewModel,
        onNavigateToChat: @escaping (String?) -> Void,
        onOpenSatellites: @escaping () -> Void = {},
        onExecuteDeepLink: @escaping (URL) -> Void
    ) {
        self.authViewModel = authViewModel
        self.onNavigateToChat = onNavigateToChat
        self.onOpenSatellites = onOpenSatellites
        self.onExecuteDeepLink = onExecuteDeepLink
    }

    public var body: some View {
        ZStack {
            GeminiBackgroundView()

            ScrollView(showsIndicators: false) {
                VStack(alignment: .leading, spacing: 22) {
                    // Header Bar
                    HStack {
                        VStack(alignment: .leading, spacing: 3) {
                            Text("LIFE AGGREGATOR")
                                .font(.system(size: 11, weight: .bold))
                                .tracking(1.1)
                                .foregroundStyle(AppColors.geminiGradient)

                            Text(L10n.Hub.briefingTitle)
                                .font(.system(size: 26, weight: .bold, design: .rounded))
                                .foregroundColor(AppColors.textPrimary)
                        }

                        Spacer()

                        Button(action: {
                            Task { await viewModel.loadDashboard() }
                        }) {
                            Image(systemName: "arrow.clockwise")
                                .font(.system(size: 15, weight: .semibold))
                                .foregroundColor(AppColors.textPrimary)
                                .frame(width: 38, height: 38)
                                .glassCard(cornerRadius: 19)
                        }
                        .buttonStyle(.plain)
                    }
                    .padding(.top, 12)

                    if let error = viewModel.errorMessage {
                        HStack(spacing: 10) {
                            Image(systemName: "wifi.exclamationmark")
                                .foregroundColor(AppColors.accentRose)
                            Text(error)
                                .font(.system(size: 13))
                                .foregroundColor(Color(red: 1.0, green: 0.72, blue: 0.76))
                            Spacer()
                        }
                        .padding(14)
                        .glassCard(cornerRadius: 16)
                    }

                    // Daily AI Briefing Card
                    let briefingCopy = viewModel.dailyBriefingText.isEmpty
                        ? "Tarik ke bawah untuk memuat ringkasan harian terbaru dari aplikasi satelit yang terhubung."
                        : viewModel.dailyBriefingText

                    GradientAiBannerView(briefingText: briefingCopy) {
                        onNavigateToChat("Bantu analisis ringkasan harian saya hari ini dan berikan saran fokus.")
                    }

                    // Bento Grid Section
                    HStack {
                        Text(L10n.Hub.satellitesTitle)
                            .font(.system(size: 18, weight: .bold, design: .rounded))
                            .foregroundColor(AppColors.textPrimary)
                        Spacer()
                        if viewModel.isLoading {
                            ProgressView().tint(AppColors.geminiBlue)
                        }
                    }

                    if viewModel.summaryCards.isEmpty && !viewModel.isLoading {
                        emptySatellitesState
                    } else {
                        LazyVGrid(columns: [
                            GridItem(.flexible(), spacing: 14),
                            GridItem(.flexible(), spacing: 14)
                        ], spacing: 14) {
                            ForEach(viewModel.summaryCards) { card in
                                BentoCardView(
                                    title: card.appName,
                                    metric: card.highlightMetric,
                                    subtitle: card.highlightLabel,
                                    iconName: iconForApp(card.appId),
                                    accentColor: accentForApp(card.appId)
                                ) {
                                    if let url = URL(string: card.deepLink) {
                                        onExecuteDeepLink(url)
                                    }
                                }
                            }
                        }
                    }

                    if !viewModel.suggestedActions.isEmpty {
                        VStack(alignment: .leading, spacing: 10) {
                            Text("Aksi Cepat")
                                .font(.system(size: 16, weight: .bold, design: .rounded))
                                .foregroundColor(AppColors.textPrimary)

                            ForEach(viewModel.suggestedActions) { action in
                                Button(action: {
                                    if let url = URL(string: action.deepLink) {
                                        onExecuteDeepLink(url)
                                    }
                                }) {
                                    HStack(spacing: 12) {
                                        Image(systemName: iconForApp(action.appId))
                                            .font(.system(size: 16, weight: .semibold))
                                            .foregroundColor(accentForApp(action.appId))
                                            .frame(width: 36, height: 36)
                                            .background(accentForApp(action.appId).opacity(0.16))
                                            .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))

                                        VStack(alignment: .leading, spacing: 2) {
                                            Text(action.title)
                                                .font(.system(size: 14, weight: .semibold))
                                                .foregroundColor(AppColors.textPrimary)
                                            Text(action.description)
                                                .font(.system(size: 12))
                                                .foregroundColor(AppColors.textSecondary)
                                        }
                                        Spacer()
                                        Image(systemName: "chevron.right")
                                            .font(.system(size: 12, weight: .bold))
                                            .foregroundColor(AppColors.textSecondary)
                                    }
                                    .padding(14)
                                    .glassCard(cornerRadius: 16)
                                }
                                .buttonStyle(.plain)
                            }
                        }
                    }

                    Spacer().frame(height: 24)
                }
                .padding(.horizontal, 18)
            }
            .refreshable {
                await viewModel.loadDashboard()
            }
        }
        .task {
            await viewModel.loadDashboard()
        }
        .onReceive(NotificationCenter.default.publisher(for: .linkedAppsDidChange)) { _ in
            Task {
                await viewModel.loadDashboard()
            }
        }
    }

    private var emptySatellitesState: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(spacing: 12) {
                Image(systemName: "app.connected.to.app.below.fill")
                    .font(.system(size: 24))
                    .foregroundStyle(AppColors.geminiGradient)

                VStack(alignment: .leading, spacing: 3) {
                    Text("Belum Ada Satelit Terhubung")
                        .font(.system(size: 15, weight: .bold))
                        .foregroundColor(.white)
                    Text("Hubungkan F-Tracker atau Task Management agar metrik keuangan & tugas muncul real-time di sini.")
                        .font(.system(size: 12))
                        .foregroundColor(AppColors.textSecondary)
                }
            }

            Button(action: onOpenSatellites) {
                HStack {
                    Image(systemName: "link")
                    Text("Buka Pengaturan Satelit")
                        .font(.system(size: 14, weight: .semibold))
                }
                .foregroundColor(.white)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 11)
                .background(AppColors.auroraButtonGradient)
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
            }
            .buttonStyle(.plain)
        }
        .padding(18)
        .glassCard(cornerRadius: 20)
    }

    private func iconForApp(_ appId: String) -> String {
        if appId.lowercased().contains("tracker") {
            return "dollarsign.circle.fill"
        }
        return "checklist"
    }

    private func accentForApp(_ appId: String) -> Color {
        if appId.lowercased().contains("tracker") {
            return AppColors.accentAmber
        }
        return AppColors.accentEmerald
    }
}
