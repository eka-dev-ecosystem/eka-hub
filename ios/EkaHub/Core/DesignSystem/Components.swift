import SwiftUI

public struct BentoCardView<Content: View>: View {
    let content: Content

    public init(@ViewBuilder content: () -> Content) {
        self.content = content()
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            content
        }
        .padding(14)
        .glassCardStyle()
    }
}

public struct GradientAiBannerView: View {
    let title: String
    let contentText: String

    public init(title: String = "Daily AI Briefing", contentText: String) {
        self.title = title
        self.contentText = contentText
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text(title)
                    .font(.caption)
                    .fontWeight(.bold)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 4)
                    .background(AppColors.primary)
                    .foregroundColor(.white)
                    .cornerRadius(8)

                Spacer()

                Text("Live Sync")
                    .font(.caption2)
                    .foregroundColor(AppColors.cyan)
            }

            Text(contentText)
                .font(.system(size: 14))
                .foregroundColor(.white)
                .lineSpacing(4)
        }
        .padding(18)
        .background(
            LinearGradient(
                colors: [Color(red: 0.19, green: 0.18, blue: 0.51), Color(red: 0.12, green: 0.11, blue: 0.29)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .cornerRadius(18)
        .overlay(
            RoundedRectangle(cornerRadius: 18)
                .stroke(AppColors.primary.opacity(0.6), lineWidth: 1)
        )
    }
}
