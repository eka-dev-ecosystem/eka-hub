import Foundation

public final class NetworkLogger: Sendable {
    public static func logRequest(_ request: URLRequest) {
        #if DEBUG
        let url = request.url?.absoluteString ?? "Unknown URL"
        let method = request.httpMethod ?? "GET"
        AppLogger.debug("🌐 [REQ] \(method) \(url)", category: .network)
        #endif
    }

    public static func logResponse(_ response: URLResponse?, data: Data?) {
        #if DEBUG
        if let http = response as? HTTPURLResponse {
            let status = http.statusCode
            AppLogger.debug("📥 [RES] Status: \(status)", category: .network)
        }
        #endif
    }
}
