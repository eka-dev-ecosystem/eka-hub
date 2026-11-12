import Foundation

@MainActor
public enum L10n {
    @MainActor
    public enum Nav {
        public static var chat: String { "nav.chat".localized }
        public static var dashboard: String { "nav.dashboard".localized }
        public static var apps: String { "nav.apps".localized }
        public static var profile: String { "nav.profile".localized }
    }

    @MainActor
    public enum Hub {
        public static var briefingTitle: String { "briefing.title".localized }
        public static var satellitesTitle: String { "satellites.title".localized }
        public static var chatPlaceholder: String { "chat.placeholder".localized }
    }

    @MainActor
    public enum Auth {
        public static var signIn: String { "auth.signin".localized }
        public static var signOut: String { "auth.signout".localized }
        public static var googleSignIn: String { "auth.google".localized }
        public static var emailOption: String { "auth.email_option".localized }
    }
}
