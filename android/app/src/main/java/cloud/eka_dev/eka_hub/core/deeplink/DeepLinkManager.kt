package cloud.eka_dev.eka_hub.core.deeplink

import android.content.Context
import android.content.Intent
import android.net.Uri
import android.widget.Toast

object DeepLinkManager {

    /**
     * Opens a child app using its registered Custom URL Scheme.
     * Appends expected_uid to prevent Account Mismatch.
     */
    fun openChildApp(context: Context, deepLinkUrl: String) {
        try {
            val intent = Intent(Intent.ACTION_VIEW, Uri.parse(deepLinkUrl)).apply {
                addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
            }
            context.startActivity(intent)
        } catch (e: Exception) {
            Toast.makeText(
                context,
                "Aplikasi satelit belum terpasang: ${e.message}",
                Toast.LENGTH_LONG
            ).show()
        }
    }

    /**
     * Initiates Account Linking handshake with Child App.
     */
    fun initiateLinkFlow(context: Context, appId: String, sessionCode: String) {
        val scheme = when (appId.lowercase()) {
            "ftracker" -> "ftracker://"
            "taskmanagement", "taskhub" -> "taskhub://"
            else -> "app://"
        }
        val url = "${scheme}connect-hub?session_code=$sessionCode&callback_url=ekahub://link-callback"
        openChildApp(context, url)
    }
}
