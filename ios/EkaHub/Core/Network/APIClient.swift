import Foundation

public extension Notification.Name {
    static let sessionExpired = Notification.Name("EkaHubSessionExpired")
    static let linkedAppsDidChange = Notification.Name("EkaHubLinkedAppsDidChange")
}

public enum APIError: LocalizedError, Sendable {
    case invalidURL
    case unauthorized(String)
    case serverError(statusCode: Int, message: String)
    case decodingError(String)
    case networkError(String)

    public var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "URL API tidak valid."
        case .unauthorized(let msg):
            return msg.isEmpty ? "Sesi berakhir, silakan login ulang." : msg
        case .serverError(_, let msg):
            return msg
        case .decodingError(let msg):
            return "Gagal memproses respons server: \(msg)"
        case .networkError(let msg):
            return msg
        }
    }
}

public actor APIClient {
    public static let shared = APIClient()

    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    private init() {
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 30
        config.timeoutIntervalForResource = 60
        self.session = URLSession(configuration: config)

        let jsonDecoder = JSONDecoder()
        jsonDecoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let str = try container.decode(String.self)

            let f1 = ISO8601DateFormatter()
            f1.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
            if let d = f1.date(from: str) { return d }

            let f2 = ISO8601DateFormatter()
            f2.formatOptions = [.withInternetDateTime]
            if let d = f2.date(from: str) { return d }

            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Invalid date format: \(str)")
        }
        self.decoder = jsonDecoder
        self.encoder = JSONEncoder()
    }

    // MARK: - Auth Endpoints

    public func oauthLogin(_ req: OAuthLoginRequest) async throws -> AuthResponse {
        let auth: AuthResponse = try await request("/auth/oauth", method: "POST", body: req, requiresAuth: false)
        persistSession(auth)
        return auth
    }

    public func login(email: String, password: String) async throws -> AuthResponse {
        let req = LoginRequest(email: email, password: password)
        let auth: AuthResponse = try await request("/auth/login", method: "POST", body: req, requiresAuth: false)
        persistSession(auth)
        return auth
    }

    public func register(name: String, email: String, password: String) async throws -> AuthResponse {
        let req = RegisterRequest(name: name, email: email, password: password)
        let auth: AuthResponse = try await request("/auth/register", method: "POST", body: req, requiresAuth: false)
        persistSession(auth)
        return auth
    }

    public func getProfile() async throws -> User {
        let user: User = try await request("/auth/profile", method: "GET")
        persistUser(user)
        return user
    }

    public func refreshToken() async -> Bool {
        guard let refresh = KeychainManager.shared.get(for: "refresh_token"), !refresh.isEmpty else {
            return false
        }
        do {
            let body = RefreshTokenRequest(refreshToken: refresh)
            let tokens: TokenPair = try await request(
                "/auth/refresh",
                method: "POST",
                body: body,
                requiresAuth: false,
                retryOn401: false
            )
            KeychainManager.shared.save(tokens.accessToken, for: "access_token")
            KeychainManager.shared.save(tokens.refreshToken, for: "refresh_token")
            return true
        } catch {
            AppLogger.warn("Token refresh failed", category: .auth, metadata: ["error": error.localizedDescription])
            return false
        }
    }

    public func logout() async {
        _ = try? await request("/auth/logout", method: "POST", retryOn401: false) as EmptyResponse?
        KeychainManager.shared.clearAll()
    }

    // MARK: - Hub Endpoints

    public func fetchDailyBriefing() async throws -> DailyBriefingResponse {
        try await request("/hub/briefing", method: "GET")
    }

    public func fetchSummaries() async throws -> [AppSummaryResponse] {
        (try await requestOptional("/hub/summary", method: "GET")) ?? []
    }

    public func fetchLinkedApps() async throws -> [LinkedAppItem] {
        (try await requestOptional("/hub/apps", method: "GET")) ?? []
    }

    public func initiateLink(appId: String) async throws -> InitiateLinkResponse {
        try await request("/hub/link/initiate", method: "POST", body: InitiateLinkRequest(appId: appId))
    }

    public func unlinkApp(appId: String) async throws {
        let _: EmptyResponse? = try await request("/hub/apps/\(appId)", method: "DELETE")
    }

    public func sendChat(message: String) async throws -> AIChatResponse {
        try await request("/hub/chat", method: "POST", body: AIChatRequest(message: message))
    }

    // MARK: - Session Persistence

    private func persistSession(_ auth: AuthResponse) {
        KeychainManager.shared.save(auth.tokens.accessToken, for: "access_token")
        KeychainManager.shared.save(auth.tokens.refreshToken, for: "refresh_token")
        persistUser(auth.user)
    }

    private func persistUser(_ user: User) {
        KeychainManager.shared.save(user.id, for: "user_id")
        KeychainManager.shared.save(user.name, for: "user_name")
        KeychainManager.shared.save(user.email, for: "user_email")
        if let avatar = user.avatarUrl, !avatar.isEmpty {
            KeychainManager.shared.save(avatar, for: "user_avatar")
        }
        if let provider = user.authProvider, !provider.isEmpty {
            KeychainManager.shared.save(provider, for: "auth_provider")
        }
    }

    // MARK: - Core HTTP Pipeline

    private func requestOptional<T: Decodable>(
        _ path: String,
        method: String = "GET",
        body: Encodable? = nil,
        requiresAuth: Bool = true,
        retryOn401: Bool = true
    ) async throws -> T? {
        do {
            return try await request(path, method: method, body: body, requiresAuth: requiresAuth, retryOn401: retryOn401) as T
        } catch APIError.decodingError(let msg) where msg == "empty_data" {
            return nil
        }
    }

    private func request<T: Decodable>(
        _ path: String,
        method: String = "GET",
        body: Encodable? = nil,
        requiresAuth: Bool = true,
        retryOn401: Bool = true
    ) async throws -> T {
        let cleanBase = AppConfig.baseURL.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        let cleanPath = path.hasPrefix("/") ? path : "/\(path)"
        guard let url = URL(string: cleanBase + cleanPath) else {
            throw APIError.invalidURL
        }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = method
        urlRequest.setValue("application/json", forHTTPHeaderField: "Accept")
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if requiresAuth, let token = KeychainManager.shared.get(for: "access_token"), !token.isEmpty {
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        if let body = body {
            urlRequest.httpBody = try encoder.encode(AnyEncodable(body))
        }

        NetworkLogger.logRequest(urlRequest)

        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await session.data(for: urlRequest)
        } catch {
            throw APIError.networkError(error.localizedDescription)
        }

        NetworkLogger.logResponse(response, data: data)

        guard let http = response as? HTTPURLResponse else {
            throw APIError.networkError("Respons HTTP tidak valid.")
        }

        if http.statusCode == 401 && requiresAuth && retryOn401 {
            if await refreshToken() {
                return try await request(path, method: method, body: body, requiresAuth: requiresAuth, retryOn401: false)
            } else {
                KeychainManager.shared.clearAll()
                await MainActor.run {
                    NotificationCenter.default.post(name: .sessionExpired, object: nil)
                }
                let serverMsg = extractErrorMessage(from: data) ?? "Sesi berakhir, silakan login ulang."
                throw APIError.unauthorized(serverMsg)
            }
        }

        guard (200...299).contains(http.statusCode) else {
            let msg = extractErrorMessage(from: data) ?? "Terjadi kesalahan pada server (\(http.statusCode))."
            if http.statusCode == 401 {
                throw APIError.unauthorized(msg)
            }
            throw APIError.serverError(statusCode: http.statusCode, message: msg)
        }

        if T.self == EmptyResponse?.self || T.self == EmptyResponse.self {
            if let empty = EmptyResponse() as? T {
                return empty
            }
        }

        do {
            let wrapped = try decoder.decode(BaseResponse<T>.self, from: data)
            if let payload = wrapped.data {
                return payload
            }
            if !wrapped.success {
                throw APIError.serverError(statusCode: http.statusCode, message: wrapped.message)
            }
            throw APIError.decodingError("empty_data")
        } catch let apiErr as APIError {
            throw apiErr
        } catch {
            throw APIError.decodingError(error.localizedDescription)
        }
    }

    private func extractErrorMessage(from data: Data) -> String? {
        struct ErrorEnvelope: Decodable {
            let message: String?
            let error: String?
        }
        if let decoded = try? JSONDecoder().decode(ErrorEnvelope.self, from: data) {
            return decoded.message ?? decoded.error
        }
        return nil
    }
}

private struct AnyEncodable: Encodable {
    private let encodeClosure: (Encoder) throws -> Void

    init(_ encodable: Encodable) {
        self.encodeClosure = { encoder in
            try encodable.encode(to: encoder)
        }
    }

    func encode(to encoder: Encoder) throws {
        try encodeClosure(encoder)
    }
}
