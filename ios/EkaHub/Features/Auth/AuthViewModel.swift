import SwiftUI

@MainActor
@Observable
public final class AuthViewModel {
    public var isAuthenticated: Bool = true
    public init() {}

    public func signOut() {
        KeychainManager.shared.clearAll()
        isAuthenticated = false
    }
}
