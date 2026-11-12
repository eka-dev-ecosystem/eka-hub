import SwiftUI
import AuthenticationServices
import GoogleSignIn
#if canImport(UIKit)
import UIKit
#endif

@MainActor
@Observable
public final class AuthViewModel {
    public var isAuthenticated: Bool = false
    public var isLoading: Bool = false
    public var isAppleLoading: Bool = false
    public var isGoogleLoading: Bool = false
    public var isEmailLoading: Bool = false
    public var errorMessage: String?
    public var currentUser: User?

    private let apiClient = APIClient.shared
    private let keychain = KeychainManager.shared

    public init() {
        restoreSession()
        observeSessionExpiration()
    }

    private func observeSessionExpiration() {
        NotificationCenter.default.addObserver(
            forName: .sessionExpired,
            object: nil,
            queue: .main
        ) { [weak self] _ in
            Task { @MainActor [weak self] in
                self?.forceLocalSignOut(message: "Sesi login telah berakhir. Silakan masuk kembali.")
            }
        }
    }

    public func restoreSession() {
        loadCachedUser()
        if let token = keychain.get(for: "access_token"), !token.isEmpty {
            isAuthenticated = true
            Task {
                await refreshUserProfile()
            }
        } else if let refresh = keychain.get(for: "refresh_token"), !refresh.isEmpty {
            isAuthenticated = true
            Task {
                let ok = await apiClient.refreshToken()
                if ok {
                    await refreshUserProfile()
                } else {
                    self.forceLocalSignOut(message: nil)
                }
            }
        } else {
            isAuthenticated = false
            currentUser = nil
        }
    }

    private func loadCachedUser() {
        guard let uid = keychain.get(for: "user_id"), !uid.isEmpty else { return }
        currentUser = User(
            id: uid,
            name: keychain.get(for: "user_name") ?? "Pengguna Hub",
            email: keychain.get(for: "user_email") ?? "",
            avatarUrl: keychain.get(for: "user_avatar"),
            authProvider: keychain.get(for: "auth_provider")
        )
    }

    public func refreshUserProfile() async {
        do {
            let user = try await apiClient.getProfile()
            self.currentUser = user
        } catch {
            AppLogger.warn("Background profile refresh failed", category: .auth, metadata: ["error": error.localizedDescription])
        }
    }

    public func handleAppleSignIn(result: Result<ASAuthorization, Error>) {
        isAppleLoading = true
        isLoading = true
        errorMessage = nil

        switch result {
        case .success(let auth):
            guard let credential = auth.credential as? ASAuthorizationAppleIDCredential,
                  let tokenData = credential.identityToken,
                  let idToken = String(data: tokenData, encoding: .utf8) else {
                isAppleLoading = false
                isLoading = false
                errorMessage = "Gagal membaca token autentikasi dari Apple."
                return
            }

            let fullName = [
                credential.fullName?.givenName,
                credential.fullName?.familyName
            ].compactMap { $0 }.joined(separator: " ")

            Task {
                do {
                    let res = try await apiClient.oauthLogin(
                        OAuthLoginRequest(
                            provider: "apple",
                            idToken: idToken,
                            code: nil,
                            email: credential.email,
                            name: fullName.isEmpty ? nil : fullName,
                            avatarUrl: nil
                        )
                    )
                    self.currentUser = res.user
                    self.isAuthenticated = true
                } catch {
                    self.errorMessage = error.localizedDescription
                }
                self.isAppleLoading = false
                self.isLoading = false
            }

        case .failure(let error):
            if let authErr = error as? ASAuthorizationError {
                switch authErr.code {
                case .canceled:
                    break
                case .unknown:
                    errorMessage = "Sign in with Apple memerlukan capability aktif di Provisioning Profile. Gunakan Google Sign-In atau Email untuk masuk."
                default:
                    errorMessage = authErr.localizedDescription
                }
            } else {
                errorMessage = error.localizedDescription
            }
            isAppleLoading = false
            isLoading = false
        }
    }

    public func handleGoogleSignIn() {
        #if canImport(UIKit)
        guard let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
              let rootVC = windowScene.windows.first(where: { $0.isKeyWindow })?.rootViewController
                ?? windowScene.windows.first?.rootViewController else {
            errorMessage = "Tidak dapat menemukan tampilan utama untuk menampilkan Google Sign-In."
            return
        }

        isGoogleLoading = true
        isLoading = true
        errorMessage = nil

        GIDSignIn.sharedInstance.configuration = GIDConfiguration(clientID: AppConfig.googleClientID)

        GIDSignIn.sharedInstance.signIn(withPresenting: rootVC) { [weak self] signInResult, error in
            Task { @MainActor [weak self] in
                guard let self else { return }

                if let error {
                    let nsError = error as NSError
                    if nsError.code != GIDSignInError.canceled.rawValue {
                        self.errorMessage = error.localizedDescription
                    }
                    self.isGoogleLoading = false
                    self.isLoading = false
                    return
                }

                guard let gidUser = signInResult?.user,
                      let idToken = gidUser.idToken?.tokenString else {
                    self.errorMessage = "Google Sign-In tidak mengembalikan token identitas yang valid."
                    self.isGoogleLoading = false
                    self.isLoading = false
                    return
                }

                do {
                    let res = try await self.apiClient.oauthLogin(
                        OAuthLoginRequest(
                            provider: "google",
                            idToken: idToken,
                            code: nil,
                            email: gidUser.profile?.email,
                            name: gidUser.profile?.name,
                            avatarUrl: gidUser.profile?.imageURL(withDimension: 200)?.absoluteString
                        )
                    )
                    self.currentUser = res.user
                    self.isAuthenticated = true
                } catch {
                    self.errorMessage = error.localizedDescription
                }
                self.isGoogleLoading = false
                self.isLoading = false
            }
        }
        #endif
    }

    public func signInWithEmail(email: String, password: String) async {
        let trimmedEmail = email.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        guard !trimmedEmail.isEmpty, !password.isEmpty else {
            errorMessage = "Email dan password wajib diisi."
            return
        }

        isEmailLoading = true
        isLoading = true
        errorMessage = nil

        do {
            let res = try await apiClient.login(email: trimmedEmail, password: password)
            currentUser = res.user
            isAuthenticated = true
        } catch {
            errorMessage = error.localizedDescription
        }
        isEmailLoading = false
        isLoading = false
    }

    public func registerWithEmail(name: String, email: String, password: String) async {
        let trimmedName = name.trimmingCharacters(in: .whitespacesAndNewlines)
        let trimmedEmail = email.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        guard trimmedName.count >= 3 else {
            errorMessage = "Nama minimal 3 karakter."
            return
        }
        guard !trimmedEmail.isEmpty, password.count >= 8 else {
            errorMessage = "Masukkan email valid dan password minimal 8 karakter."
            return
        }

        isEmailLoading = true
        isLoading = true
        errorMessage = nil

        do {
            let res = try await apiClient.register(name: trimmedName, email: trimmedEmail, password: password)
            currentUser = res.user
            isAuthenticated = true
        } catch {
            errorMessage = error.localizedDescription
        }
        isEmailLoading = false
        isLoading = false
    }

    public func signOut() {
        #if canImport(UIKit)
        GIDSignIn.sharedInstance.signOut()
        #endif
        Task {
            await apiClient.logout()
        }
        keychain.clearAll()
        currentUser = nil
        errorMessage = nil
        isAuthenticated = false
    }

    private func forceLocalSignOut(message: String?) {
        keychain.clearAll()
        currentUser = nil
        errorMessage = message
        isAuthenticated = false
    }
}
