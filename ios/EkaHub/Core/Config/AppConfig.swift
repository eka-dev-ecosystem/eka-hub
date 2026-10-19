import Foundation

public enum EnvReader {
    private static var cachedValues: [String: String]?

    public static func value(for key: String) -> String? {
        if let envVal = ProcessInfo.processInfo.environment[key], !envVal.isEmpty {
            return envVal
        }
        if let infoVal = Bundle.main.object(forInfoDictionaryKey: key) as? String, !infoVal.isEmpty {
            return infoVal
        }
        if cachedValues == nil {
            cachedValues = loadDotEnv()
        }
        return cachedValues?[key]
    }

    private static func loadDotEnv() -> [String: String] {
        var dict: [String: String] = [:]
        guard let url = Bundle.main.url(forResource: ".env", withExtension: nil) ??
                        Bundle.main.url(forResource: "env", withExtension: nil),
              let content = try? String(contentsOf: url, encoding: .utf8) else {
            return dict
        }
        for line in content.components(separatedBy: .newlines) {
            let trimmed = line.trimmingCharacters(in: .whitespacesAndNewlines)
            if trimmed.isEmpty || trimmed.hasPrefix("#") { continue }
            let parts = trimmed.split(separator: "=", maxSplits: 1).map(String.init)
            if parts.count == 2 {
                let k = parts[0].trimmingCharacters(in: .whitespaces)
                var v = parts[1].trimmingCharacters(in: .whitespaces)
                if (v.hasPrefix("\"") && v.hasSuffix("\"")) || (v.hasPrefix("'") && v.hasSuffix("'")) {
                    v = String(v.dropFirst().dropLast())
                }
                dict[k] = v
            }
        }
        return dict
    }
}

public enum AppEnvironment: String, CaseIterable, Sendable {
    case development = "Development"
    case staging = "Staging"
    case production = "Production"

    public static var current: AppEnvironment {
        if let envStr = EnvReader.value(for: "ENVIRONMENT")?.lowercased() {
            switch envStr {
            case "staging": return .staging
            case "production", "prod": return .production
            default: return .development
            }
        }
        #if DEBUG
        return .development
        #else
        return .production
        #endif
    }

    public static let bundleIdentifier = "cloud.eka-dev.ekahub"

    public var baseURL: String {
        if let custom = EnvReader.value(for: "API_BASE_URL"), !custom.isEmpty {
            return custom
        }
        return "https://be-hub.eka-dev.cloud/api/v1"
    }

    public var sseURL: String {
        return "\(baseURL)/realtime/sse/events"
    }

    public var googleClientID: String {
        if let custom = EnvReader.value(for: "GOOGLE_CLIENT_ID"), !custom.isEmpty {
            return custom
        }
        return "897905079551-86q6g4ov6b966r2t7sbmqsmks5bea682.apps.googleusercontent.com"
    }
}

public struct AppConfig: Sendable {
    public static var environment: AppEnvironment {
        return AppEnvironment.current
    }

    public static var baseURL: String {
        return environment.baseURL
    }

    public static var sseURL: String {
        return environment.sseURL
    }

    public static var googleClientID: String {
        return environment.googleClientID
    }
}
