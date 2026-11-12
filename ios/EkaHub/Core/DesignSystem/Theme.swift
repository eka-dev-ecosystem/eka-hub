import SwiftUI

public enum AppColors {
    // Layered Gemini-inspired deep cosmic palette
    public static let backgroundBase = Color(red: 0.024, green: 0.035, blue: 0.075)      // #060913
    public static let surfaceElevated = Color(red: 0.067, green: 0.090, blue: 0.165)     // #11172A
    public static let surfaceCard = Color(red: 0.086, green: 0.114, blue: 0.200)         // #161D33

    // Accent & Gemini Aurora Spectrum
    public static let geminiBlue = Color(red: 0.259, green: 0.522, blue: 0.957)          // #4285F4
    public static let geminiPurple = Color(red: 0.608, green: 0.447, blue: 0.969)        // #9B72F7
    public static let geminiCyan = Color(red: 0.133, green: 0.827, blue: 0.933)          // #22D3EE
    public static let geminiRose = Color(red: 0.851, green: 0.396, blue: 0.439)          // #D96570

    public static let primary = geminiBlue
    public static let accentBlue = geminiBlue
    public static let accentCyan = geminiCyan
    public static let accentPurple = geminiPurple
    public static let accentEmerald = Color(red: 0.063, green: 0.725, blue: 0.506)       // #10B981
    public static let accentAmber = Color(red: 0.961, green: 0.620, blue: 0.043)         // #F59E0B
    public static let accentRose = Color(red: 0.957, green: 0.247, blue: 0.369)          // #F43F5E

    public static let textPrimary = Color.white
    public static let textSecondary = Color(red: 0.612, green: 0.663, blue: 0.788)       // #9CA9C9
    public static let textMuted = Color(red: 0.420, green: 0.475, blue: 0.600)           // #6B7999
    public static let borderSubtle = Color.white.opacity(0.11)

    public static var geminiGradient: LinearGradient {
        LinearGradient(
            colors: [geminiBlue, geminiPurple, geminiRose],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    public static var auroraButtonGradient: LinearGradient {
        LinearGradient(
            colors: [geminiBlue, geminiPurple],
            startPoint: .leading,
            endPoint: .trailing
        )
    }

    public static var cardSurfaceGradient: LinearGradient {
        LinearGradient(
            colors: [
                Color.white.opacity(0.085),
                Color.white.opacity(0.035)
            ],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }
}
