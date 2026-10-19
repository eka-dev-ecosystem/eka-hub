import Foundation
import SwiftUI

public enum AppLanguage: String, CaseIterable, Sendable {
    case en = "en"
    case id = "id"

    public var displayName: String {
        switch self {
        case .en: return "English"
        case .id: return "Bahasa Indonesia"
        }
    }
}

public final class I18n: @unchecked Sendable {
    public static let shared = I18n()

    private let userDefaultsKey = "selected_app_language"
    public private(set) var currentLanguage: AppLanguage

    private init() {
        if let saved = UserDefaults.standard.string(forKey: userDefaultsKey),
           let lang = AppLanguage(rawValue: saved) {
            self.currentLanguage = lang
        } else {
            self.currentLanguage = .id
        }
    }

    public func setLanguage(_ language: AppLanguage) {
        self.currentLanguage = language
        UserDefaults.standard.set(language.rawValue, forKey: userDefaultsKey)
    }

    public func localizedString(for key: String) -> String {
        if let dict = translations[key], let val = dict[currentLanguage.rawValue] {
            return val
        }
        return key
    }

    private let translations: [String: [String: String]] = [
        "nav.dashboard": ["en": "Dashboard", "id": "Dasbor"],
        "nav.chat": ["en": "AI Assistant", "id": "Asisten AI"],
        "nav.apps": ["en": "Satellites", "id": "Satelit"],
        "nav.profile": ["en": "Profile", "id": "Profil"],
        "briefing.title": ["en": "Daily AI Briefing", "id": "Daily AI Briefing"],
        "satellites.title": ["en": "Connected Satellites", "id": "Aplikasi Satelit Terhubung"],
        "chat.placeholder": ["en": "Ask about finances or tasks...", "id": "Tanyakan pengeluaran atau tugas..."],
        "auth.signin": ["en": "Sign In", "id": "Masuk"],
        "auth.signout": ["en": "Sign Out", "id": "Keluar"],
        "mismatch.title": ["en": "Account Mismatch", "id": "Akun Tidak Sesuai"]
    ]
}

public extension String {
    var localized: String {
        I18n.shared.localizedString(for: self)
    }
}
