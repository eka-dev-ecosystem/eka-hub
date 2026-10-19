package cloud.eka_dev.eka_hub.core.guard

import android.net.Uri
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.WarningAmber
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.eka_hub.core.model.AccountMismatchState

object ChildAppDeepLinkGuard {

    private var pendingDeepLink: Uri? = null

    sealed class GuardResult {
        data class Allow(val targetUri: Uri) : GuardResult()
        data class RequireLogin(val pendingUri: Uri, val hintMessage: String) : GuardResult()
        data class ShowMismatchDialog(val state: AccountMismatchState) : GuardResult()
    }

    fun evaluate(
        incomingUri: Uri,
        currentUid: String?,
        currentEmail: String?
    ): GuardResult {
        val expectedUid = incomingUri.getQueryParameter("expected_uid")

        if (expectedUid.isNullOrEmpty()) {
            return GuardResult.Allow(incomingUri)
        }

        if (currentUid.isNullOrEmpty()) {
            pendingDeepLink = incomingUri
            val hint = "Tautan berasal dari Personal AI Hub. Harap login dengan akun yang terhubung."
            return GuardResult.RequireLogin(incomingUri, hint)
        }

        if (currentUid == expectedUid) {
            return GuardResult.Allow(incomingUri)
        }

        val mismatchState = AccountMismatchState(
            isMismatch = true,
            expectedUid = expectedUid,
            expectedEmailHint = incomingUri.getQueryParameter("expected_email") ?: "terdaftar di Hub ($expectedUid)",
            currentUid = currentUid,
            currentEmail = currentEmail ?: "Akun Aktif",
            targetDeepLinkUri = incomingUri
        )
        return GuardResult.ShowMismatchDialog(mismatchState)
    }

    fun consumePendingDeepLink(): Uri? {
        val link = pendingDeepLink
        pendingDeepLink = null
        return link
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AccountMismatchBottomSheet(
    state: AccountMismatchState,
    onSwitchAccount: () -> Unit,
    onUpdateHubConnection: () -> Unit,
    onCancel: () -> Unit
) {
    ModalBottomSheet(
        onDismissRequest = onCancel,
        containerColor = Color(0xFF1E293B),
        shape = RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(24.dp)
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    imageVector = Icons.Default.WarningAmber,
                    contentDescription = "Warning",
                    tint = Color(0xFFF59E0B),
                    modifier = Modifier.size(28.dp)
                )
                Spacer(modifier = Modifier.width(10.dp))
                Text(
                    text = "Akun Tidak Sesuai",
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
            }

            Spacer(modifier = Modifier.height(12.dp))

            Text(
                text = "Tautan dari Hub meminta akun:",
                fontSize = 13.sp,
                color = Color(0xFF94A3B8)
            )
            Text(
                text = state.expectedEmailHint,
                fontSize = 14.sp,
                fontWeight = FontWeight.SemiBold,
                color = Color(0xFF38BDF8)
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "Sedangkan akun yang sedang aktif saat ini:",
                fontSize = 13.sp,
                color = Color(0xFF94A3B8)
            )
            Text(
                text = state.currentEmail,
                fontSize = 14.sp,
                fontWeight = FontWeight.SemiBold,
                color = Color(0xFF10B981)
            )

            Spacer(modifier = Modifier.height(24.dp))

            Button(
                onClick = onSwitchAccount,
                modifier = Modifier.fillMaxWidth(),
                colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF4F46E5)),
                shape = RoundedCornerShape(12.dp)
            ) {
                Text(
                    text = "1. Ganti ke Akun Terkait Hub",
                    fontWeight = FontWeight.SemiBold,
                    color = Color.White
                )
            }

            Spacer(modifier = Modifier.height(8.dp))

            OutlinedButton(
                onClick = onUpdateHubConnection,
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.outlinedButtonColors(contentColor = Color(0xFF38BDF8))
            ) {
                Text(
                    text = "2. Perbarui Koneksi Hub ke Akun Ini",
                    fontWeight = FontWeight.SemiBold
                )
            }

            Spacer(modifier = Modifier.height(8.dp))

            TextButton(
                onClick = onCancel,
                modifier = Modifier.fillMaxWidth()
            ) {
                Text(
                    text = "3. Batal (Tetap di Akun Saat Ini)",
                    color = Color(0xFF94A3B8)
                )
            }
        }
    }
}
