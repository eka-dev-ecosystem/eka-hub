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
                                        onExecuteDeepLink: onExecuteDeepLink
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
    }

    // MARK: - Top Navigation Header

    private var geminiTopBar: some View {
        HStack(spacing: 12) {
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

            if !viewModel.messages.isEmpty {
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
            isInputFocused = false
            Task {
                await viewModel.sendQuickPrompt(prompt)
            }
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
        HStack(alignment: .top, spacing: 12) {
            GeminiSparkleView(size: 20, animate: true)
                .frame(width: 30, height: 30)

            VStack(alignment: .leading, spacing: 6) {
                Text("Sedang berpikir...")
                    .font(.system(size: 13, weight: .medium))
                    .foregroundStyle(AppColors.geminiGradient)

                RoundedRectangle(cornerRadius: 8)
                    .fill(AppColors.geminiGradient.opacity(0.28))
                    .frame(width: 140, height: 6)
            }
            .padding(.top, 4)

            Spacer()
        }
    }

    // MARK: - Composer Input Bar

    private var geminiComposerBar: some View {
        VStack(spacing: 0) {
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
    @State private var copied = false

    var body: some View {
        if message.role == .user {
            HStack {
                Spacer(minLength: 48)
                Text(message.content)
                    .font(.system(size: 15))
                    .foregroundColor(.white)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)
                    .background(
                        LinearGradient(
                            colors: [AppColors.geminiBlue.opacity(0.32), AppColors.geminiPurple.opacity(0.28)],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                    .overlay(
                        RoundedRectangle(cornerRadius: 20, style: .continuous)
                            .stroke(Color.white.opacity(0.16), lineWidth: 1)
                    )
                    .clipShape(RoundedRectangle(cornerRadius: 20, style: .continuous))
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

                    Text(LocalizedStringKey(message.content))
                        .font(.system(size: 15))
                        .foregroundColor(AppColors.textPrimary.opacity(0.95))
                        .lineSpacing(4)
                        .textSelection(.enabled)

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
