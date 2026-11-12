import SwiftUI

public struct HubBackgroundView: View {
    public init() {}

    public var body: some View {
        ZStack {
            AppColors.backgroundBase
                .ignoresSafeArea()

            // Top ambient AI aurora orb
            RadialGradient(
                gradient: Gradient(colors: [
                    AppColors.aiBlue.opacity(0.22),
                    AppColors.aiPurple.opacity(0.14),
                    Color.clear
                ]),
                center: UnitPoint(x: 0.2, y: 0.0),
                startRadius: 20,
                endRadius: 420
            )
            .ignoresSafeArea()

            // Secondary cyan glow right
            RadialGradient(
                gradient: Gradient(colors: [
                    AppColors.aiCyan.opacity(0.12),
                    Color.clear
                ]),
                center: UnitPoint(x: 0.9, y: 0.15),
                startRadius: 10,
                endRadius: 340
            )
            .ignoresSafeArea()

            // Bottom subtle violet depth
            RadialGradient(
                gradient: Gradient(colors: [
                    AppColors.aiPurple.opacity(0.10),
                    Color.clear
                ]),
                center: UnitPoint(x: 0.5, y: 1.0),
                startRadius: 10,
                endRadius: 380
            )
            .ignoresSafeArea()
        }
    }
}
public typealias GeminiBackgroundView = HubBackgroundView

public struct HubSparkleView: View {
    public let size: CGFloat
    public let animate: Bool
    @State private var isGlowing = false

    public init(size: CGFloat = 28, animate: Bool = false) {
        self.size = size
        self.animate = animate
    }

    public var body: some View {
        ZStack {
            Circle()
                .fill(AppColors.aiGradient)
                .frame(width: size * 1.45, height: size * 1.45)
                .blur(radius: size * 0.45)
                .opacity(animate ? (isGlowing ? 0.55 : 0.25) : 0.35)

            Image(systemName: "sparkles")
                .font(.system(size: size, weight: .semibold))
                .foregroundStyle(AppColors.aiGradient)
                .scaleEffect(animate && isGlowing ? 1.08 : 1.0)
        }
        .onAppear {
            if animate {
                withAnimation(.easeInOut(duration: 1.8).repeatForever(autoreverses: true)) {
                    isGlowing = true
                }
            }
        }
    }
}
public typealias GeminiSparkleView = HubSparkleView

public struct GlassCardModifier: ViewModifier {
    public var cornerRadius: CGFloat

    public init(cornerRadius: CGFloat = 20) {
        self.cornerRadius = cornerRadius
    }

    public func body(content: Content) -> some View {
        content
            .background(
                RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
                    .fill(AppColors.surfaceElevated.opacity(0.82))
                    .overlay(
                        RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
                            .fill(AppColors.cardSurfaceGradient)
                    )
            )
            .overlay(
                RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
                    .stroke(
                        LinearGradient(
                            colors: [
                                Color.white.opacity(0.18),
                                Color.white.opacity(0.05)
                            ],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ),
                        lineWidth: 1
                    )
            )
            .shadow(color: Color.black.opacity(0.32), radius: 14, x: 0, y: 8)
    }
}

public extension View {
    func glassCard(cornerRadius: CGFloat = 20) -> some View {
        modifier(GlassCardModifier(cornerRadius: cornerRadius))
    }
}

public struct BentoCardView: View {
    public let title: String
    public let metric: String
    public let subtitle: String
    public let iconName: String
    public let accentColor: Color
    public let action: () -> Void

    public init(
        title: String,
        metric: String,
        subtitle: String,
        iconName: String,
        accentColor: Color,
        action: @escaping () -> Void
    ) {
        self.title = title
        self.metric = metric
        self.subtitle = subtitle
        self.iconName = iconName
        self.accentColor = accentColor
        self.action = action
    }

    public var body: some View {
        Button(action: action) {
            VStack(alignment: .leading, spacing: 14) {
                HStack {
                    ZStack {
                        RoundedRectangle(cornerRadius: 11, style: .continuous)
                            .fill(accentColor.opacity(0.18))
                            .frame(width: 36, height: 36)
                        Image(systemName: iconName)
                            .font(.system(size: 16, weight: .semibold))
                            .foregroundColor(accentColor)
                    }

                    Spacer()

                    Image(systemName: "arrow.up.right")
                        .font(.system(size: 12, weight: .bold))
                        .foregroundColor(AppColors.textSecondary)
                        .padding(6)
                        .background(Color.white.opacity(0.06))
                        .clipShape(Circle())
                }

                VStack(alignment: .leading, spacing: 4) {
                    Text(metric)
                        .font(.system(size: 20, weight: .bold, design: .rounded))
                        .foregroundColor(AppColors.textPrimary)
                        .lineLimit(1)
                        .minimumScaleFactor(0.75)

                    Text(subtitle)
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(AppColors.textSecondary)
                        .lineLimit(2)
                }

                Text(title)
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundColor(accentColor)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 3)
                    .background(accentColor.opacity(0.14))
                    .clipShape(Capsule())
            }
            .padding(16)
            .frame(maxWidth: .infinity, alignment: .leading)
            .glassCard(cornerRadius: 20)
        }
        .buttonStyle(.plain)
    }
}

public struct GradientAiBannerView: View {
    public let briefingText: String
    public let onAskAiTapped: () -> Void

    public init(briefingText: String, onAskAiTapped: @escaping () -> Void) {
        self.briefingText = briefingText
        self.onAskAiTapped = onAskAiTapped
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(spacing: 10) {
                GeminiSparkleView(size: 18, animate: false)
                Text("Daily AI Briefing")
                    .font(.system(size: 13, weight: .bold))
                    .foregroundStyle(AppColors.geminiGradient)
                Spacer()
                Text("AI Hub")
                    .font(.system(size: 10, weight: .bold))
                    .foregroundColor(AppColors.geminiCyan)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 3)
                    .background(AppColors.geminiCyan.opacity(0.14))
                    .clipShape(Capsule())
            }

            Text(LocalizedStringKey(briefingText))
                .font(.system(size: 14, weight: .regular))
                .foregroundColor(AppColors.textPrimary.opacity(0.92))
                .lineSpacing(4)

            Button(action: onAskAiTapped) {
                HStack(spacing: 8) {
                    Image(systemName: "wand.and.stars")
                        .font(.system(size: 13, weight: .semibold))
                    Text("Tanyakan ke Asisten AI")
                        .font(.system(size: 13, weight: .semibold))
                    Spacer()
                    Image(systemName: "chevron.right")
                        .font(.system(size: 11, weight: .bold))
                }
                .foregroundColor(.white)
                .padding(.horizontal, 14)
                .padding(.vertical, 10)
                .background(AppColors.auroraButtonGradient)
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                .shadow(color: AppColors.geminiBlue.opacity(0.35), radius: 8, x: 0, y: 4)
            }
            .buttonStyle(.plain)
        }
        .padding(18)
        .glassCard(cornerRadius: 22)
    }
}
