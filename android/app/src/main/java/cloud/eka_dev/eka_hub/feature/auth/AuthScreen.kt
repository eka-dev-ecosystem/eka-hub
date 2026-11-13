package cloud.eka_dev.eka_hub.feature.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.blur
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.eka_hub.core.datastore.TokenManager
import cloud.eka_dev.eka_hub.core.designsystem.*
import kotlinx.coroutines.launch

@Composable
fun AuthScreen(
    tokenManager: TokenManager,
    onLoginSuccess: () -> Unit
) {
    val scope = rememberCoroutineScope()

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(HubDarkBg),
        contentAlignment = Alignment.Center
    ) {
        // Ambient background glow
        Box(
            modifier = Modifier
                .size(300.dp)
                .offset(y = (-80).dp)
                .alpha(0.2f)
                .blur(70.dp)
                .background(
                    brush = Brush.radialGradient(
                        colors = listOf(HubIndigo, HubCyan, Color.Transparent)
                    ),
                    shape = CircleShape
                )
        )

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(28.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            // Brand Logo Box
            Box(
                modifier = Modifier
                    .size(76.dp)
                    .background(HubDarkCard, shape = RoundedCornerShape(22.dp))
                    .border(
                        width = 1.5.dp,
                        brush = Brush.linearGradient(
                            colors = listOf(HubCyan, HubIndigo, Color(0xFFD96570))
                        ),
                        shape = RoundedCornerShape(22.dp)
                    ),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    imageVector = Icons.Default.AutoAwesome,
                    contentDescription = "Personal AI Hub",
                    tint = HubCyan,
                    modifier = Modifier.size(36.dp)
                )
            }

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "Personal AI Hub",
                fontSize = 28.sp,
                fontWeight = FontWeight.Bold,
                color = Color.White
            )

            Spacer(modifier = Modifier.height(6.dp))

            Text(
                text = "Masuk untuk menyatukan ekosistem & data personal Anda",
                fontSize = 14.sp,
                color = Color(0xFF94A3B8),
                textAlign = TextAlign.Center,
                lineHeight = 20.sp
            )

            Spacer(modifier = Modifier.height(36.dp))

            Button(
                onClick = {
                    scope.launch {
                        tokenManager.saveTokens("hub_active_session_token", "hub_active_refresh_token")
                        tokenManager.saveUserInfo("user-1", "Eka Dev", "eka@eka-dev.cloud")
                        onLoginSuccess()
                    }
                },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(52.dp),
                colors = ButtonDefaults.buttonColors(containerColor = HubIndigo),
                shape = RoundedCornerShape(14.dp)
            ) {
                Text(
                    text = "Masuk dengan Google",
                    fontWeight = FontWeight.SemiBold,
                    fontSize = 15.sp,
                    color = Color.White
                )
            }
        }
    }
}
