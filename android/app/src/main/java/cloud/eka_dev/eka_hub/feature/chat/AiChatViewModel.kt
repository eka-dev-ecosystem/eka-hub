package cloud.eka_dev.eka_hub.feature.chat

import androidx.lifecycle.ViewModel
import cloud.eka_dev.eka_hub.core.model.ChatMessage
import cloud.eka_dev.eka_hub.core.model.SuggestedAction
import cloud.eka_dev.eka_hub.core.network.ApiClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class AiChatViewModel(private val apiClient: ApiClient) : ViewModel() {

    private val _messages = MutableStateFlow<List<ChatMessage>>(
        listOf(
            ChatMessage(
                role = "assistant",
                content = "Halo! Saya Personal AI Hub Anda. Ada yang ingin Anda tanyakan seputar keuangan F-Tracker atau agenda tugas Task Management?",
                actions = listOf(
                    SuggestedAction("Lihat F-Tracker", deepLink = "ftracker://expenses?expected_uid=usr_ftracker_123"),
                    SuggestedAction("Buka Task Hari Ini", deepLink = "taskhub://tasks?expected_uid=usr_task_456")
                )
            )
        )
    )
    val messages: StateFlow<List<ChatMessage>> = _messages.asStateFlow()

    fun sendMessage(query: String) {
        val trimmed = query.trim()
        if (trimmed.isEmpty()) return

        val userMsg = ChatMessage(role = "user", content = trimmed)
        _messages.value = _messages.value + userMsg

        val lower = trimmed.lowercase()
        val reply = when {
            lower.contains("uang") || lower.contains("pengeluaran") || lower.contains("budget") -> {
                ChatMessage(
                    role = "assistant",
                    content = "Pengeluaran F-Tracker tercatat Rp 4.250.000. Sisa budget bulan ini Rp 5.750.000.",
                    actions = listOf(
                        SuggestedAction("Detail Pengeluaran", deepLink = "ftracker://expenses?category=food&expected_uid=usr_ftracker_123")
                    )
                )
            }
            lower.contains("tugas") || lower.contains("task") -> {
                ChatMessage(
                    role = "assistant",
                    content = "Anda memiliki 5 tugas pending di Task Management. 2 deadline jatuh tempo hari ini.",
                    actions = listOf(
                        SuggestedAction("Buka Task Hub", deepLink = "taskhub://tasks?filter=today&expected_uid=usr_task_456")
                    )
                )
            }
            else -> {
                ChatMessage(
                    role = "assistant",
                    content = "Semua satelit Anda tersinkronisasi. Anda dapat langsung membuka salah satu aplikasi di bawah:",
                    actions = listOf(
                        SuggestedAction("Buka F-Tracker", deepLink = "ftracker://dashboard?expected_uid=usr_ftracker_123"),
                        SuggestedAction("Buka Task Management", deepLink = "taskhub://dashboard?expected_uid=usr_task_456")
                    )
                )
            }
        }

        _messages.value = _messages.value + reply
    }
}
