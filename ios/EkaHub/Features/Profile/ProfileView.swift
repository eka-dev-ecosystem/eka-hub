import SwiftUI

public struct ProfileView: View {
    var authViewModel: AuthViewModel
    @State private var selectedLang: AppLanguage = I18n.shared.currentLanguage
    @State private var showSignOutConfirmation = false

    public init(authViewModel: AuthViewModel) {
        self.authViewModel = authViewModel
    }

    private var userName: String {
        authViewModel.currentUser?.name ?? "Pengguna Hub"
    }

    private var userEmail: String {
        let e = authViewModel.currentUser?.email ?? ""
        return e.isEmpty ? "Email belum tersedia" : e
    }

    private var userInitials: String {
        let parts = userName.split(separator: " ")
        let initials = parts.prefix(2).compactMap { $0.first }.map(String.init).joined()
        return initials.isEmpty ? "EH" : initials.uppercased()
    }

    private var providerLabel: String {
        switch authViewModel.currentUser?.authProvider?.lowercased() {
        case "google": return "Google Account"
        case "apple": return "Apple ID"
        default: return "Eka-Hub Account"
        }
    }

    private var appVersionText: String {
        let version = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "1.0.0"
        let build = Bundle.main.infoDictionary?["CFBundleVersion"] as? String ?? "1"
        return "\(version) (\(build))"
    }

    public var body: some View {
        ZStack {
            GeminiBackgroundView()

            ScrollView(showsIndicators: false) {
                VStack(spacing: 22) {
                    // User Identity Hero Card
                    VStack(spacing: 14) {
                        ZStack {
                            Circle()
                                .fill(AppColors.geminiGradient)
                                .frame(width: 82, height: 82)
                                .blur(radius: 12)
                                .opacity(0.45)

                            if let avatarStr = authViewModel.currentUser?.avatarUrl,
                               let url = URL(string: avatarStr),
                               !avatarStr.isEmpty {
                                AsyncImage(url: url) { phase in
                                    switch phase {
                                    case .success(let img):
                                        img.resizable().scaledToFill()
                                    default:
                                        initialsAvatar
                                    }
                                }
                                .frame(width: 76, height: 76)
                                .clipShape(Circle())
                                .overlay(Circle().stroke(AppColors.geminiGradient, lineWidth: 2))
                            } else {
                                initialsAvatar
                            }
                        }

                        VStack(spacing: 4) {
                            Text(userName)
                                .font(.system(size: 20, weight: .bold, design: .rounded))
                                .foregroundColor(.white)

                            Text(userEmail)
                                .font(.system(size: 13))
                                .foregroundColor(AppColors.textSecondary)

                            Text(providerLabel)
                                .font(.system(size: 11, weight: .bold))
                                .foregroundColor(AppColors.geminiCyan)
                                .padding(.horizontal, 10)
                                .padding(.vertical, 4)
                                .background(AppColors.geminiCyan.opacity(0.14))
                                .clipShape(Capsule())
                                .padding(.top, 4)
                        }
                    }
                    .frame(maxWidth: .infinity)
                    .padding(22)
                    .glassCard(cornerRadius: 24)
                    .padding(.top, 14)

                    // Preferences Section
                    VStack(alignment: .leading, spacing: 14) {
                        Text("PREFERENSI")
                            .font(.system(size: 11, weight: .bold))
                            .tracking(0.9)
                            .foregroundColor(AppColors.textMuted)

                        VStack(spacing: 14) {
                            HStack {
                                Label("Bahasa / Language", systemImage: "globe")
                                    .font(.system(size: 14, weight: .medium))
                                    .foregroundColor(.white)
                                Spacer()
                                Picker("", selection: $selectedLang) {
                                    ForEach(AppLanguage.allCases, id: \.self) { lang in
                                        Text(lang.displayName).tag(lang)
                                    }
                                }
                                .pickerStyle(.menu)
                                .tint(AppColors.geminiBlue)
                                .onChange(of: selectedLang) { _, newValue in
                                    I18n.shared.setLanguage(newValue)
                                }
                            }

                            Divider().background(AppColors.borderSubtle)

                            HStack {
                                Label("Versi Aplikasi", systemImage: "info.circle")
                                    .font(.system(size: 14, weight: .medium))
                                    .foregroundColor(.white)
                                Spacer()
                                Text(appVersionText)
                                    .font(.system(size: 13, weight: .medium))
                                    .foregroundColor(AppColors.textSecondary)
                            }
                        }
                        .padding(16)
                        .glassCard(cornerRadius: 20)
                    }

                    // Sign Out Button
                    Button(action: {
                        showSignOutConfirmation = true
                    }) {
                        HStack(spacing: 8) {
                            Image(systemName: "rectangle.portrait.and.arrow.right")
                            Text(L10n.Auth.signOut)
                                .font(.system(size: 15, weight: .bold))
                        }
                        .foregroundColor(AppColors.accentRose)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 14)
                        .background(AppColors.accentRose.opacity(0.12))
                        .overlay(
                            RoundedRectangle(cornerRadius: 16, style: .continuous)
                                .stroke(AppColors.accentRose.opacity(0.32), lineWidth: 1)
                        )
                        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
                    }
                    .buttonStyle(.plain)

                    Spacer().frame(height: 24)
                }
                .padding(.horizontal, 18)
            }
        }
        .confirmationDialog(
            "Keluar dari Personal AI Hub?",
            isPresented: $showSignOutConfirmation,
            titleVisibility: .visible
        ) {
            Button(L10n.Auth.signOut, role: .destructive) {
                authViewModel.signOut()
            }
            Button("Batal", role: .cancel) {}
        } message: {
            Text("Anda perlu masuk lagi untuk mengakses asisten AI dan satelit terhubung.")
        }
    }

    private var initialsAvatar: some View {
        ZStack {
            Circle()
                .fill(AppColors.surfaceCard)
                .frame(width: 76, height: 76)
                .overlay(Circle().stroke(AppColors.geminiGradient, lineWidth: 2))
            Text(userInitials)
                .font(.system(size: 24, weight: .bold, design: .rounded))
                .foregroundStyle(AppColors.geminiGradient)
        }
    }
}
