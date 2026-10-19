import Foundation

public final class PushNotificationManager: @unchecked Sendable {
    public static let shared = PushNotificationManager()
    private init() {}

    public func configure() {
        AppLogger.info("Configured push notifications for EkaHub", category: .app)
    }

    public func handleDeviceTokenData(_ tokenData: Data) {
        let token = tokenData.map { String(format: "%02.2hhx", $0) }.joined()
        AppLogger.info("Received device token: \(token)", category: .app)
    }
}
