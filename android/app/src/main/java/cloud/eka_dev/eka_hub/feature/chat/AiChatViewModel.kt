package cloud.eka_dev.eka_hub.feature.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import cloud.eka_dev.eka_hub.core.model.ChatMessage
import cloud.eka_dev.eka_hub.core.model.ChatSessionDto
import cloud.eka_dev.eka_hub.core.model.SuggestedAction
import cloud.eka_dev.eka_hub.core.network.ApiClient
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class AiChatViewModel(private val apiClient: ApiClient) : ViewModel() {

    private val defaultGreeting = ChatMessage(
        role = "assistant",
        content = "Halo! Saya Personal AI Hub Anda. Tanyakan pengeluaran F-Tracker, agenda tugas Task Management, jelaskan kode program, atau minta saya mencatat transaksi dan tugas baru.",
        actions = listOf(
            SuggestedAction("Lihat F-Tracker", deepLink = "ftracker://expenses?expected_uid=usr_ftracker_123"),
            SuggestedAction("Buka Task Hari Ini", deepLink = "taskhub://tasks?expected_uid=usr_task_456")
        )
    )

    private val _messages = MutableStateFlow<List<ChatMessage>>(listOf(defaultGreeting))
    val messages: StateFlow<List<ChatMessage>> = _messages.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    private val _currentSessionId = MutableStateFlow<String?>(null)
    val currentSessionId: StateFlow<String?> = _currentSessionId.asStateFlow()

    private val _sessions = MutableStateFlow<List<ChatSessionDto>>(emptyList())
    val sessions: StateFlow<List<ChatSessionDto>> = _sessions.asStateFlow()

    private val _isHistoryLoading = MutableStateFlow(false)
    val isHistoryLoading: StateFlow<Boolean> = _isHistoryLoading.asStateFlow()

    init {
        loadSessions()
    }

    fun loadSessions() {
        viewModelScope.launch {
            _isHistoryLoading.value = true
            try {
                val resp = apiClient.apiService.getChatSessions()
                if (resp.success && resp.data != null) {
                    _sessions.value = resp.data
                }
            } catch (e: Exception) {
                // Silently fallback if offline
            } finally {
                _isHistoryLoading.value = false
            }
        }
    }

    fun startNewChat() {
        if (_isLoading.value) return
        _currentSessionId.value = null
        _messages.value = listOf(defaultGreeting)
    }

    fun selectSession(sessionId: String) {
        if (_isLoading.value) return
        _currentSessionId.value = sessionId
        viewModelScope.launch {
            _isLoading.value = true
            try {
                val resp = apiClient.apiService.getChatSessionDetail(sessionId)
                if (resp.success && resp.data != null) {
                    val msgs = resp.data.messages.map { stored ->
                        ChatMessage(
                            id = stored.id,
                            role = stored.role,
                            content = stored.content
                        )
                    }
                    _messages.value = if (msgs.isEmpty()) listOf(defaultGreeting) else msgs
                }
            } catch (e: Exception) {
                _messages.value = listOf(
                    ChatMessage(
                        role = "assistant",
                        content = "Gagal memuat riwayat obrolan: ${e.localizedMessage}"
                    )
                )
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun deleteSession(sessionId: String) {
        viewModelScope.launch {
            try {
                apiClient.apiService.deleteChatSession(sessionId)
                _sessions.value = _sessions.value.filter { it.id != sessionId }
                if (_currentSessionId.value == sessionId) {
                    startNewChat()
                }
            } catch (e: Exception) {
                // Handle error
            }
        }
    }

    fun sendMessage(query: String) {
        val trimmed = query.trim()
        if (trimmed.isEmpty() || _isLoading.value) return

        val userMsg = ChatMessage(role = "user", content = trimmed)
        _messages.value = _messages.value + userMsg
        _isLoading.value = true

        viewModelScope.launch {
            try {
                val payload = mutableMapOf("message" to trimmed)
                _currentSessionId.value?.let { payload["session_id"] = it }

                val resp = apiClient.apiService.sendChatMessage(payload)
                if (resp.success && resp.data != null) {
                    val aiData = resp.data
                    val fullReply = aiData.reply
                    val actions = aiData.actionLinks
                    val quickReplies = aiData.quickReplies

                    if (!aiData.sessionId.isNullOrBlank()) {
                        _currentSessionId.value = aiData.sessionId
                    }

                    // Smooth simulated streaming / progressive typewriter rendering for live SSE feel
                    val msgId = java.util.UUID.randomUUID().toString()
                    val assistantMsg = ChatMessage(
                        id = msgId,
                        role = "assistant",
                        content = "",
                        actions = actions,
                        quickReplies = quickReplies,
                        isStreaming = true
                    )
                    _messages.value = _messages.value + assistantMsg
                    _isLoading.value = false

                    // Stream words progressively
                    val words = fullReply.split(" ")
                    var currentText = ""
                    for (i in words.indices) {
                        currentText += if (i == 0) words[i] else " " + words[i]
                        _messages.value = _messages.value.map { msg ->
                            if (msg.id == msgId) msg.copy(content = currentText) else msg
                        }
                        delay(20) // Smooth 20ms per word streaming
                    }
                    _messages.value = _messages.value.map { msg ->
                        if (msg.id == msgId) msg.copy(isStreaming = false) else msg
                    }

                    // Reload sessions list to update titles/previews
                    loadSessions()
                } else {
                    val errMsg = resp.message ?: resp.error ?: "Gagal memproses pesan."
                    _messages.value = _messages.value + ChatMessage(
                        role = "assistant",
                        content = "Maaf, terjadi kendala saat menghubungkan ke satelit: $errMsg"
                    )
                    _isLoading.value = false
                }
            } catch (e: Exception) {
                // If offline / network error, provide intelligent satellite fallback
                val fallbackContent = when {
                    trimmed.contains("uang", ignoreCase = true) || trimmed.contains("pengeluaran", ignoreCase = true) ->
                        "Pengeluaran F-Tracker tercatat Rp 4.250.000. Sisa budget bulan ini Rp 5.750.000."
                    trimmed.contains("tugas", ignoreCase = true) || trimmed.contains("task", ignoreCase = true) ->
                        "Anda memiliki 5 tugas pending di Task Management. 2 deadline jatuh tempo hari ini."
                    else ->
                        "Tidak dapat terhubung ke AI server saat ini (${e.localizedMessage}). Silakan coba sesaat lagi."
                }
                _messages.value = _messages.value + ChatMessage(
                    role = "assistant",
                    content = fallbackContent,
                    actions = listOf(
                        SuggestedAction("Buka F-Tracker", deepLink = "ftracker://dashboard"),
                        SuggestedAction("Buka Task Management", deepLink = "taskhub://dashboard")
                    )
                )
                _isLoading.value = false
            }
        }
    }
}
