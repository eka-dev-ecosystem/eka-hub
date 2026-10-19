import Foundation

public enum L10n {
    public enum Nav {
        public static var dashboard: String { "nav.dashboard".localized }
        public static var chat: String { "nav.chat".localized }
        public static var apps: String { "nav.apps".localized }
        public static var profile: String { "nav.profile".localized }
    }

    public enum Hub {
        public static var briefingTitle: String { "briefing.title".localized }
        public static var satellitesTitle: String { "satellites.title".localized }
        public static var chatPlaceholder: String { "chat.placeholder".localized }
    }

    public enum Auth {
        public static var signIn: String { "auth.signin".localized }
        public static var signOut: String { "auth.signout".localized }
    }
}
