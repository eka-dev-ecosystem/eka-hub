import Foundation
import os.log

public enum LogLevel: Int, Comparable, Sendable {
    case debug = 0
    case info = 1
    case success = 2
    case warning = 3
    case error = 4

    public static func < (lhs: LogLevel, rhs: LogLevel) -> Bool {
        lhs.rawValue < rhs.rawValue
    }

    public var title: String {
        switch self {
        case .debug:   return "DEBUG"
        case .info:    return "INFO"
        case .success: return "SUCCESS"
        case .warning: return "WARN"
        case .error:   return "ERROR"
        }
    }

    public var icon: String {
        switch self {
        case .debug:   return "🔍"
        case .info:    return "ℹ️"
        case .success: return "✅"
        case .warning: return "⚠️"
        case .error:   return "❌"
        }
    }

    public var badge: String {
        "\(icon) [\(title)]"
    }

    public var osLogType: OSLogType {
        switch self {
        case .debug:   return .debug
        case .info:    return .info
        case .success: return .default
        case .warning: return .default
        case .error:   return .error
        }
    }
}

public struct LogCategory: Sendable, RawRepresentable, ExpressibleByStringLiteral, Hashable {
    public let rawValue: String

    public init(rawValue: String) {
        self.rawValue = rawValue
    }

    public init(stringLiteral value: String) {
        self.rawValue = value
    }

    public static let app: LogCategory = "APP"
    public static let auth: LogCategory = "AUTH"
    public static let network: LogCategory = "NETWORK"
    public static let hub: LogCategory = "HUB"
    public static let deeplink: LogCategory = "DEEPLINK"
}

public final class AppLogger: Sendable {
    public static let shared = AppLogger()

    #if DEBUG
    public static var minimumLevel: LogLevel = .debug
    #else
    public static var minimumLevel: LogLevel = .info
    #endif

    private static let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "HH:mm:ss.SSS"
        return formatter
    }()

    private static let subsystem = Bundle.main.bundleIdentifier ?? "cloud.eka-dev.ekahub"

    private init() {}

    public static func debug(
        _ message: String,
        category: LogCategory = .app,
        metadata: [String: Any]? = nil,
        file: String = #file,
        line: Int = #line
    ) {
        log(level: .debug, message: message, category: category, error: nil, metadata: metadata, file: file, line: line)
    }

    public static func info(
        _ message: String,
        category: LogCategory = .app,
        metadata: [String: Any]? = nil,
        file: String = #file,
        line: Int = #line
    ) {
        log(level: .info, message: message, category: category, error: nil, metadata: metadata, file: file, line: line)
    }

    public static func success(
        _ message: String,
        category: LogCategory = .app,
        metadata: [String: Any]? = nil,
        file: String = #file,
        line: Int = #line
    ) {
        log(level: .success, message: message, category: category, error: nil, metadata: metadata, file: file, line: line)
    }

    public static func warn(
        _ message: String,
        category: LogCategory = .app,
        metadata: [String: Any]? = nil,
        file: String = #file,
        line: Int = #line
    ) {
        log(level: .warning, message: message, category: category, error: nil, metadata: metadata, file: file, line: line)
    }

    public static func error(
        _ message: String,
        error: Error? = nil,
        category: LogCategory = .app,
        metadata: [String: Any]? = nil,
        file: String = #file,
        line: Int = #line
    ) {
        log(level: .error, message: message, category: category, error: error, metadata: metadata, file: file, line: line)
    }

    private static func log(
        level: LogLevel,
        message: String,
        category: LogCategory,
        error: Error?,
        metadata: [String: Any]?,
        file: String,
        line: Int
    ) {
        guard level >= minimumLevel else { return }

        let fileName = (file as NSString).lastPathComponent
        let timestamp = dateFormatter.string(from: Date())

        var fullMessage = message
        if let error = error {
            fullMessage += " | Error: \(error.localizedDescription)"
        }

        var metaString = ""
        if let metadata = metadata, !metadata.isEmpty {
            let pairs = metadata.map { "\($0.key)=\($0.value)" }.joined(separator: ", ")
            metaString = " {\(pairs)}"
        }

        let formattedString = "[\(timestamp)] \(level.badge) [\(category.rawValue)] (\(fileName):\(line)) › \(fullMessage)\(metaString)"

        #if DEBUG
        print(formattedString)
        #endif

        let osLog = OSLog(subsystem: subsystem, category: category.rawValue)
        os_log("%{public}@", log: osLog, type: level.osLogType, formattedString)
    }
}
