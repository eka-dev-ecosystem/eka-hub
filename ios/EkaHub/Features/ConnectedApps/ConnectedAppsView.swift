import SwiftUI

private struct SatelliteDefinition: Identifiable {
    let id: String
    let name: String
    let subtitle: String
    let iconName: String
    let accentColor: Color
    let scheme: String
}

@MainActor
@Observable
public final class ConnectedAppsViewModel {
    public var linkedApps: [LinkedAppItem] = []
    public var isLoading = false
    public var linkingAppId: String?
    public var statusMessage: String?
    public var errorMessage: String?

    private let apiClient = APIClient.shared

    public init() {}

    public func loadApps() async {
        isLoading = true
        errorMessage = nil
        do {
            linkedApps = try await apiClient.fetchLinkedApps()
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    public func connectApp(_ appId: String, name: String) async {
        linkingAppId = appId
        errorMessage = nil
        statusMessage = nil

        do {
            let initRes = try await apiClient.initiateLink(appId: appId)
            DeepLinkManager.initiateLinkFlow(appId: appId, sessionCode: initRes.sessionCode)
            statusMessage = "Membuka \(name) untuk menyetujui koneksi..."
        } catch {
            errorMessage = "Gagal memulai handshake \(name): \(error.localizedDescription)"
        }

        linkingAppId = nil
    }

    public func disconnectApp(_ appId: String) async {
        linkingAppId = appId
        errorMessage = nil
        statusMessage = nil

        do {
            try await apiClient.unlinkApp(appId: appId)
            linkedApps.removeAll { $0.appId.lowercased() == appId.lowercased() }
            NotificationCenter.default.post(name: .linkedAppsDidChange, object: nil)
        } catch {
            errorMessage = "Gagal memutus koneksi: \(error.localizedDescription)"
        }

        linkingAppId = nil
    }
}

public struct ConnectedAppsView: View {
    @State private var viewModel = ConnectedAppsViewModel()

    private let catalog: [SatelliteDefinition] = [
        SatelliteDefinition(
            id: "ftracker",
            name: "F-Tracker",
            subtitle: "Expense, budget & cashflow real-time",
            iconName: "wallet.pass.fill",
            accentColor: AppColors.accentAmber,
            scheme: "ftracker://"
        ),
        SatelliteDefinition(
            id: "taskmanagement",
            name: "Task Management",
            subtitle: "Daily focus, deadlines & projects via MCP",
            iconName: "checklist",
            accentColor: AppColors.accentEmerald,
            scheme: "taskhub://"
        )
    ]

    public init() {}

    public var body: some View {
        ZStack {
            GeminiBackgroundView()

            ScrollView(showsIndicators: false) {
                VStack(alignment: .leading, spacing: 20) {
                    VStack(alignment: .leading, spacing: 4) {
                        Text("EKOSISTEM SATELIT")
                            .font(.system(size: 11, weight: .bold))
                            .tracking(1.1)
                            .foregroundStyle(AppColors.geminiGradient)

                        Text(L10n.Hub.satellitesTitle)
                            .font(.system(size: 26, weight: .bold, design: .rounded))
                            .foregroundColor(AppColors.textPrimary)

                        Text("Hubungkan aplikasi satelit dengan handshake aman agar Asisten AI bisa membaca konteks & menjalankan aksi langsung.")
                            .font(.system(size: 13))
                            .foregroundColor(AppColors.textSecondary)
                            .padding(.top, 2)
                    }
                    .padding(.top, 12)

                    if let status = viewModel.statusMessage {
                        HStack(spacing: 10) {
                            Image(systemName: "link.badge.plus")
                                .foregroundColor(AppColors.geminiCyan)
                            Text(status)
                                .font(.system(size: 13))
                                .foregroundColor(AppColors.geminiCyan)
                            Spacer()
                        }
                        .padding(14)
                        .glassCard(cornerRadius: 14)
                    }

                    if let error = viewModel.errorMessage {
                        HStack(spacing: 10) {
                            Image(systemName: "exclamationmark.triangle.fill")
                                .foregroundColor(AppColors.accentRose)
                            Text(error)
                                .font(.system(size: 13))
                                .foregroundColor(Color(red: 1.0, green: 0.72, blue: 0.76))
                            Spacer()
                        }
                        .padding(14)
                        .glassCard(cornerRadius: 14)
                    }

                    ForEach(catalog) { sat in
                        let linked = matchedApp(for: sat.id)
                        satelliteCard(definition: sat, linked: linked)
                    }

                    if viewModel.isLoading {
                        HStack {
                            Spacer()
                            ProgressView().tint(AppColors.geminiBlue)
                            Spacer()
                        }
                        .padding(.top, 8)
                    }

                    Spacer().frame(height: 24)
                }
                .padding(.horizontal, 18)
            }
            .refreshable {
                await viewModel.loadApps()
            }
        }
        .task {
            await viewModel.loadApps()
        }
        .onReceive(NotificationCenter.default.publisher(for: .linkedAppsDidChange)) { _ in
            Task { await viewModel.loadApps() }
        }
    }

    private func matchedApp(for satId: String) -> LinkedAppItem? {
        viewModel.linkedApps.first { item in
            let lower = item.appId.lowercased()
            if satId == "ftracker" { return lower.contains("tracker") }
            if satId == "taskmanagement" { return lower.contains("task") }
            return lower == satId
        }
    }

    private func satelliteCard(definition: SatelliteDefinition, linked: LinkedAppItem?) -> some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(spacing: 14) {
                ZStack {
                    RoundedRectangle(cornerRadius: 14, style: .continuous)
                        .fill(definition.accentColor.opacity(0.18))
                        .frame(width: 48, height: 48)
                    Image(systemName: definition.iconName)
                        .font(.system(size: 20, weight: .semibold))
                        .foregroundColor(definition.accentColor)
                }

                VStack(alignment: .leading, spacing: 3) {
                    HStack(spacing: 8) {
                        Text(definition.name)
                            .font(.system(size: 17, weight: .bold))
                            .foregroundColor(.white)

                        if linked != nil {
                            Text("Terhubung")
                                .font(.system(size: 10, weight: .bold))
                                .foregroundColor(AppColors.accentEmerald)
                                .padding(.horizontal, 8)
                                .padding(.vertical, 3)
                                .background(AppColors.accentEmerald.opacity(0.16))
                                .clipShape(Capsule())
                        }
                    }

                    Text(linked.map { "Akun: \($0.connectedEmail)" } ?? definition.subtitle)
                        .font(.system(size: 12))
                        .foregroundColor(AppColors.textSecondary)
                }
                Spacer()
            }

            if let linked, !linked.highlightMetric.isEmpty {
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text(linked.highlightLabel.isEmpty ? "Ringkasan Terakhir" : linked.highlightLabel)
                            .font(.system(size: 11))
                            .foregroundColor(AppColors.textMuted)
                        Text(linked.highlightMetric)
                            .font(.system(size: 16, weight: .bold, design: .rounded))
                            .foregroundColor(definition.accentColor)
                    }
                    Spacer()
                }
                .padding(12)
                .background(Color.white.opacity(0.04))
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
            }

            HStack(spacing: 10) {
                Button(action: {
                    Task { await viewModel.connectApp(definition.id, name: definition.name) }
                }) {
                    HStack(spacing: 6) {
                        if viewModel.linkingAppId == definition.id {
                            ProgressView().tint(.white)
                        } else {
                            Image(systemName: linked != nil ? "arrow.triangle.2.circlepath" : "link")
                        }
                        Text(linked != nil ? "Sinkronkan Ulang" : "Hubungkan Aplikasi")
                            .font(.system(size: 14, weight: .semibold))
                    }
                    .foregroundColor(.white)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 11)
                    .background(AppColors.auroraButtonGradient)
                    .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                }
                .buttonStyle(.plain)
                .disabled(viewModel.linkingAppId != nil)

                if let linked {
                    Button(action: {
                        Task { await viewModel.disconnectApp(linked.appId) }
                    }) {
                        Image(systemName: "trash")
                            .font(.system(size: 14, weight: .semibold))
                            .foregroundColor(AppColors.accentRose)
                            .frame(width: 44, height: 42)
                            .background(AppColors.accentRose.opacity(0.14))
                            .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    }
                    .buttonStyle(.plain)
                    .disabled(viewModel.linkingAppId != nil)
                }
            }
        }
        .padding(18)
        .glassCard(cornerRadius: 20)
    }
}
