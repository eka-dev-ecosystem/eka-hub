import SwiftUI
import AuthenticationServices

public struct AuthView: View {
    @Bindable var viewModel: AuthViewModel

    @State private var showEmailForm = false
    @State private var isRegisterMode = false
    @State private var fullName = ""
    @State private var email = ""
    @State private var password = ""

    public init(viewModel: AuthViewModel) {
        self.viewModel = viewModel
    }

    public var body: some View {
        ZStack {
            GeminiBackgroundView()

            VStack(spacing: 0) {
                // Top language switcher bar
                HStack {
                    Spacer()
                    Button(action: {
                        I18n.shared.toggleLanguage()
                    }) {
                        HStack(spacing: 6) {
                            Image(systemName: "globe")
                                .font(.system(size: 11, weight: .semibold))
                            Text(I18n.shared.currentLanguage.displayName)
                                .font(.system(size: 12, weight: .semibold))
                            Text(I18n.shared.currentLanguage.flag)
                                .font(.system(size: 12))
                        }
                        .foregroundColor(AppColors.textPrimary)
                        .padding(.horizontal, 12)
                        .padding(.vertical, 7)
                        .glassCard(cornerRadius: 20)
                    }
                    .buttonStyle(.plain)
                }
                .padding(.horizontal, 20)
                .padding(.top, 12)

                ScrollView(showsIndicators: false) {
                    VStack(spacing: 24) {
                        Spacer().frame(height: 12)

                        // Hero Gemini-style Emblem & Brand Copy
                        VStack(spacing: 14) {
                            ZStack {
                                Circle()
                                    .fill(AppColors.geminiGradient.opacity(0.22))
                                    .frame(width: 92, height: 92)
                                    .blur(radius: 16)

                                RoundedRectangle(cornerRadius: 26, style: .continuous)
                                    .fill(AppColors.surfaceElevated)
                                    .frame(width: 78, height: 78)
                                    .overlay(
                                        RoundedRectangle(cornerRadius: 26, style: .continuous)
                                            .stroke(AppColors.geminiGradient, lineWidth: 1.5)
                                    )
                                    .shadow(color: AppColors.geminiBlue.opacity(0.35), radius: 18, x: 0, y: 8)

                                GeminiSparkleView(size: 34, animate: true)
                            }

                            VStack(spacing: 6) {
                                Text("Personal AI Hub")
                                    .font(.system(size: 30, weight: .bold, design: .rounded))
                                    .foregroundStyle(
                                        LinearGradient(
                                            colors: [.white, AppColors.textSecondary],
                                            startPoint: .top,
                                            endPoint: .bottom
                                        )
                                    )

                                Text("Asisten Gemini personal yang terintegrasi langsung dengan F-Tracker & Task Management.")
                                    .font(.system(size: 14))
                                    .foregroundColor(AppColors.textSecondary)
                                    .multilineTextAlignment(.center)
                                    .padding(.horizontal, 24)
                            }
                        }

                        // Capabilities Showcase Card
                        VStack(spacing: 12) {
                            AuthFeatureRow(
                                icon: "sparkles",
                                title: "Gemini AI Orchestrator",
                                desc: "Tanya, catat transaksi, atau buat tugas lewat satu chat.",
                                color: AppColors.geminiBlue
                            )
                            AuthFeatureRow(
                                icon: "wallet.pass.fill",
                                title: "F-Tracker Integration",
                                desc: "Pantau pengeluaran & ringkasan finansial real-time.",
                                color: AppColors.accentAmber
                            )
                            AuthFeatureRow(
                                icon: "checklist",
                                title: "Task Management Sync",
                                desc: "Kelola prioritas harian & fokus tanpa pindah konteks.",
                                color: AppColors.accentEmerald
                            )
                        }
                        .padding(18)
                        .glassCard(cornerRadius: 22)
                        .padding(.horizontal, 22)

                        // Inline Error Message
                        if let error = viewModel.errorMessage {
                            HStack(alignment: .top, spacing: 10) {
                                Image(systemName: "exclamationmark.triangle.fill")
                                    .foregroundColor(AppColors.accentRose)
                                    .font(.system(size: 15))
                                Text(error)
                                    .font(.system(size: 13))
                                    .foregroundColor(Color(red: 1.0, green: 0.72, blue: 0.76))
                                    .fixedSize(horizontal: false, vertical: true)
                                Spacer()
                            }
                            .padding(14)
                            .background(AppColors.accentRose.opacity(0.14))
                            .overlay(
                                RoundedRectangle(cornerRadius: 14, style: .continuous)
                                    .stroke(AppColors.accentRose.opacity(0.35), lineWidth: 1)
                            )
                            .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                            .padding(.horizontal, 22)
                        }

                        // Primary OAuth Buttons
                        VStack(spacing: 12) {
                            // Google Sign-In
                            Button(action: {
                                viewModel.handleGoogleSignIn()
                            }) {
                                HStack(spacing: 10) {
                                    if viewModel.isGoogleLoading {
                                        ProgressView()
                                            .tint(.black)
                                    } else {
                                        ZStack {
                                            Circle()
                                                .fill(AppColors.geminiBlue.opacity(0.12))
                                                .frame(width: 26, height: 26)
                                            Text("G")
                                                .font(.system(size: 15, weight: .heavy, design: .rounded))
                                                .foregroundStyle(AppColors.geminiGradient)
                                        }
                                    }

                                    Text(viewModel.isGoogleLoading ? "Menghubungkan Google..." : L10n.Auth.googleSignIn)
                                        .font(.system(size: 16, weight: .semibold))
                                        .foregroundColor(.black)
                                }
                                .frame(maxWidth: .infinity)
                                .frame(height: 52)
                                .background(Color.white)
                                .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
                                .shadow(color: Color.white.opacity(0.14), radius: 10, x: 0, y: 4)
                            }
                            .buttonStyle(.plain)
                            .disabled(viewModel.isLoading)

                            // Sign in with Apple
                            ZStack {
                                SignInWithAppleButton(.signIn) { request in
                                    request.requestedScopes = [.fullName, .email]
                                } onCompletion: { result in
                                    viewModel.handleAppleSignIn(result: result)
                                }
                                .signInWithAppleButtonStyle(.whiteOutline)
                                .frame(height: 52)
                                .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
                                .disabled(viewModel.isLoading)

                                if viewModel.isAppleLoading {
                                    RoundedRectangle(cornerRadius: 16, style: .continuous)
                                        .fill(Color.black.opacity(0.85))
                                        .frame(height: 52)
                                        .overlay(
                                            HStack(spacing: 8) {
                                                ProgressView().tint(.white)
                                                Text("Memverifikasi Apple...")
                                                    .font(.system(size: 15, weight: .semibold))
                                                    .foregroundColor(.white)
                                            }
                                        )
                                }
                            }

                            // Divider + Toggle Email Login
                            Button(action: {
                                withAnimation(.spring(response: 0.35, dampingFraction: 0.85)) {
                                    showEmailForm.toggle()
                                    viewModel.errorMessage = nil
                                }
                            }) {
                                HStack(spacing: 6) {
                                    Image(systemName: showEmailForm ? "chevron.up" : "envelope.fill")
                                        .font(.system(size: 12, weight: .semibold))
                                    Text(L10n.Auth.emailOption)
                                        .font(.system(size: 14, weight: .semibold))
                                }
                                .foregroundColor(AppColors.textSecondary)
                                .padding(.vertical, 10)
                                .frame(maxWidth: .infinity)
                            }
                            .buttonStyle(.plain)

                            if showEmailForm {
                                emailAuthForm
                                    .transition(.opacity.combined(with: .move(edge: .top)))
                            }
                        }
                        .padding(.horizontal, 22)

                        Spacer().frame(height: 28)
                    }
                }
            }
        }
    }

    private var emailAuthForm: some View {
        VStack(spacing: 14) {
            Picker("", selection: $isRegisterMode) {
                Text("Masuk").tag(false)
                Text("Daftar Akun").tag(true)
            }
            .pickerStyle(.segmented)

            if isRegisterMode {
                TextField("Nama Lengkap", text: $fullName)
                    .textContentType(.name)
                    .padding(14)
                    .background(Color.white.opacity(0.06))
                    .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    .foregroundColor(.white)
            }

            TextField("Email", text: $email)
                .textContentType(.emailAddress)
                .keyboardType(.emailAddress)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .padding(14)
                .background(Color.white.opacity(0.06))
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                .foregroundColor(.white)

            SecureField("Password (min. 8 karakter)", text: $password)
                .textContentType(isRegisterMode ? .newPassword : .password)
                .padding(14)
                .background(Color.white.opacity(0.06))
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                .foregroundColor(.white)

            Button(action: {
                Task {
                    if isRegisterMode {
                        await viewModel.registerWithEmail(name: fullName, email: email, password: password)
                    } else {
                        await viewModel.signInWithEmail(email: email, password: password)
                    }
                }
            }) {
                HStack {
                    if viewModel.isEmailLoading {
                        ProgressView().tint(.white)
                    } else {
                        Text(isRegisterMode ? "Buat Akun & Masuk" : "Masuk dengan Email")
                            .font(.system(size: 15, weight: .semibold))
                    }
                }
                .foregroundColor(.white)
                .frame(maxWidth: .infinity)
                .frame(height: 48)
                .background(AppColors.auroraButtonGradient)
                .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
            }
            .buttonStyle(.plain)
            .disabled(viewModel.isLoading)
        }
        .padding(16)
        .glassCard(cornerRadius: 18)
    }
}

private struct AuthFeatureRow: View {
    let icon: String
    let title: String
    let desc: String
    let color: Color

    var body: some View {
        HStack(spacing: 12) {
            ZStack {
                Circle()
                    .fill(color.opacity(0.16))
                    .frame(width: 36, height: 36)
                Image(systemName: icon)
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundColor(color)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(AppColors.textPrimary)
                Text(desc)
                    .font(.system(size: 12))
                    .foregroundColor(AppColors.textSecondary)
            }
            Spacer()
        }
    }
}
