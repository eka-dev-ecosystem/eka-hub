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

    private let apiClient = APIClient.shared

    public init() {
        Task {
            await loadLinkedApps()
        }
    }

    public func loadLinkedApps() async {
        do {
            linkedApps = try await apiClient.fetchLinkedApps()
        } catch {
            AppLogger.warn("Failed to load linked apps for chat header", category: .hub, metadata: ["error": error.localizedDescription])
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
        messages.removeAll()
        errorBanner = nil
    }

    private func submitPrompt(_ prompt: String) async {
        errorBanner = nil
        let userMsg = ChatMessageItem(role: .user, content: prompt)
        messages.append(userMsg)
        isSending = true

        do {
            let response = try await apiClient.sendChat(message: prompt)
            isSending = false

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
