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

    private val _lastFailedPrompt = MutableStateFlow<String?>(null)
    val lastFailedPrompt: StateFlow<String?> = _lastFailedPrompt.asStateFlow()

    private val _lastFailedMessageId = MutableStateFlow<String?>(null)
    val lastFailedMessageId: StateFlow<String?> = _lastFailedMessageId.asStateFlow()

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

    fun retryMessage(msgId: String) {
        val msg = _messages.value.find { it.id == msgId } ?: return
        sendMessage(msg.content, targetMessageId = msgId)
    }

    fun retryLastMessage() {
        val lastId = _lastFailedMessageId.value
        if (lastId != null) {
            retryMessage(lastId)
        } else {
            _lastFailedPrompt.value?.let { sendMessage(it) }
        }
    }

    fun clearFailedPrompt() {
        _lastFailedPrompt.value = null
        _lastFailedMessageId.value = null
    }

    fun sendMessage(query: String, targetMessageId: String? = null) {
        val trimmed = query.trim()
        if (trimmed.isEmpty() || _isLoading.value) return

        val userMsgId: String
        if (targetMessageId != null) {
            userMsgId = targetMessageId
            _messages.value = _messages.value.map {
                if (it.id == targetMessageId) it.copy(isFailed = false, errorMessage = null) else it
            }
        } else {
            val userMsg = ChatMessage(role = "user", content = trimmed)
            userMsgId = userMsg.id
            _messages.value = _messages.value + userMsg
        }
        _isLoading.value = true

        viewModelScope.launch {
            try {
                val payload = mutableMapOf("message" to trimmed)
                _currentSessionId.value?.let { payload["session_id"] = it }

                val resp = apiClient.apiService.sendChatMessage(payload)
                if (resp.success && resp.data != null) {
                    _lastFailedPrompt.value = null
                    _lastFailedMessageId.value = null

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
                    _lastFailedPrompt.value = trimmed
                    _lastFailedMessageId.value = userMsgId
                    _messages.value = _messages.value.map {
                        if (it.id == userMsgId) it.copy(isFailed = true, errorMessage = errMsg) else it
                    }
                    _isLoading.value = false
                }
            } catch (e: Exception) {
                _lastFailedPrompt.value = trimmed
                _lastFailedMessageId.value = userMsgId

                val errMsg = e.localizedMessage ?: "Koneksi terputus"
                val isNetworkError = errMsg.contains("timeout", ignoreCase = true) ||
                        errMsg.contains("connect", ignoreCase = true) ||
                        errMsg.contains("cancel", ignoreCase = true) ||
                        errMsg.contains("network", ignoreCase = true) ||
                        errMsg.contains("offline", ignoreCase = true) ||
                        errMsg.contains("host", ignoreCase = true) ||
                        errMsg.contains("broken pipe", ignoreCase = true)

                val userFriendly = if (isNetworkError) {
                    "Koneksi terputus saat aplikasi ditutup/berpindah. Ketuk Coba Lagi."
                } else {
                    "Gagal memproses pesan: $errMsg"
                }

                _messages.value = _messages.value.map {
                    if (it.id == userMsgId) it.copy(isFailed = true, errorMessage = userFriendly) else it
                }
                _isLoading.value = false
            }
        }
    }
}
