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

    public var lastFailedPrompt: String?
    public var lastFailedMessageId: UUID?

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

    public func retryMessage(id: UUID) async {
        guard !isSending else { return }
        guard let index = messages.firstIndex(where: { $0.id == id }) else { return }
        let prompt = messages[index].content
        await submitPrompt(prompt, targetMessageId: id)
    }

    public func retryLastMessage() async {
        guard !isSending else { return }
        if let targetId = lastFailedMessageId, messages.contains(where: { $0.id == targetId }) {
            await retryMessage(id: targetId)
        } else if let prompt = lastFailedPrompt, !prompt.isEmpty {
            await submitPrompt(prompt)
        }
    }

    public func restorePrompt(_ prompt: String) {
        inputText = prompt
        errorBanner = nil
    }

    public func startNewChat() {
        guard !isSending else { return }
        currentSessionId = nil
        messages.removeAll()
        errorBanner = nil
        lastFailedPrompt = nil
        lastFailedMessageId = nil
    }

    private func submitPrompt(_ prompt: String, targetMessageId: UUID? = nil) async {
        errorBanner = nil
        let userMsgId: UUID

        if let existingId = targetMessageId, let idx = messages.firstIndex(where: { $0.id == existingId }) {
            userMsgId = existingId
            messages[idx].isFailed = false
            messages[idx].errorMessage = nil
        } else {
            let userMsg = ChatMessageItem(role: .user, content: prompt)
            messages.append(userMsg)
            userMsgId = userMsg.id
        }
        isSending = true

        do {
            let response = try await apiClient.sendChat(message: prompt, sessionId: currentSessionId)
            isSending = false
            lastFailedPrompt = nil
            lastFailedMessageId = nil

            if let sid = response.sessionId, !sid.isEmpty {
                currentSessionId = sid
            }

            // Progressive streaming typewriter for live, real-time response feel
            let streamingMsg = ChatMessageItem(
                role: .assistant,
                content: "",
                toolCalls: response.toolCalls,
                actionLinks: response.actionLinks,
                quickReplies: response.quickReplies
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
            lastFailedPrompt = prompt
            lastFailedMessageId = userMsgId

            let errDesc = error.localizedDescription
            let isNetworkIssue = errDesc.lowercased().contains("cancel") ||
                errDesc.lowercased().contains("network") ||
                errDesc.lowercased().contains("internet") ||
                errDesc.lowercased().contains("connection") ||
                errDesc.lowercased().contains("offline") ||
                errDesc.lowercased().contains("timed out") ||
                errDesc.lowercased().contains("terputus")

            let userFriendlyMsg = isNetworkIssue
                ? "Koneksi terputus saat aplikasi ditutup/berpindah. Ketuk Coba Lagi."
                : "Gagal terkirim: \(errDesc)"

            if let idx = messages.firstIndex(where: { $0.id == userMsgId }) {
                messages[idx].isFailed = true
                messages[idx].errorMessage = userFriendlyMsg
            }
            errorBanner = userFriendlyMsg
        }
    }
}
