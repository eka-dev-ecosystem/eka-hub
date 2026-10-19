import Foundation

public final class SSEClient: Sendable {
    public static let shared = SSEClient()
    private init() {}

    public func streamEvents() -> AsyncThrowingStream<String, Error> {
        AsyncThrowingStream { continuation in
            guard let url = URL(string: AppConfig.sseURL) else {
                continuation.finish()
                return
            }

            let task = Task {
                do {
                    var request = URLRequest(url: url)
                    request.timeoutInterval = 3600
                    let (bytes, response) = try await URLSession.shared.bytes(for: request)
                    guard (response as? HTTPURLResponse)?.statusCode == 200 else {
                        continuation.finish()
                        return
                    }

                    for try await line in bytes.lines {
                        if line.hasPrefix("data:") {
                            let data = String(line.dropFirst(5)).trimmingCharacters(in: .whitespaces)
                            continuation.yield(data)
                        }
                    }
                    continuation.finish()
                } catch {
                    continuation.finish(throwing: error)
                }
            }

            continuation.onTermination = { _ in
                task.cancel()
            }
        }
    }
}
