import Foundation

public actor APIClient {
    public static let shared = APIClient()

    private let session = URLSession.shared
    private let decoder = JSONDecoder()

    private init() {}

    public func fetchSummaries() async throws -> [AppSummaryResponse] {
        guard let url = URL(string: "\(AppConfig.baseURL)/hub/summaries") else {
            return []
        }
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        if let token = KeychainManager.shared.get(for: "access_token") {
            request.addValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        NetworkLogger.logRequest(request)
        let (data, response) = try await session.data(for: request)
        NetworkLogger.logResponse(response, data: data)

        let decoded = try decoder.decode([AppSummaryResponse].self, from: data)
        return decoded
    }
}
