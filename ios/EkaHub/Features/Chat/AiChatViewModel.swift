import SwiftUI

@MainActor
@Observable
public final class AiChatViewModel {
    public var messages: [ChatMessageItem] = [
        ChatMessageItem(
            role: "assistant",
            content: "Halo! Saya Personal AI Hub Anda. Ada yang ingin Anda tanyakan seputar keuangan F-Tracker atau agenda tugas Task Management?",
            actions: [
                SuggestedAction(title: "Lihat F-Tracker", deepLink: "ftracker://expenses?expected_uid=usr_ftracker_123"),
                SuggestedAction(title: "Buka Task Hari Ini", deepLink: "taskhub://tasks?expected_uid=usr_task_456")
            ]
        )
    ]

    public init() {}

    public func sendMessage(_ text: String) {
        let trimmed = text.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty { return }

        messages.append(ChatMessageItem(role: "user", content: trimmed))

        let lower = trimmed.lowercased()
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) { [weak self] in
            guard let self = self else { return }
            if lower.contains("uang") || lower.contains("pengeluaran") || lower.contains("budget") {
                self.messages.append(
                    ChatMessageItem(
                        role: "assistant",
                        content: "Pengeluaran F-Tracker tercatat Rp 4.250.000. Sisa budget bulan ini Rp 5.750.000.",
                        actions: [
                            SuggestedAction(title: "Detail F-Tracker", deepLink: "ftracker://expenses?category=food&expected_uid=usr_ftracker_123")
                        ]
                    )
                )
            } else if lower.contains("tugas") || lower.contains("task") {
                self.messages.append(
                    ChatMessageItem(
                        role: "assistant",
                        content: "Anda memiliki 5 tugas pending di Task Management. 2 deadline jatuh hari ini.",
                        actions: [
                            SuggestedAction(title: "Buka Task Hub", deepLink: "taskhub://tasks?filter=today&expected_uid=usr_task_456")
                        ]
                    )
                )
            } else {
                self.messages.append(
                    ChatMessageItem(
                        role: "assistant",
                        content: "Finansial dan tugas Anda tersinkronisasi. Mau saya buka aplikasi yang mana?",
                        actions: [
                            SuggestedAction(title: "F-Tracker", deepLink: "ftracker://dashboard?expected_uid=usr_ftracker_123"),
                            SuggestedAction(title: "Task Management", deepLink: "taskhub://dashboard?expected_uid=usr_task_456")
                        ]
                    )
                )
            }
        }
    }
}
