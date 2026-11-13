import Foundation

public struct BaseResponse<T: Decodable>: Decodable {
    public let success: Bool
    public let message: String
    public let data: T?
    public let code: String?
}

public struct EmptyResponse: Decodable, Sendable {}

public enum JSONValue: Codable, Sendable, Hashable {
    case string(String)
    case int(Int)
    case double(Double)
    case bool(Bool)
    case array([JSONValue])
    case object([String: JSONValue])
    case null

    public init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        if container.decodeNil() {
            self = .null
        } else if let b = try? container.decode(Bool.self) {
            self = .bool(b)
        } else if let i = try? container.decode(Int.self) {
            self = .int(i)
        } else if let d = try? container.decode(Double.self) {
            self = .double(d)
        } else if let s = try? container.decode(String.self) {
            self = .string(s)
        } else if let a = try? container.decode([JSONValue].self) {
            self = .array(a)
        } else if let o = try? container.decode([String: JSONValue].self) {
            self = .object(o)
        } else {
            self = .null
        }
    }

    public func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        switch self {
        case .string(let s): try container.encode(s)
        case .int(let i): try container.encode(i)
        case .double(let d): try container.encode(d)
        case .bool(let b): try container.encode(b)
        case .array(let a): try container.encode(a)
        case .object(let o): try container.encode(o)
        case .null: try container.encodeNil()
        }
    }

    public var stringValue: String? {
        switch self {
        case .string(let s): return s
        case .int(let i): return String(i)
        case .double(let d): return String(d)
        case .bool(let b): return b ? "true" : "false"
        default: return nil
        }
    }
}

public struct User: Codable, Sendable, Equatable {
    public let id: String
    public let name: String
    public let email: String
    public let avatarUrl: String?
    public let authProvider: String?
    public let createdAt: Date?

    enum CodingKeys: String, CodingKey {
        case id, name, email
        case avatarUrl = "avatar_url"
        case authProvider = "auth_provider"
        case createdAt = "created_at"
    }

    public init(
        id: String,
        name: String,
        email: String,
        avatarUrl: String? = nil,
        authProvider: String? = nil,
        createdAt: Date? = nil
    ) {
        self.id = id
        self.name = name
        self.email = email
        self.avatarUrl = avatarUrl
        self.authProvider = authProvider
        self.createdAt = createdAt
    }
}

public struct TokenPair: Codable, Sendable {
    public let accessToken: String
    public let refreshToken: String
    public let expiresAt: Date?
    public let tokenType: String?

    enum CodingKeys: String, CodingKey {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case expiresAt = "expires_at"
        case tokenType = "token_type"
    }
}

public struct AuthResponse: Codable, Sendable {
    public let user: User
    public let tokens: TokenPair
}

public struct OAuthLoginRequest: Codable, Sendable {
    public let provider: String
    public let idToken: String?
    public let code: String?
    public let email: String?
    public let name: String?
    public let avatarUrl: String?

    enum CodingKeys: String, CodingKey {
        case provider
        case idToken = "id_token"
        case code, email, name
        case avatarUrl = "avatar_url"
    }

    public init(
        provider: String,
        idToken: String? = nil,
        code: String? = nil,
        email: String? = nil,
        name: String? = nil,
        avatarUrl: String? = nil
    ) {
        self.provider = provider
        self.idToken = idToken
        self.code = code
        self.email = email
        self.name = name
        self.avatarUrl = avatarUrl
    }
}

public struct LoginRequest: Codable, Sendable {
    public let email: String
    public let password: String
}

public struct RegisterRequest: Codable, Sendable {
    public let name: String
    public let email: String
    public let password: String
}

public struct RefreshTokenRequest: Codable, Sendable {
    public let refreshToken: String
    enum CodingKeys: String, CodingKey {
        case refreshToken = "refresh_token"
    }
}

public struct SuggestedAction: Codable, Identifiable, Sendable, Equatable {
    public var id: String { deepLink + title }
    public let title: String
    public let description: String
    public let appId: String
    public let deepLink: String

    enum CodingKeys: String, CodingKey {
        case title, description
        case appId = "app_id"
        case deepLink = "deep_link"
    }

    public init(title: String, description: String, appId: String, deepLink: String) {
        self.title = title
        self.description = description
        self.appId = appId
        self.deepLink = deepLink
    }
}

public struct AppSummaryResponse: Codable, Identifiable, Sendable {
    public var id: String { appId }
    public let appId: String
    public let appName: String
    public let connectedUid: String
    public let connectedEmail: String
    public let highlightMetric: String
    public let highlightLabel: String
    public let aiContextData: [String: JSONValue]?
    public let deepLink: String

    enum CodingKeys: String, CodingKey {
        case appId = "app_id"
        case appName = "app_name"
        case connectedUid = "connected_uid"
        case connectedEmail = "connected_email"
        case highlightMetric = "highlight_metric"
        case highlightLabel = "highlight_label"
        case aiContextData = "ai_context_data"
        case deepLink = "deep_link"
    }

    public init(
        appId: String,
        appName: String,
        connectedUid: String,
        connectedEmail: String,
        highlightMetric: String,
        highlightLabel: String,
        aiContextData: [String: JSONValue]? = nil,
        deepLink: String
    ) {
        self.appId = appId
        self.appName = appName
        self.connectedUid = connectedUid
        self.connectedEmail = connectedEmail
        self.highlightMetric = highlightMetric
        self.highlightLabel = highlightLabel
        self.aiContextData = aiContextData
        self.deepLink = deepLink
    }
}

public struct DailyBriefingResponse: Codable, Sendable {
    public let generatedAt: Date?
    public let dailyAiBriefing: String
    public let miniSummaryCards: [AppSummaryResponse]?
    public let suggestedActions: [SuggestedAction]?

    enum CodingKeys: String, CodingKey {
        case generatedAt = "generated_at"
        case dailyAiBriefing = "daily_ai_briefing"
        case miniSummaryCards = "mini_summary_cards"
        case suggestedActions = "suggested_actions"
    }
}

public struct LinkedAppItem: Codable, Identifiable, Sendable {
    public let id: String
    public let hubUserId: String
    public let appId: String
    public let appName: String
    public let connectedUid: String
    public let connectedEmail: String
    public let apiBaseUrl: String
    public let customUrlScheme: String
    public let isActive: Bool
    public let highlightMetric: String
    public let highlightLabel: String
    public let lastSyncedAt: Date?

    enum CodingKeys: String, CodingKey {
        case id
        case hubUserId = "hub_user_id"
        case appId = "app_id"
        case appName = "app_name"
        case connectedUid = "connected_uid"
        case connectedEmail = "connected_email"
        case apiBaseUrl = "api_base_url"
        case customUrlScheme = "custom_url_scheme"
        case isActive = "is_active"
        case highlightMetric = "highlight_metric"
        case highlightLabel = "highlight_label"
        case lastSyncedAt = "last_synced_at"
    }
}

public struct InitiateLinkRequest: Codable, Sendable {
    public let appId: String
    enum CodingKeys: String, CodingKey {
        case appId = "app_id"
    }
}

public struct InitiateLinkResponse: Codable, Sendable {
    public let sessionCode: String
    public let deepLinkUrl: String
    public let expiresInSeconds: Int64?

    enum CodingKeys: String, CodingKey {
        case sessionCode = "session_code"
        case deepLinkUrl = "deep_link_url"
        case expiresInSeconds = "expires_in_seconds"
    }
}

public struct ToolCallDetail: Codable, Sendable, Equatable {
    public let name: String
    public let arguments: [String: JSONValue]?
    public let result: JSONValue?
}

public struct AIChatRequest: Codable, Sendable {
    public let message: String
}

public struct AIChatResponse: Codable, Sendable {
    public let reply: String
    public let toolCalls: [ToolCallDetail]?
    public let actionLinks: [SuggestedAction]?

    enum CodingKeys: String, CodingKey {
        case reply
        case toolCalls = "tool_calls"
        case actionLinks = "action_links"
    }
}

public struct ChatMessageItem: Identifiable, Sendable {
    public let id: UUID
    public let role: ChatRole
    public var content: String
    public let toolCalls: [ToolCallDetail]?
    public let actionLinks: [SuggestedAction]?
    public let timestamp: Date

    public enum ChatRole: String, Sendable {
        case user
        case assistant
    }

    public init(
        id: UUID = UUID(),
        role: ChatRole,
        content: String,
        toolCalls: [ToolCallDetail]? = nil,
        actionLinks: [SuggestedAction]? = nil,
        timestamp: Date = Date()
    ) {
        self.id = id
        self.role = role
        self.content = content
        self.toolCalls = toolCalls
        self.actionLinks = actionLinks
        self.timestamp = timestamp
    }
}

public struct AccountMismatchState: Identifiable, Sendable {
    public var id: String { expectedUID + targetURL.absoluteString }
    public let expectedUID: String
    public let expectedEmailHint: String
    public let currentUID: String
    public let currentEmail: String
    public let targetURL: URL
}
