import SwiftUI

public struct ConnectedAppsView: View {
    public init() {}

    public var body: some View {
        NavigationStack {
            List {
                Section(header: Text("Aplikasi Satelit")) {
                    HStack(spacing: 12) {
                        Image(systemName: "wallet.pass.fill")
                            .foregroundColor(AppColors.emerald)
                            .font(.title2)
                        VStack(alignment: .leading) {
                            Text("F-Tracker")
                                .fontWeight(.semibold)
                                .foregroundColor(.white)
                            Text("Terhubung: eka@finance.app")
                                .font(.caption)
                                .foregroundColor(AppColors.cyan)
                        }
                        Spacer()
                        Button("Sync") {
                            let sessionCode = "sess_\(Int(Date().timeIntervalSince1970))"
                            DeepLinkManager.initiateLinkFlow(appId: "ftracker", sessionCode: sessionCode)
                        }
                        .font(.caption)
                        .buttonStyle(.borderedProminent)
                        .tint(AppColors.primary)
                    }
                    .listRowBackground(AppColors.darkSurface)

                    HStack(spacing: 12) {
                        Image(systemName: "checklist")
                            .foregroundColor(AppColors.primary)
                            .font(.title2)
                        VStack(alignment: .leading) {
                            Text("Task Management")
                                .fontWeight(.semibold)
                                .foregroundColor(.white)
                            Text("Terhubung: eka@taskhub.app")
                                .font(.caption)
                                .foregroundColor(AppColors.cyan)
                        }
                        Spacer()
                        Button("Sync") {
                            let sessionCode = "sess_\(Int(Date().timeIntervalSince1970))"
                            DeepLinkManager.initiateLinkFlow(appId: "taskhub", sessionCode: sessionCode)
                        }
                        .font(.caption)
                        .buttonStyle(.borderedProminent)
                        .tint(AppColors.primary)
                    }
                    .listRowBackground(AppColors.darkSurface)
                }
            }
            .scrollContentBackground(.hidden)
            .background(AppColors.darkBackground)
            .navigationTitle(L10n.Nav.apps)
        }
    }
}
