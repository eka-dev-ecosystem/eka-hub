package cloud.eka_dev.eka_hub.core.model

import android.net.Uri
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonObject

@Serializable
data class BaseResponse<T>(
    val success: Boolean,
    val message: String? = null,
    val data: T? = null,
    val error: String? = null
)

@Serializable
data class User(
    val id: String,
    val name: String,
    val email: String,
    val avatar: String? = null
)

@Serializable
data class AuthResponse(
    val user: User,
    @SerialName("access_token") val accessToken: String,
    @SerialName("refresh_token") val refreshToken: String
)

@Serializable
data class TokenPair(
    @SerialName("access_token") val accessToken: String,
    @SerialName("refresh_token") val refreshToken: String
)

@Serializable
data class AppSummaryResponse(
    @SerialName("app_id") val appId: String,
    @SerialName("app_name") val appName: String,
    @SerialName("connected_uid") val connectedUid: String,
    @SerialName("connected_email") val connectedEmail: String,
    @SerialName("highlight_metric") val highlightMetric: String,
    @SerialName("highlight_label") val highlightLabel: String,
    @SerialName("ai_context_data") val aiContextData: JsonObject? = null,
    @SerialName("deep_link") val deepLink: String
)

@Serializable
data class SuggestedAction(
    val title: String,
    val description: String? = null,
    @SerialName("app_id") val appId: String? = null,
    @SerialName("deep_link") val deepLink: String
)

@Serializable
data class DailyBriefingResponse(
    @SerialName("generated_at") val generatedAt: String? = null,
    @SerialName("daily_ai_briefing") val dailyAiBriefing: String,
    @SerialName("mini_summary_cards") val miniSummaryCards: List<AppSummaryResponse> = emptyList(),
    @SerialName("suggested_actions") val suggestedActions: List<SuggestedAction> = emptyList()
)

@Serializable
data class ToolCallDetail(
    val name: String = "",
    val arguments: JsonObject? = null
)

@Serializable
data class AIChatResponse(
    val reply: String = "",
    @SerialName("session_id") val sessionId: String? = null,
    @SerialName("tool_calls") val toolCalls: List<ToolCallDetail> = emptyList(),
    @SerialName("action_links") val actionLinks: List<SuggestedAction> = emptyList(),
    @SerialName("quick_replies") val quickReplies: List<String> = emptyList()
)

@Serializable
data class ChatSessionDto(
    val id: String,
    @SerialName("hub_user_id") val hubUserId: String,
    val title: String,
    @SerialName("created_at") val createdAt: String,
    @SerialName("updated_at") val updatedAt: String,
    @SerialName("last_preview") val lastPreview: String = "",
    @SerialName("message_count") val messageCount: Int = 0
)

@Serializable
data class StoredChatMessageDto(
    val id: String,
    @SerialName("hub_user_id") val hubUserId: String,
    @SerialName("session_id") val sessionId: String? = null,
    val role: String,
    val content: String,
    @SerialName("created_at") val createdAt: String
)

@Serializable
data class ChatSessionDetailDto(
    val id: String,
    @SerialName("hub_user_id") val hubUserId: String,
    val title: String,
    @SerialName("created_at") val createdAt: String,
    @SerialName("updated_at") val updatedAt: String,
    val messages: List<StoredChatMessageDto> = emptyList()
)

data class ChatMessage(
    val id: String = java.util.UUID.randomUUID().toString(),
    val role: String, // "user" or "assistant"
    val content: String,
    val actions: List<SuggestedAction> = emptyList(),
    val quickReplies: List<String> = emptyList(),
    val isStreaming: Boolean = false,
    val isFailed: Boolean = false,
    val errorMessage: String? = null
)

data class AccountMismatchState(
    val isMismatch: Boolean = false,
    val expectedUid: String = "",
    val expectedEmailHint: String = "",
    val currentUid: String = "",
    val currentEmail: String = "",
    val targetDeepLinkUri: Uri? = null
)
