import SwiftUI
#if canImport(UIKit)
import UIKit
#endif

public struct SplashView: View {
    public let onFinished: () -> Void

    @State private var isEntranceReady = false
    @State private var isPulsing = false
    @State private var shimmerOffset: CGFloat = -50

    public init(onFinished: @escaping () -> Void) {
        self.onFinished = onFinished
    }

    public var body: some View {
        ZStack {
            HubBackgroundView()

            VStack(spacing: 0) {
                Spacer()

                // Center Emblem & Brand Aura
                VStack(spacing: 24) {
                    ZStack {
                        // Ambient Breathing Aura
                        Circle()
                            .fill(AppColors.aiGradient)
                            .frame(width: 130, height: 130)
                            .blur(radius: 30)
                            .opacity(isPulsing ? 0.45 : 0.2)
                            .scaleEffect(isPulsing ? 1.12 : 0.95)

                        // Expanding Concentric Rings
                        Circle()
                            .stroke(
                                AppColors.aiGradient.opacity(isPulsing ? 0.35 : 0.15),
                                lineWidth: 1
                            )
                            .frame(width: 122, height: 122)
                            .scaleEffect(isPulsing ? 1.08 : 0.98)

                        // Main Emblem Container
                        RoundedRectangle(cornerRadius: 26, style: .continuous)
                            .fill(AppColors.surfaceElevated)
                            .frame(width: 82, height: 82)
                            .overlay(
                                RoundedRectangle(cornerRadius: 26, style: .continuous)
                                    .stroke(AppColors.aiGradient, lineWidth: 1.5)
                            )
                            .shadow(color: AppColors.aiBlue.opacity(0.4), radius: 20, x: 0, y: 8)

                        // Centered Animated Sparkle
                        HubSparkleView(size: 38, animate: true)
                    }
                    .scaleEffect(isEntranceReady ? 1.0 : 0.4)
                    .opacity(isEntranceReady ? 1.0 : 0.0)

                    // Brand Typography
                    VStack(spacing: 8) {
                        Text("Personal AI Hub")
                            .font(.system(size: 30, weight: .bold, design: .rounded))
                            .foregroundStyle(
                                LinearGradient(
                                    colors: [.white, AppColors.textSecondary],
                                    startPoint: .top,
                                    endPoint: .bottom
                                )
                            )

                        // Modern Pill Tagline
                        HStack(spacing: 6) {
                            Text("UNIFIED AI ECOSYSTEM")
                                .font(.system(size: 11, weight: .semibold))
                                .foregroundColor(AppColors.textSecondary)
                                .tracking(1.4)
                        }
                        .padding(.horizontal, 14)
                        .padding(.vertical, 6)
                        .glassCard(cornerRadius: 20)
                    }
                    .offset(y: isEntranceReady ? 0 : 16)
                    .opacity(isEntranceReady ? 1.0 : 0.0)
                }

                Spacer()

                // Fluid Shimmer Capsule Loader at Bottom
                VStack(spacing: 12) {
                    ZStack(alignment: .leading) {
                        // Capsule track
                        Capsule()
                            .fill(AppColors.surfaceCard.opacity(0.8))
                            .frame(width: 140, height: 3.5)

                        // Silky smooth continuous shimmer progress beam
                        Capsule()
                            .fill(
                                LinearGradient(
                                    colors: [
                                        Color.clear,
                                        AppColors.aiCyan,
                                        AppColors.aiBlue,
                                        Color.clear
                                    ],
                                    startPoint: .leading,
                                    endPoint: .trailing
                                )
                            )
                            .frame(width: 50, height: 3.5)
                            .offset(x: shimmerOffset)
                            .shadow(color: AppColors.aiCyan.opacity(0.8), radius: 6)
                    }
                    .frame(width: 140, height: 3.5)
                    .clipShape(Capsule())

                    Text("INITIALIZING ECOSYSTEM...")
                        .font(.system(size: 10, weight: .bold))
                        .foregroundColor(AppColors.textMuted)
                        .tracking(2.0)
                }
                .padding(.bottom, 48)
                .opacity(isEntranceReady ? 1.0 : 0.0)
            }
        }
        .task {
            #if canImport(UIKit)
            let generator = UIImpactFeedbackGenerator(style: .medium)
            generator.prepare()
            generator.impactOccurred()
            #endif

            withAnimation(.spring(response: 0.7, dampingFraction: 0.65)) {
                isEntranceReady = true
            }

            withAnimation(.easeInOut(duration: 1.8).repeatForever(autoreverses: true)) {
                isPulsing = true
            }

            withAnimation(.linear(duration: 1.5).repeatForever(autoreverses: false)) {
                shimmerOffset = 140
            }

            // Minimum display duration for silky smooth splash animation
            try? await Task.sleep(nanoseconds: 1_400_000_000)
            onFinished()
        }
    }
}
