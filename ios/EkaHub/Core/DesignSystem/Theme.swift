import SwiftUI

public enum AppColors {
    public static let primary = Color(red: 0.39, green: 0.40, blue: 0.95) // Indigo
    public static let indigoDark = Color(red: 0.31, green: 0.27, blue: 0.90)
    public static let cyan = Color(red: 0.02, green: 0.71, blue: 0.83)
    public static let emerald = Color(red: 0.06, green: 0.73, blue: 0.51)
    public static let amber = Color(red: 0.96, green: 0.62, blue: 0.04)
    public static let darkBackground = Color(red: 0.06, green: 0.09, blue: 0.16)
    public static let darkSurface = Color(red: 0.12, green: 0.16, blue: 0.23)
    public static let darkBorder = Color(red: 0.20, green: 0.25, blue: 0.33)
}

public struct GlassCardModifier: ViewModifier {
    public func body(content: Content) -> some View {
        content
            .background(AppColors.darkSurface)
            .cornerRadius(16)
            .overlay(
                RoundedRectangle(cornerRadius: 16)
                    .stroke(AppColors.darkBorder, lineWidth: 1)
            )
    }
}

public extension View {
    func glassCardStyle() -> some View {
        modifier(GlassCardModifier())
    }
}
