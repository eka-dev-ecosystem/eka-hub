import Foundation
import SwiftUI

@MainActor
@Observable
public final class AiChatViewModel {
    public var messages: [ChatMessageItem] = []
    public var inputText: String = ""
    public var isSending: Bool = false
    public var linkedApps: [LinkedAppItem] = []
    public var errorBanner: String?

    public var currentSessionId: String?
    public var sessions: [ChatSessionItem] = []
    public var isHistoryLoading: Bool = false
    public var showHistorySheet: Bool = false

    private let apiClient = APIClient.shared

    public init() {
        Task {
            await loadLinkedApps()
            await loadSessions()
        }
    }

    public func loadLinkedApps() async {
        do {
            linkedApps = try await apiClient.fetchLinkedApps()
        } catch {
            AppLogger.warn("Failed to load linked apps for chat header", category: .hub, metadata: ["error": error.localizedDescription])
        }
    }

    public func loadSessions() async {
        isHistoryLoading = true
        do {
            sessions = try await apiClient.fetchChatSessions()
        } catch {
            AppLogger.warn("Failed to load chat sessions", category: .hub, metadata: ["error": error.localizedDescription])
        }
        isHistoryLoading = false
    }

    public func selectSession(id: String) async {
        guard !isSending else { return }
        currentSessionId = id
        isSending = true
        do {
            let detail = try await apiClient.fetchSessionDetail(id: id)
            messages = detail.messages.map { stored in
                ChatMessageItem(
                    role: stored.role == "user" ? .user : .assistant,
                    content: stored.content
                )
            }
        } catch {
            errorBanner = "Gagal memuat riwayat obrolan: \(error.localizedDescription)"
        }
        isSending = false
    }

    public func deleteSession(id: String) async {
        do {
            try await apiClient.deleteChatSession(id: id)
            sessions.removeAll { $0.id == id }
            if currentSessionId == id {
                startNewChat()
            }
        } catch {
            AppLogger.warn("Failed to delete chat session", category: .hub, metadata: ["error": error.localizedDescription])
        }
    }

    public func sendMessage() async {
        let prompt = inputText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !prompt.isEmpty, !isSending else { return }
        inputText = ""
        await submitPrompt(prompt)
    }

    public func sendQuickPrompt(_ prompt: String) async {
        guard !prompt.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty, !isSending else { return }
        await submitPrompt(prompt)
    }

    public func startNewChat() {
        guard !isSending else { return }
        currentSessionId = nil
        messages.removeAll()
        errorBanner = nil
    }

    private func submitPrompt(_ prompt: String) async {
        errorBanner = nil
        let userMsg = ChatMessageItem(role: .user, content: prompt)
        messages.append(userMsg)
        isSending = true

        do {
            let response = try await apiClient.sendChat(message: prompt, sessionId: currentSessionId)
            isSending = false

            if let sid = response.sessionId, !sid.isEmpty {
                currentSessionId = sid
            }

            // Progressive streaming typewriter for live, real-time response feel
            let streamingMsg = ChatMessageItem(
                role: .assistant,
                content: "",
                toolCalls: response.toolCalls,
                actionLinks: response.actionLinks
            )
            messages.append(streamingMsg)
            let streamingId = streamingMsg.id

            let words = response.reply.split(separator: " ")
            var accumulated = ""
            for (idx, word) in words.enumerated() {
                accumulated += (idx == 0 ? "" : " ") + word
                if let index = messages.firstIndex(where: { $0.id == streamingId }) {
                    messages[index].content = accumulated
                }
                try? await Task.sleep(nanoseconds: 20_000_000) // 20ms per word
            }
            if let index = messages.firstIndex(where: { $0.id == streamingId }) {
                messages[index].content = response.reply
            }
            await loadLinkedApps()
            await loadSessions()
        } catch {
            isSending = false
            errorBanner = error.localizedDescription
            let fallbackMsg = ChatMessageItem(
                role: .assistant,
                content: "Maaf, permintaan belum berhasil diproses: \(error.localizedDescription)"
            )
            messages.append(fallbackMsg)
        }
    }
}
