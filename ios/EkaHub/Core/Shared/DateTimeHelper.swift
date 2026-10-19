import Foundation

public enum DateTimeHelper {
    public static func formattedToday() -> String {
        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        return formatter.string(from: Date())
    }
}
