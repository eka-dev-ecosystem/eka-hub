import SwiftUI

// AccountMismatchState is defined in Models.swift

public enum DeepLinkGuardResult {
    case allow(targetURL: URL)
    case requireLogin(pendingURL: URL, hint: String)
    case showMismatchSheet(state: AccountMismatchState)
}

public final class ChildAppDeepLinkGuard {
    public static let shared = ChildAppDeepLinkGuard()
    private init() {}

    public private(set) var pendingDeepLink: URL?

    public func evaluate(
        url: URL,
        currentUID: String?,
        currentEmail: String?
    ) -> DeepLinkGuardResult {
        guard let components = URLComponents(url: url, resolvingAgainstBaseURL: false) else {
            return .allow(targetURL: url)
        }

        let expectedUID = components.queryItems?.first(where: { $0.name == "expected_uid" })?.value

        // Regular deep link without account enforcement
        guard let expected = expectedUID, !expected.isEmpty else {
            return .allow(targetURL: url)
        }

        // Condition B: Sedang tidak login (null)
        guard let activeUID = currentUID, !activeUID.isEmpty else {
            pendingDeepLink = url
            let hint = "Tautan berasal dari Personal AI Hub. Silakan login untuk melanjutkan."
            return .requireLogin(pendingURL: url, hint: hint)
        }

        // Condition A: Akun cocok persis
        if activeUID == expected {
            return .allow(targetURL: url)
        }

        // Condition C: User ganti akun di tengah jalan (activeUID != expected)
        let expectedEmail = components.queryItems?.first(where: { $0.name == "expected_email" })?.value ?? "Akun Hub (\(expected))"
        let mismatchState = AccountMismatchState(
            expectedUID: expected,
            expectedEmailHint: expectedEmail,
            currentUID: activeUID,
            currentEmail: currentEmail ?? "Akun Aktif",
            targetURL: url
        )
        return .showMismatchSheet(state: mismatchState)
    }

    public func consumePendingDeepLink() -> URL? {
        let link = pendingDeepLink
        pendingDeepLink = nil
        return link
    }
}

/// SwiftUI View for Account Mismatch BottomSheet Dialog
public struct AccountMismatchSheetView: View {
    public let state: AccountMismatchState
    public let onSwitchAccount: () -> Void
    public let onUpdateHubConnection: () -> Void
    public let onCancel: () -> Void

    public init(
        state: AccountMismatchState,
        onSwitchAccount: @escaping () -> Void,
        onUpdateHubConnection: @escaping () -> Void,
        onCancel: @escaping () -> Void
    ) {
        self.state = state
        self.onSwitchAccount = onSwitchAccount
        self.onUpdateHubConnection = onUpdateHubConnection
        self.onCancel = onCancel
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack(spacing: 10) {
                Image(systemName: "exclamationmark.triangle.fill")
                    .foregroundColor(.yellow)
                    .font(.title2)
                Text("Akun Tidak Sesuai")
                    .font(.headline)
                    .foregroundColor(.white)
            }

            VStack(alignment: .leading, spacing: 4) {
                Text("Tautan dari Hub meminta akun:")
                    .font(.caption)
                    .foregroundColor(.gray)
                Text(state.expectedEmailHint)
                    .font(.subheadline)
                    .fontWeight(.semibold)
                    .foregroundColor(.cyan)
            }

            VStack(alignment: .leading, spacing: 4) {
                Text("Sedangkan akun yang sedang aktif saat ini:")
                    .font(.caption)
                    .foregroundColor(.gray)
                Text(state.currentEmail)
                    .font(.subheadline)
                    .fontWeight(.semibold)
                    .foregroundColor(.green)
            }

            Divider().background(Color.gray)

            // Option 1
            Button(action: onSwitchAccount) {
                Text("1. Ganti ke Akun Terkait Hub")
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.indigo)
                    .foregroundColor(.white)
                    .cornerRadius(12)
            }

            // Option 2
            Button(action: onUpdateHubConnection) {
                Text("2. Perbarui Koneksi Hub ke Akun Ini")
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.blue.opacity(0.2))
                    .foregroundColor(.cyan)
                    .overlay(
                        RoundedRectangle(cornerRadius: 12)
                            .stroke(Color.cyan, lineWidth: 1)
                    )
            }

            // Option 3
            Button(action: onCancel) {
                Text("3. Batal (Tetap di Akun Saat Ini)")
                    .frame(maxWidth: .infinity)
                    .padding(.top, 4)
                    .foregroundColor(.gray)
            }
        }
        .padding(24)
        .background(Color(red: 0.12, green: 0.16, blue: 0.23))
        .presentationDetents([.fraction(0.55)])
    }
}
