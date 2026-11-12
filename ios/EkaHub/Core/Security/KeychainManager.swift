import Foundation
import Security

public final class KeychainManager: @unchecked Sendable {
    public static let shared = KeychainManager()
    private init() {}

    private let service = "cloud.eka-dev.eka-hub"

    public func save(_ value: String, for key: String) {
        guard let data = value.data(using: .utf8) else { return }
        delete(for: key)

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleAfterFirstUnlock
        ]
        SecItemAdd(query as CFDictionary, nil)
    }

    public func get(for key: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        var dataTypeRef: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &dataTypeRef)
        guard status == errSecSuccess, let data = dataTypeRef as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    public func delete(for key: String) {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key
        ]
        SecItemDelete(query as CFDictionary)
    }

    public func clearAll() {
        let keys = [
            "access_token",
            "refresh_token",
            "user_id",
            "user_name",
            "user_email",
            "user_avatar",
            "auth_provider"
        ]
        for key in keys {
            delete(for: key)
        }
    }
}
