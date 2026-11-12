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

    public var flag: String {
        switch self {
        case .en: return "🇺🇸"
        case .id: return "🇮🇩"
        }
    }
}

@MainActor
@Observable
public final class I18n {
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

    public func toggleLanguage() {
        setLanguage(currentLanguage == .id ? .en : .id)
    }

    public func localizedString(for key: String) -> String {
        if let dict = translations[key], let val = dict[currentLanguage.rawValue] {
            return val
        }
        return key
    }

    private let translations: [String: [String: String]] = [
        "nav.chat": ["en": "Gemini", "id": "Gemini AI"],
        "nav.dashboard": ["en": "Briefing", "id": "Briefing"],
        "nav.apps": ["en": "Satellites", "id": "Satelit"],
        "nav.profile": ["en": "Profile", "id": "Profil"],
        "briefing.title": ["en": "Daily AI Briefing", "id": "Daily AI Briefing"],
        "satellites.title": ["en": "Connected Satellites", "id": "Aplikasi Satelit"],
        "chat.placeholder": ["en": "Ask Gemini, finances, or tasks...", "id": "Tanya Gemini, keuangan, atau tugas..."],
        "auth.signin": ["en": "Sign In", "id": "Masuk"],
        "auth.signout": ["en": "Sign Out", "id": "Keluar"],
        "auth.google": ["en": "Continue with Google", "id": "Lanjutkan dengan Google"],
        "auth.email_option": ["en": "Use Email & Password", "id": "Gunakan Email & Password"],
        "mismatch.title": ["en": "Account Mismatch", "id": "Akun Tidak Sesuai"]
    ]
}

public extension String {
    @MainActor
    var localized: String {
        I18n.shared.localizedString(for: self)
    }
}
