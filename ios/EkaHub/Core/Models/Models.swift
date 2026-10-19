import Foundation

public struct BaseResponse<T: Codable>: Codable {
    public let success: BooleanLiteralType
    public let message: String?
    public let data: T?
    public let error: String?
}

public struct User: Codable, Identifiable {
    public let id: String
    public let name: String
    public let email: String
    public let avatar: String?
}

public struct AuthResponse: Codable {
    public let user: User
    public let accessToken: String
    public let refreshToken: String

    enum CodingKeys: String, CodingKey {
        case user
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
    }
}

public struct AppSummaryResponse: Codable, Identifiable {
    public var id: String { appId }
    public let appId: String
    public let appName: String
    public let connectedUid: String
    public let connectedEmail: String
    public let highlightMetric: String
    public let highlightLabel: String
    public let deepLink: String

    enum CodingKeys: String, CodingKey {
        case appId = "app_id"
        case appName = "app_name"
        case connectedUid = "connected_uid"
        case connectedEmail = "connected_email"
        case highlightMetric = "highlight_metric"
        case highlightLabel = "highlight_label"
        case deepLink = "deep_link"
    }

    public init(
        appId: String,
        appName: String,
        connectedUid: String,
        connectedEmail: String,
        highlightMetric: String,
        highlightLabel: String,
        deepLink: String
    ) {
        self.appId = appId
        self.appName = appName
        self.connectedUid = connectedUid
        self.connectedEmail = connectedEmail
        self.highlightMetric = highlightMetric
        self.highlightLabel = highlightLabel
        self.deepLink = deepLink
    }
}

public struct SuggestedAction: Codable, Identifiable {
    public var id: String { title }
    public let title: String
    public let deepLink: String

    enum CodingKeys: String, CodingKey {
        case title
        case deepLink = "deep_link"
    }

    public init(title: String, deepLink: String) {
        self.title = title
        self.deepLink = deepLink
    }
}

public struct ChatMessageItem: Identifiable {
    public let id = UUID()
    public let role: String // "user" or "assistant"
    public let content: String
    public let actions: [SuggestedAction]

    public init(role: String, content: String, actions: [SuggestedAction] = []) {
        self.role = role
        self.content = content
        self.actions = actions
    }
}

public struct AccountMismatchState: Identifiable {
    public let id = UUID()
    public let expectedUID: String
    public let expectedEmailHint: String
    public let currentUID: String
    public let currentEmail: String
    public let targetURL: URL
}
