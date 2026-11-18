import SwiftUI
#if canImport(UIKit)
import UIKit
#endif

public struct AiChatView: View {
    @State private var viewModel = AiChatViewModel()
    var authViewModel: AuthViewModel
    @Binding var incomingPrompt: String?
    var onOpenSatellites: () -> Void
    var onExecuteDeepLink: (URL) -> Void

    @FocusState private var isInputFocused: Bool

    public init(
        authViewModel: AuthViewModel,
        incomingPrompt: Binding<String?> = .constant(nil),
        onOpenSatellites: @escaping () -> Void = {},
        onExecuteDeepLink: @escaping (URL) -> Void
    ) {
        self.authViewModel = authViewModel
        self._incomingPrompt = incomingPrompt
        self.onOpenSatellites = onOpenSatellites
        self.onExecuteDeepLink = onExecuteDeepLink
    }

    private var firstName: String {
        let name = authViewModel.currentUser?.name ?? "Kamu"
        return name.split(separator: " ").first.map(String.init) ?? name
    }

    public var body: some View {
        ZStack {
            GeminiBackgroundView()

            VStack(spacing: 0) {
                geminiTopBar

                ScrollViewReader { proxy in
                    ScrollView(showsIndicators: false) {
                        if viewModel.messages.isEmpty {
                            welcomeHeroView
                                .padding(.horizontal, 20)
                                .padding(.top, 24)
                                .padding(.bottom, 20)
                        } else {
                            LazyVStack(spacing: 20) {
                                ForEach(viewModel.messages) { msg in
                                    GeminiMessageRow(
                                        message: msg,
                                        onExecuteDeepLink: onExecuteDeepLink,
                                        onSendQuickReply: { reply in
                                            Task {
                                                await viewModel.sendQuickPrompt(reply)
                                            }
                                        },
                                        onRetryMessage: { id in
                                            Task {
                                                await viewModel.retryMessage(id: id)
                                            }
                                        },
                                        onRestorePrompt: { text in
                                            viewModel.restorePrompt(text)
                                            isInputFocused = true
                                        }
                                    )
                                    .id(msg.id)
                                }

                                if viewModel.isSending {
                                    geminiThinkingRow
                                        .id("thinking-row")
                                }
                            }
                            .padding(.horizontal, 16)
                            .padding(.top, 14)
                            .padding(.bottom, 24)
                        }
                    }
                    .scrollDismissesKeyboard(.interactively)
                    .onChange(of: viewModel.messages.count) { _, _ in
                        if let lastId = viewModel.messages.last?.id {
                            withAnimation(.easeOut(duration: 0.25)) {
                                proxy.scrollTo(lastId, anchor: .bottom)
                            }
                        }
                    }
                    .onChange(of: viewModel.isSending) { _, sending in
                        if sending {
                            withAnimation(.easeOut(duration: 0.25)) {
                                proxy.scrollTo("thinking-row", anchor: .bottom)
                            }
                        }
                    }
                }

                geminiComposerBar
            }
        }
        .task {
            await viewModel.loadLinkedApps()
        }
        .onChange(of: incomingPrompt) { _, newPrompt in
            guard let prompt = newPrompt, !prompt.isEmpty else { return }
            incomingPrompt = nil
            Task {
                await viewModel.sendQuickPrompt(prompt)
            }
        }
        .onReceive(NotificationCenter.default.publisher(for: .linkedAppsDidChange)) { _ in
            Task {
                await viewModel.loadLinkedApps()
            }
        }
        .sheet(isPresented: $viewModel.showHistorySheet) {
            ChatHistorySheetView(
                sessions: viewModel.sessions,
                currentSessionId: viewModel.currentSessionId,
                isLoading: viewModel.isHistoryLoading,
                onSelectSession: { id in
                    Task {
                        await viewModel.selectSession(id: id)
                    }
                },
                onNewChat: {
                    viewModel.startNewChat()
                },
                onDeleteSession: { id in
                    Task {
                        await viewModel.deleteSession(id: id)
                    }
                }
            )
        }
    }

    // MARK: - Top Navigation Header

    private var geminiTopBar: some View {
        HStack(spacing: 10) {
            Button(action: {
                Task {
                    await viewModel.loadSessions()
                    viewModel.showHistorySheet = true
                }
            }) {
                Image(systemName: "clock.arrow.circlepath")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundColor(AppColors.textPrimary)
                    .frame(width: 34, height: 34)
                    .background(Color.white.opacity(0.08))
                    .clipShape(Circle())
            }
            .buttonStyle(.plain)

            HStack(spacing: 8) {
                GeminiSparkleView(size: 18, animate: false)
                VStack(alignment: .leading, spacing: 1) {
                    HStack(spacing: 6) {
                        Text("Eka-Hub")
                            .font(.system(size: 17, weight: .bold, design: .rounded))
                            .foregroundColor(AppColors.textPrimary)

                        Text("AI")
                            .font(.system(size: 10, weight: .bold))
                            .foregroundStyle(AppColors.geminiGradient)
                            .padding(.horizontal, 7)
                            .padding(.vertical, 2)
                            .background(Color.white.opacity(0.08))
                            .clipShape(Capsule())
                    }

                    Text(satelliteStatusSubtitle)
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(AppColors.textSecondary)
                }
            }

            Spacer()

            Button(action: onOpenSatellites) {
                HStack(spacing: 5) {
                    Circle()
                        .fill(viewModel.linkedApps.isEmpty ? AppColors.accentAmber : AppColors.accentEmerald)
                        .frame(width: 7, height: 7)
                    Text("\(viewModel.linkedApps.count) Satelit")
                        .font(.system(size: 11, weight: .semibold))
                        .foregroundColor(AppColors.textPrimary)
                }
                .padding(.horizontal, 10)
                .padding(.vertical, 6)
                .background(Color.white.opacity(0.08))
                .clipShape(Capsule())
            }
            .buttonStyle(.plain)

            Button(action: {
                withAnimation {
                    viewModel.startNewChat()
                }
            }) {
                Image(systemName: "square.and.pencil")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundColor(AppColors.textPrimary)
                    .frame(width: 34, height: 34)
                    .background(Color.white.opacity(0.08))
                    .clipShape(Circle())
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 10)
        .background(AppColors.backgroundBase.opacity(0.65))
    }

    private var satelliteStatusSubtitle: String {
        if viewModel.linkedApps.isEmpty {
            return "Siap membantu • Ketuk Satelit untuk hubungkan app"
        }
        let names = viewModel.linkedApps.map { $0.appName }.joined(separator: " • ")
        return "Terhubung: \(names)"
    }

    // MARK: - Gemini Welcome Hero

    private var welcomeHeroView: some View {
        VStack(alignment: .leading, spacing: 26) {
            VStack(alignment: .leading, spacing: 12) {
                GeminiSparkleView(size: 36, animate: true)
                    .padding(.bottom, 4)

                Text("Halo, \(firstName)")
                    .font(.system(size: 32, weight: .bold, design: .rounded))
                    .foregroundStyle(AppColors.geminiGradient)

                Text("Ada yang bisa dibantu hari ini?")
                    .font(.system(size: 24, weight: .semibold, design: .rounded))
                    .foregroundColor(AppColors.textPrimary.opacity(0.88))

                Text("Tanyakan apa saja, atau langsung minta saya mencatat transaksi di F-Tracker dan mengelola tugas di Task Management.")
                    .font(.system(size: 14))
                    .foregroundColor(AppColors.textSecondary)
                    .lineSpacing(4)
            }

            // Connected Satellites Quick Strip
            HStack(spacing: 10) {
                satelliteQuickPill(
                    title: "F-Tracker",
                    icon: "wallet.pass.fill",
                    color: AppColors.accentAmber,
                    connected: viewModel.linkedApps.contains(where: { $0.appId.lowercased().contains("tracker") })
                )
                satelliteQuickPill(
                    title: "Task Hub",
                    icon: "checklist",
                    color: AppColors.accentEmerald,
                    connected: viewModel.linkedApps.contains(where: { $0.appId.lowercased().contains("task") })
                )
            }

            // Gemini Suggested Prompt Cards
            VStack(alignment: .leading, spacing: 10) {
                Text("MULAI PERCAKAPAN")
                    .font(.system(size: 11, weight: .bold))
                    .tracking(0.8)
                    .foregroundColor(AppColors.textMuted)

                LazyVGrid(columns: [GridItem(.flexible(), spacing: 10), GridItem(.flexible(), spacing: 10)], spacing: 10) {
                    promptTile(
                        icon: "chart.bar.doc.horizontal.fill",
                        color: AppColors.geminiBlue,
                        title: "Ringkas Hari Ini",
                        prompt: "Berikan ringkasan keuangan F-Tracker dan prioritas tugas saya hari ini."
                    )
                    promptTile(
                        icon: "cup.and.saucer.fill",
                        color: AppColors.accentAmber,
                        title: "Catat Pengeluaran",
                        prompt: "Catat pengeluaran kopi Rp 35.000 di F-Tracker hari ini."
                    )
                    promptTile(
                        icon: "plus.circle.fill",
                        color: AppColors.accentEmerald,
                        title: "Buat Task Baru",
                        prompt: "Buat tugas prioritas tinggi di Task Management: Review target minggu ini."
                    )
                    promptTile(
                        icon: "sparkles",
                        color: AppColors.geminiPurple,
                        title: "Atur Jadwal Fokus",
                        prompt: "Lihat semua tugas yang pending lalu susunkan rencana fokus terbaik hari ini."
                    )
                }
            }
        }
    }

    private func satelliteQuickPill(title: String, icon: String, color: Color, connected: Bool) -> some View {
        Button(action: onOpenSatellites) {
            HStack(spacing: 6) {
                Image(systemName: icon)
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(connected ? color : AppColors.textMuted)
                Text(title)
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(AppColors.textPrimary)
                Image(systemName: connected ? "checkmark.circle.fill" : "plus.circle")
                    .font(.system(size: 11))
                    .foregroundColor(connected ? AppColors.accentEmerald : AppColors.textSecondary)
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 8)
            .glassCard(cornerRadius: 16)
        }
        .buttonStyle(.plain)
    }

    private func promptTile(icon: String, color: Color, title: String, prompt: String) -> some View {
        Button(action: {
            viewModel.inputText = prompt
            isInputFocused = true
        }) {
            VStack(alignment: .leading, spacing: 10) {
                ZStack {
                    Circle()
                        .fill(color.opacity(0.18))
                        .frame(width: 32, height: 32)
                    Image(systemName: icon)
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundColor(color)
                }

                Text(title)
                    .font(.system(size: 13, weight: .bold))
                    .foregroundColor(AppColors.textPrimary)

                Text(prompt)
                    .font(.system(size: 11))
                    .foregroundColor(AppColors.textSecondary)
                    .lineLimit(2)
                    .multilineTextAlignment(.leading)
            }
            .padding(14)
            .frame(maxWidth: .infinity, minHeight: 112, alignment: .topLeading)
            .glassCard(cornerRadius: 18)
        }
        .buttonStyle(.plain)
    }

    // MARK: - Thinking Indicator

    private var geminiThinkingRow: some View {
        ThinkingIndicatorView()
    }

    // MARK: - Composer Input Bar

    private var geminiComposerBar: some View {
        VStack(spacing: 0) {
            if let banner = viewModel.errorBanner, let _ = viewModel.lastFailedPrompt {
                HStack(spacing: 8) {
                    Image(systemName: "wifi.exclamationmark")
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundColor(Color(hex: "#F87171"))

                    Text(banner)
                        .font(.system(size: 12))
                        .foregroundColor(Color(hex: "#FCA5A5"))
                        .lineLimit(2)

                    Spacer()

                    Button(action: {
                        Task {
                            await viewModel.retryLastMessage()
                        }
                    }) {
                        HStack(spacing: 4) {
                            Image(systemName: "arrow.clockwise")
                                .font(.system(size: 11, weight: .bold))
                            Text("Coba Lagi")
                                .font(.system(size: 12, weight: .semibold))
                        }
                        .foregroundColor(.white)
                        .padding(.horizontal, 10)
                        .padding(.vertical, 5)
                        .background(Color(hex: "#EF4444"))
                        .clipShape(Capsule())
                    }
                    .buttonStyle(.plain)

                    Button(action: {
                        withAnimation {
                            viewModel.errorBanner = nil
                        }
                    }) {
                        Image(systemName: "xmark")
                            .font(.system(size: 10, weight: .bold))
                            .foregroundColor(AppColors.textMuted)
                            .padding(4)
                    }
                    .buttonStyle(.plain)
                }
                .padding(.horizontal, 12)
                .padding(.vertical, 8)
                .background(Color(hex: "#1F1315").opacity(0.95))
                .overlay(
                    RoundedRectangle(cornerRadius: 12)
                        .stroke(Color(hex: "#EF4444").opacity(0.35), lineWidth: 1)
                )
                .clipShape(RoundedRectangle(cornerRadius: 12))
                .padding(.horizontal, 14)
                .padding(.bottom, 6)
            }

            HStack(alignment: .bottom, spacing: 10) {
                TextField(L10n.Hub.chatPlaceholder, text: $viewModel.inputText, axis: .vertical)
                    .lineLimit(1...5)
                    .focused($isInputFocused)
                    .font(.system(size: 15))
                    .foregroundColor(.white)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)

                let canSend = !viewModel.inputText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty && !viewModel.isSending

                Button(action: {
                    isInputFocused = false
                    Task {
                        await viewModel.sendMessage()
                    }
                }) {
                    ZStack {
                        Circle()
                            .fill(
                                canSend
                                    ? AnyShapeStyle(AppColors.auroraButtonGradient)
                                    : AnyShapeStyle(Color.white.opacity(0.08))
                            )
                            .frame(width: 38, height: 38)

                        Image(systemName: "arrow.up")
                            .font(.system(size: 15, weight: .bold))
                            .foregroundColor(canSend ? .white : AppColors.textMuted)
                    }
                }
                .buttonStyle(.plain)
                .disabled(!canSend)
                .padding(.trailing, 6)
                .padding(.bottom, 6)
            }
            .background(
                RoundedRectangle(cornerRadius: 26, style: .continuous)
                    .fill(AppColors.surfaceElevated.opacity(0.95))
            )
            .overlay(
                RoundedRectangle(cornerRadius: 26, style: .continuous)
                    .stroke(
                        isInputFocused
                            ? AnyShapeStyle(AppColors.geminiGradient)
                            : AnyShapeStyle(Color.white.opacity(0.14)),
                        lineWidth: 1
                    )
            )
            .shadow(color: Color.black.opacity(0.4), radius: 14, x: 0, y: 6)
            .padding(.horizontal, 14)
            .padding(.top, 8)
            .padding(.bottom, 10)
        }
    }
}

private struct GeminiMessageRow: View {
    let message: ChatMessageItem
    let onExecuteDeepLink: (URL) -> Void
    var onSendQuickReply: ((String) -> Void)? = nil
    var onRetryMessage: ((UUID) -> Void)? = nil
    var onRestorePrompt: ((String) -> Void)? = nil
    @State private var copied = false

    var body: some View {
        if message.role == .user {
            VStack(alignment: .trailing, spacing: 6) {
                HStack(alignment: .bottom, spacing: 8) {
                    Spacer(minLength: 48)

                    if message.isFailed {
                        Button(action: {
                            onRetryMessage?(message.id)
                        }) {
                            HStack(spacing: 4) {
                                Image(systemName: "arrow.clockwise")
                                    .font(.system(size: 11, weight: .bold))
                                Text("Coba Lagi")
                                    .font(.system(size: 11, weight: .semibold))
                            }
                            .foregroundColor(Color(hex: "#F87171"))
                            .padding(.horizontal, 10)
                            .padding(.vertical, 6)
                            .background(Color(hex: "#EF4444").opacity(0.18))
                            .overlay(
                                Capsule().stroke(Color(hex: "#EF4444").opacity(0.4), lineWidth: 1)
                            )
                            .clipShape(Capsule())
                        }
                        .buttonStyle(.plain)
                    }

                    Text(message.content)
                        .font(.system(size: 15))
                        .foregroundColor(.white)
                        .padding(.horizontal, 16)
                        .padding(.vertical, 12)
                        .background(
                            message.isFailed
                                ? LinearGradient(
                                    colors: [Color(hex: "#7F1D1D").opacity(0.6), Color(hex: "#991B1B").opacity(0.4)],
                                    startPoint: .topLeading,
                                    endPoint: .bottomTrailing
                                )
                                : LinearGradient(
                                    colors: [AppColors.geminiBlue.opacity(0.32), AppColors.geminiPurple.opacity(0.28)],
                                    startPoint: .topLeading,
                                    endPoint: .bottomTrailing
                                )
                        )
                        .overlay(
                            RoundedRectangle(cornerRadius: 20, style: .continuous)
                                .stroke(
                                    message.isFailed
                                        ? Color(hex: "#EF4444").opacity(0.55)
                                        : Color.white.opacity(0.16),
                                    lineWidth: 1
                                )
                        )
                        .clipShape(RoundedRectangle(cornerRadius: 20, style: .continuous))
                }

                if message.isFailed {
                    HStack(spacing: 10) {
                        Spacer(minLength: 48)

                        Label(message.errorMessage ?? "Gagal terkirim", systemImage: "exclamationmark.triangle.fill")
                            .font(.system(size: 11, weight: .medium))
                            .foregroundColor(Color(hex: "#F87171"))

                        Button(action: {
                            onRestorePrompt?(message.content)
                        }) {
                            Text("Salin ke Input")
                                .font(.system(size: 11, weight: .semibold))
                                .foregroundColor(AppColors.geminiCyan)
                                .underline()
                        }
                        .buttonStyle(.plain)
                    }
                }
            }
        } else {
            HStack(alignment: .top, spacing: 12) {
                GeminiSparkleView(size: 18, animate: false)
                    .frame(width: 28, height: 28)
                    .padding(.top, 2)

                VStack(alignment: .leading, spacing: 10) {
                    if let toolCalls = message.toolCalls, !toolCalls.isEmpty {
                        ScrollView(.horizontal, showsIndicators: false) {
                            HStack(spacing: 6) {
                                ForEach(toolCalls.indices, id: \.self) { idx in
                                    HStack(spacing: 4) {
                                        Image(systemName: "bolt.fill")
                                            .font(.system(size: 10))
                                            .foregroundColor(AppColors.geminiCyan)
                                        Text(toolCalls[idx].name.replacingOccurrences(of: "_", with: " "))
                                            .font(.system(size: 11, weight: .semibold))
                                            .foregroundColor(AppColors.geminiCyan)
                                    }
                                    .padding(.horizontal, 9)
                                    .padding(.vertical, 4)
                                    .background(AppColors.geminiCyan.opacity(0.12))
                                    .clipShape(Capsule())
                                }
                            }
                        }
                    }

                    GeminiMarkdownView(content: message.content)
                        .textSelection(.enabled)

                    if let quickReplies = message.quickReplies, !quickReplies.isEmpty {
                        ScrollView(.horizontal, showsIndicators: false) {
                            HStack(spacing: 8) {
                                ForEach(quickReplies, id: \.self) { reply in
                                    let isPositive = reply.lowercased().contains("ya") || reply.contains("✅")
                                    Button(action: { onSendQuickReply?(reply) }) {
                                        Text(reply)
                                            .font(.system(size: 13, weight: .semibold))
                                            .foregroundColor(isPositive ? Color(hex: "#34D399") : Color(hex: "#F87171"))
                                            .padding(.horizontal, 14)
                                            .padding(.vertical, 8)
                                            .background((isPositive ? Color(hex: "#10B981") : Color(hex: "#EF4444")).opacity(0.18))
                                            .overlay(
                                                RoundedRectangle(cornerRadius: 12)
                                                    .stroke((isPositive ? Color(hex: "#10B981") : Color(hex: "#EF4444")).opacity(0.5), lineWidth: 1)
                                            )
                                            .clipShape(RoundedRectangle(cornerRadius: 12))
                                    }
                                    .buttonStyle(.plain)
                                }
                            }
                            .padding(.vertical, 2)
                        }
                    }

                    if let actions = message.actionLinks, !actions.isEmpty {
                        VStack(spacing: 8) {
                            ForEach(actions) { action in
                                if let url = URL(string: action.deepLink) {
                                    Button(action: { onExecuteDeepLink(url) }) {
                                        HStack(spacing: 8) {
                                            Image(systemName: "arrow.up.forward.app.fill")
                                                .foregroundColor(AppColors.geminiBlue)
                                            VStack(alignment: .leading, spacing: 1) {
                                                Text(action.title)
                                                    .font(.system(size: 13, weight: .semibold))
                                                    .foregroundColor(.white)
                                                Text(action.description)
                                                    .font(.system(size: 11))
                                                    .foregroundColor(AppColors.textSecondary)
                                            }
                                            Spacer()
                                            Image(systemName: "chevron.right")
                                                .font(.system(size: 11, weight: .bold))
                                                .foregroundColor(AppColors.textSecondary)
                                        }
                                        .padding(12)
                                        .glassCard(cornerRadius: 14)
                                    }
                                    .buttonStyle(.plain)
                                }
                            }
                        }
                    }

                    HStack(spacing: 14) {
                        Button(action: {
                            #if canImport(UIKit)
                            UIPasteboard.general.string = message.content
                            #endif
                            copied = true
                            DispatchQueue.main.asyncAfter(deadline: .now() + 1.5) {
                                copied = false
                            }
                        }) {
                            Label(copied ? "Disalin" : "Salin", systemImage: copied ? "checkmark" : "doc.on.doc")
                                .font(.system(size: 11, weight: .medium))
                                .foregroundColor(AppColors.textMuted)
                        }
                        .buttonStyle(.plain)
                    }
                    .padding(.top, 2)
                }

                Spacer(minLength: 12)
            }
        }
    }
}

public struct ThinkingIndicatorView: View {
    @State private var dot1Offset: CGFloat = 0
    @State private var dot2Offset: CGFloat = 0
    @State private var dot3Offset: CGFloat = 0
    @State private var shimmerPhase: CGFloat = 0

    private var primaryX: CGFloat {
        -56 + (shimmerPhase * 196)
    }

    private var followX: CGFloat {
        primaryX - 196
    }

    private var exitX: CGFloat {
        primaryX + 196
    }

    @ViewBuilder
    private func shimmerBeam(offset: CGFloat) -> some View {
        Capsule()
            .fill(
                LinearGradient(
                    colors: [
                        Color.clear,
                        AppColors.aiCyan.opacity(0.85),
                        AppColors.aiBlue.opacity(0.85),
                        Color.clear
                    ],
                    startPoint: .leading,
                    endPoint: .trailing
                )
            )
            .frame(width: 56, height: 3.5)
            .offset(x: offset)
            .shadow(color: AppColors.aiCyan.opacity(0.8), radius: 6)
    }

    public init() {}

    public var body: some View {
        HStack(alignment: .top, spacing: 12) {
            ZStack {
                Circle()
                    .fill(AppColors.aiGradient)
                    .frame(width: 32, height: 32)
                    .blur(radius: 8)
                    .opacity(0.3)

                RoundedRectangle(cornerRadius: 10, style: .continuous)
                    .fill(AppColors.surfaceElevated)
                    .frame(width: 30, height: 30)
                    .overlay(
                        RoundedRectangle(cornerRadius: 10, style: .continuous)
                            .stroke(AppColors.aiGradient, lineWidth: 1)
                    )

                HubSparkleView(size: 16, animate: true)
            }

            VStack(alignment: .leading, spacing: 8) {
                HStack(spacing: 8) {
                    Text("Sedang berpikir & menganalisis data...")
                        .font(.system(size: 13, weight: .medium))
                        .foregroundStyle(
                            LinearGradient(
                                colors: [.white, AppColors.textSecondary],
                                startPoint: .leading,
                                endPoint: .trailing
                            )
                        )

                    // 3 Animated Bouncing Dots
                    HStack(spacing: 4) {
                        Circle()
                            .fill(AppColors.aiCyan)
                            .frame(width: 5, height: 5)
                            .offset(y: dot1Offset)

                        Circle()
                            .fill(AppColors.aiPurple)
                            .frame(width: 5, height: 5)
                            .offset(y: dot2Offset)

                        Circle()
                            .fill(AppColors.aiRose)
                            .frame(width: 5, height: 5)
                            .offset(y: dot3Offset)
                    }
                }

                // Fluid Carousel Shimmer Capsule Loader (full edge exit, zero pop)
                ZStack(alignment: .leading) {
                    Capsule()
                        .fill(AppColors.surfaceCard.opacity(0.8))
                        .frame(width: 140, height: 3.5)

                    shimmerBeam(offset: primaryX)
                    shimmerBeam(offset: followX)
                    shimmerBeam(offset: exitX)
                }
                .frame(width: 140, height: 3.5)
                .clipShape(Capsule())
            }
            .padding(.top, 4)

            Spacer()
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .glassCard(cornerRadius: 16)
        .onAppear {
            withAnimation(.easeInOut(duration: 0.5).repeatForever(autoreverses: true)) {
                dot1Offset = -5
            }
            withAnimation(.easeInOut(duration: 0.5).repeatForever(autoreverses: true).delay(0.15)) {
                dot2Offset = -5
            }
            withAnimation(.easeInOut(duration: 0.5).repeatForever(autoreverses: true).delay(0.3)) {
                dot3Offset = -5
            }
            withAnimation(.linear(duration: 1.25).repeatForever(autoreverses: false)) {
                shimmerPhase = 1.0
            }
        }
    }
}

// MARK: - Markdown Block Data Model

public enum MarkdownContentBlock: Identifiable, Sendable {
    case code(id: UUID = UUID(), language: String, code: String)
    case heading(id: UUID = UUID(), level: Int, text: String)
    case bullet(id: UUID = UUID(), text: String)
    case numbered(id: UUID = UUID(), number: String, text: String)
    case paragraph(id: UUID = UUID(), text: String)

    public var id: UUID {
        switch self {
        case .code(let id, _, _): return id
        case .heading(let id, _, _): return id
        case .bullet(let id, _): return id
        case .numbered(let id, _, _): return id
        case .paragraph(let id, _): return id
        }
    }
}

// MARK: - Gemini-Grade Markdown View

public struct GeminiMarkdownView: View {
    public let content: String
    private let blocks: [MarkdownContentBlock]

    public init(content: String) {
        self.content = content
        self.blocks = GeminiMarkdownParser.parse(content)
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            ForEach(blocks) { block in
                switch block {
                case .code(_, let language, let code):
                    GeminiCodeBlockCard(language: language, code: code)

                case .heading(_, let level, let text):
                    let (fontSize, weight): (CGFloat, Font.Weight) = {
                        switch level {
                        case 1: return (18, .bold)
                        case 2: return (16, .semibold)
                        default: return (14, .semibold)
                        }
                    }()
                    Text(GeminiMarkdownParser.inlineMarkdown(text))
                        .font(.system(size: fontSize, weight: weight))
                        .foregroundColor(.white)
                        .padding(.top, 4)
                        .padding(.bottom, 2)

                case .bullet(_, let text):
                    HStack(alignment: .top, spacing: 8) {
                        Text("•")
                            .font(.system(size: 15, weight: .bold))
                            .foregroundColor(AppColors.geminiCyan)
                        Text(GeminiMarkdownParser.inlineMarkdown(text))
                            .font(.system(size: 14))
                            .foregroundColor(AppColors.textPrimary.opacity(0.95))
                            .lineSpacing(3)
                    }
                    .padding(.leading, 4)

                case .numbered(_, let number, let text):
                    HStack(alignment: .top, spacing: 8) {
                        Text("\(number).")
                            .font(.system(size: 13, weight: .bold))
                            .foregroundColor(AppColors.geminiBlue)
                        Text(GeminiMarkdownParser.inlineMarkdown(text))
                            .font(.system(size: 14))
                            .foregroundColor(AppColors.textPrimary.opacity(0.95))
                            .lineSpacing(3)
                    }
                    .padding(.leading, 4)

                case .paragraph(_, let text):
                    Text(GeminiMarkdownParser.inlineMarkdown(text))
                        .font(.system(size: 14))
                        .foregroundColor(AppColors.textPrimary.opacity(0.95))
                        .lineSpacing(3)
                }
            }
        }
    }
}

// MARK: - Gemini Code Block Card

public struct GeminiCodeBlockCard: View {
    public let language: String
    public let code: String
    @State private var copied: Bool = false

    private var displayLanguage: String {
        language.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? "CODE" : language.uppercased()
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            // Header Bar
            HStack {
                Text(displayLanguage)
                    .font(.system(size: 11, weight: .semibold, design: .monospaced))
                    .foregroundColor(Color(hex: "94A3B8"))

                Spacer()

                Button(action: copyToClipboard) {
                    HStack(spacing: 4) {
                        Image(systemName: copied ? "checkmark" : "doc.on.doc")
                            .font(.system(size: 11, weight: .medium))
                            .foregroundColor(copied ? Color(hex: "4ADE80") : Color(hex: "94A3B8"))

                        Text(copied ? "Tersalin!" : "Salin kode")
                            .font(.system(size: 11, weight: .medium))
                            .foregroundColor(copied ? Color(hex: "4ADE80") : Color(hex: "94A3B8"))
                    }
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(Color.white.opacity(0.04))
                    .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
                }
                .buttonStyle(.plain)
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 8)
            .background(Color(hex: "131722"))

            // Code Body with horizontal scrolling
            ScrollView(.horizontal, showsIndicators: false) {
                Text(GeminiSyntaxHighlighter.highlight(code))
                    .font(.system(size: 13, design: .monospaced))
                    .foregroundColor(Color(hex: "F1F5F9"))
                    .lineSpacing(4)
                    .padding(.horizontal, 14)
                    .padding(.vertical, 12)
            }
            .background(Color(hex: "0B0E17"))
        }
        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .stroke(Color(hex: "232838"), lineWidth: 1)
        )
    }

    private func copyToClipboard() {
        #if canImport(UIKit)
        UIPasteboard.general.string = code
        #endif
        withAnimation(.easeInOut(duration: 0.2)) {
            copied = true
        }
        Task {
            try? await Task.sleep(nanoseconds: 2_000_000_000)
            withAnimation(.easeInOut(duration: 0.2)) {
                copied = false
            }
        }
    }
}

// MARK: - Markdown Parser

public enum GeminiMarkdownParser {
    public static func parse(_ text: String) -> [MarkdownContentBlock] {
        var blocks: [MarkdownContentBlock] = []
        let lines = text.components(separatedBy: "\n")
        var i = 0

        while i < lines.count {
            let line = lines[i]
            let trimmed = line.trimmingCharacters(in: .whitespaces)

            // Code Block: ```language
            if trimmed.hasPrefix("```") {
                let lang = String(trimmed.dropFirst(3)).trimmingCharacters(in: .whitespaces)
                var codeLines: [String] = []
                i += 1
                while i < lines.count && !lines[i].trimmingCharacters(in: .whitespaces).hasPrefix("```") {
                    codeLines.append(lines[i])
                    i += 1
                }
                if i < lines.count && lines[i].trimmingCharacters(in: .whitespaces).hasPrefix("```") {
                    i += 1
                }
                blocks.append(.code(language: lang, code: codeLines.joined(separator: "\n")))
                continue
            }

            // Headings: # H1, ## H2, ### H3
            if trimmed.hasPrefix("#") {
                var level = 0
                for char in trimmed {
                    if char == "#" { level += 1 } else { break }
                }
                let headingText = String(trimmed.dropFirst(level)).trimmingCharacters(in: .whitespaces)
                if !headingText.isEmpty {
                    blocks.append(.heading(level: min(max(level, 1), 4), text: headingText))
                    i += 1
                    continue
                }
            }

            // Bullet items: - item or * item
            if trimmed.hasPrefix("- ") || trimmed.hasPrefix("* ") {
                blocks.append(.bullet(text: String(trimmed.dropFirst(2)).trimmingCharacters(in: .whitespaces)))
                i += 1
                continue
            }

            // Numbered items: 1. item, 2. item
            if let match = matchNumberedList(trimmed) {
                blocks.append(.numbered(number: match.number, text: match.text))
                i += 1
                continue
            }

            // Regular paragraph
            if !trimmed.isEmpty {
                blocks.append(.paragraph(text: line))
            }
            i += 1
        }

        return blocks
    }

    private static func matchNumberedList(_ text: String) -> (number: String, text: String)? {
        guard let dotIndex = text.firstIndex(of: ".") else { return nil }
        let prefix = text[..<dotIndex]
        guard let num = Int(prefix) else { return nil }
        let remainder = text[text.index(after: dotIndex)...].trimmingCharacters(in: .whitespaces)
        return (number: "\(num)", text: remainder)
    }

    public static func inlineMarkdown(_ text: String) -> AttributedString {
        if let attr = try? AttributedString(markdown: text, options: AttributedString.MarkdownParsingOptions(interpretedSyntax: .inlineOnlyPreservingWhitespace)) {
            return attr
        }
        return AttributedString(text)
    }
}

// MARK: - Lexical Syntax Highlighter

public enum GeminiSyntaxHighlighter {
    public static func highlight(_ code: String) -> AttributedString {
        var attributed = AttributedString(code)

        let keywordColor = Color(hex: "C084FC")
        let stringColor = Color(hex: "86EFAC")
        let numberColor = Color(hex: "FDBA74")
        let commentColor = Color(hex: "94A3B8")
        let typeColor = Color(hex: "38BDF8")

        let keywords: Set<String> = [
            "func", "let", "var", "struct", "class", "enum", "protocol", "extension", "guard",
            "if", "else", "switch", "case", "default", "return", "import", "package", "public",
            "private", "async", "await", "throw", "throws", "try", "catch", "defer", "while",
            "for", "in", "type", "const", "def", "lambda", "with", "yield", "SELECT", "FROM",
            "WHERE", "JOIN", "ORDER", "BY", "GROUP", "LIMIT", "INSERT", "INTO", "UPDATE", "SET",
            "DELETE", "AND", "OR", "NOT", "fun", "val", "override"
        ]

        let types: Set<String> = [
            "String", "Int", "Double", "Float", "Bool", "Void", "UUID", "Date", "URL",
            "string", "int", "int64", "bool", "float64", "error", "byte",
            "true", "false", "nil", "null", "None"
        ]

        let nsCode = code as NSString
        let fullRange = NSRange(location: 0, length: nsCode.length)

        if let wordRegex = try? NSRegularExpression(pattern: "\\b[a-zA-Z_][a-zA-Z0-9_]*\\b") {
            let matches = wordRegex.matches(in: code, range: fullRange)
            for match in matches {
                let word = nsCode.substring(with: match.range)
                if keywords.contains(word) {
                    if let range = Range(match.range, in: code),
                       let attrRange = Range(range, in: attributed) {
                        attributed[attrRange].foregroundColor = keywordColor
                        attributed[attrRange].inlinePresentationIntent = .stronglyEmphasized
                    }
                } else if types.contains(word) {
                    if let range = Range(match.range, in: code),
                       let attrRange = Range(range, in: attributed) {
                        attributed[attrRange].foregroundColor = typeColor
                    }
                }
            }
        }

        if let strRegex = try? NSRegularExpression(pattern: "\"(\\\\.|[^\"])*\"|'(\\\\.|[^'])*'") {
            let matches = strRegex.matches(in: code, range: fullRange)
            for match in matches {
                if let range = Range(match.range, in: code),
                   let attrRange = Range(range, in: attributed) {
                    attributed[attrRange].foregroundColor = stringColor
                }
            }
        }

        if let numRegex = try? NSRegularExpression(pattern: "\\b\\d+(\\.\\d+)?\\b") {
            let matches = numRegex.matches(in: code, range: fullRange)
            for match in matches {
                if let range = Range(match.range, in: code),
                   let attrRange = Range(range, in: attributed) {
                    attributed[attrRange].foregroundColor = numberColor
                }
            }
        }

        if let commentRegex = try? NSRegularExpression(pattern: "//.*|#.*|/\\*[\\s\\S]*?\\*/") {
            let matches = commentRegex.matches(in: code, range: fullRange)
            for match in matches {
                if let range = Range(match.range, in: code),
                   let attrRange = Range(range, in: attributed) {
                    attributed[attrRange].foregroundColor = commentColor
                    attributed[attrRange].font = .system(size: 13, design: .monospaced).italic()
                }
            }
        }

        return attributed
    }
}

// MARK: - Chat History Sheet View

public struct ChatHistorySheetView: View {
    public let sessions: [ChatSessionItem]
    public let currentSessionId: String?
    public let isLoading: Bool
    public let onSelectSession: (String) -> Void
    public let onNewChat: () -> Void
    public let onDeleteSession: (String) -> Void
    @Environment(\.dismiss) private var dismiss

    public init(
        sessions: [ChatSessionItem],
        currentSessionId: String?,
        isLoading: Bool,
        onSelectSession: @escaping (String) -> Void,
        onNewChat: @escaping () -> Void,
        onDeleteSession: @escaping (String) -> Void
    ) {
        self.sessions = sessions
        self.currentSessionId = currentSessionId
        self.isLoading = isLoading
        self.onSelectSession = onSelectSession
        self.onNewChat = onNewChat
        self.onDeleteSession = onDeleteSession
    }

    public var body: some View {
        NavigationStack {
            ZStack {
                AppColors.backgroundBase.ignoresSafeArea()

                VStack(spacing: 16) {
                    Button(action: {
                        onNewChat()
                        dismiss()
                    }) {
                        HStack(spacing: 12) {
                            ZStack {
                                Circle()
                                    .fill(AppColors.geminiBlue.opacity(0.2))
                                    .frame(width: 34, height: 34)
                                Image(systemName: "plus")
                                    .font(.system(size: 14, weight: .bold))
                                    .foregroundColor(AppColors.geminiCyan)
                            }

                            Text("Mulai Obrolan Baru")
                                .font(.system(size: 15, weight: .semibold))
                                .foregroundColor(.white)

                            Spacer()

                            Image(systemName: "square.and.pencil")
                                .font(.system(size: 14))
                                .foregroundColor(AppColors.textMuted)
                        }
                        .padding(.horizontal, 16)
                        .padding(.vertical, 12)
                        .background(
                            RoundedRectangle(cornerRadius: 14, style: .continuous)
                                .fill(AppColors.surfaceElevated)
                        )
                        .overlay(
                            RoundedRectangle(cornerRadius: 14, style: .continuous)
                                .stroke(AppColors.geminiBlue.opacity(0.3), lineWidth: 1)
                        )
                    }
                    .buttonStyle(.plain)
                    .padding(.horizontal, 16)
                    .padding(.top, 8)

                    if sessions.isEmpty {
                        Spacer()
                        VStack(spacing: 10) {
                            Image(systemName: "bubble.left.and.bubble.right")
                                .font(.system(size: 36))
                                .foregroundColor(AppColors.textMuted.opacity(0.6))
                            Text(isLoading ? "Memuat riwayat obrolan..." : "Belum ada riwayat percakapan")
                                .font(.system(size: 14))
                                .foregroundColor(AppColors.textSecondary)
                        }
                        Spacer()
                    } else {
                        List {
                            ForEach(sessions) { session in
                                let isSelected = session.id == currentSessionId
                                Button(action: {
                                    onSelectSession(session.id)
                                    dismiss()
                                }) {
                                    HStack(alignment: .center, spacing: 12) {
                                        Image(systemName: isSelected ? "bubble.left.fill" : "bubble.left")
                                            .font(.system(size: 15))
                                            .foregroundColor(isSelected ? AppColors.geminiCyan : AppColors.textMuted)

                                        VStack(alignment: .leading, spacing: 3) {
                                            Text(session.title)
                                                .font(.system(size: 14, weight: isSelected ? .semibold : .medium))
                                                .foregroundColor(isSelected ? .white : AppColors.textPrimary)
                                                .lineLimit(1)

                                            if !session.lastPreview.isEmpty {
                                                Text(session.lastPreview)
                                                    .font(.system(size: 12))
                                                    .foregroundColor(AppColors.textMuted)
                                                    .lineLimit(1)
                                            }
                                        }

                                        Spacer()

                                        if isSelected {
                                            Circle()
                                                .fill(AppColors.geminiCyan)
                                                .frame(width: 7, height: 7)
                                        }
                                    }
                                    .padding(.vertical, 4)
                                }
                                .listRowBackground(
                                    isSelected
                                        ? AppColors.surfaceElevated.opacity(0.9)
                                        : Color.clear
                                )
                                .swipeActions(edge: .trailing, allowsFullSwipe: true) {
                                    Button(role: .destructive) {
                                        onDeleteSession(session.id)
                                    } label: {
                                        Label("Hapus", systemImage: "trash")
                                    }
                                }
                            }
                        }
                        .listStyle(.plain)
                        .scrollContentBackground(.hidden)
                    }
                }
            }
            .navigationTitle("Riwayat Obrolan")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Selesai") {
                        dismiss()
                    }
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundColor(AppColors.geminiCyan)
                }
            }
        }
    }
}
