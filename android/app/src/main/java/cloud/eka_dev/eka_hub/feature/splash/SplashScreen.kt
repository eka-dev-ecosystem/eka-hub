package cloud.eka_dev.eka_hub.feature.splash

import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.blur
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.eka_hub.core.datastore.TokenManager
import cloud.eka_dev.eka_hub.core.designsystem.*
import kotlinx.coroutines.delay

@Composable
fun SplashScreen(
    tokenManager: TokenManager,
    onSplashFinished: (isLoggedIn: Boolean) -> Unit
) {
    // Entrance animations
    val entranceTransition = remember { Animatable(0f) }
    
    // Continuous breathing pulse for glowing aura
    val infiniteTransition = rememberInfiniteTransition(label = "pulse")
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 1f,
        targetValue = 1.08f,
        animationSpec = infiniteRepeatable(
            animation = tween(1800, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "pulseScale"
    )
    val pulseAlpha by infiniteTransition.animateFloat(
        initialValue = 0.35f,
        targetValue = 0.7f,
        animationSpec = infiniteRepeatable(
            animation = tween(1800, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "pulseAlpha"
    )

    // Shimmer sweep for capsule progress
    val shimmerOffset by infiniteTransition.animateFloat(
        initialValue = -1f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(1400, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "shimmerOffset"
    )

    var hasSession by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        // Trigger entrance choreography
        entranceTransition.animateTo(
            targetValue = 1f,
            animationSpec = spring(
                dampingRatio = Spring.DampingRatioMediumBouncy,
                stiffness = Spring.StiffnessLow
            )
        )
    }

    LaunchedEffect(Unit) {
        // Verify session in parallel
        val token = tokenManager.getAccessToken()
        hasSession = !token.isNullOrEmpty()

        // Minimum display duration for smooth brand animation
        delay(1400)
        onSplashFinished(hasSession)
    }

    val logoScale = 0.5f + (entranceTransition.value * 0.5f)
    val contentAlpha = entranceTransition.value.coerceIn(0f, 1f)

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(HubDarkBg),
        contentAlignment = Alignment.Center
    ) {
        // Ambient background glowing gradient orbs
        Box(
            modifier = Modifier
                .size(340.dp)
                .offset(y = (-60).dp)
                .alpha(0.28f)
                .blur(80.dp)
                .background(
                    brush = Brush.radialGradient(
                        colors = listOf(HubIndigo, HubCyan, Color.Transparent)
                    ),
                    shape = CircleShape
                )
        )

        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 32.dp, vertical = 48.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.weight(1.2f))

            // Center Brand Emblem with Ambient Pulsing Glow
            Box(
                contentAlignment = Alignment.Center,
                modifier = Modifier.scale(logoScale)
            ) {
                // Expanding / Breathing Aura Glow
                Box(
                    modifier = Modifier
                        .size(110.dp)
                        .scale(pulseScale)
                        .alpha(pulseAlpha)
                        .blur(20.dp)
                        .background(
                            brush = Brush.linearGradient(
                                colors = listOf(HubCyan, HubIndigo, Color(0xFFD96570))
                            ),
                            shape = RoundedCornerShape(32.dp)
                        )
                )

                // Concentric Ambient Ring
                Box(
                    modifier = Modifier
                        .size(116.dp)
                        .scale(pulseScale * 1.05f)
                        .alpha(pulseAlpha * 0.45f)
                        .border(
                            width = 1.dp,
                            brush = Brush.sweepGradient(
                                colors = listOf(HubCyan, HubIndigo, Color(0xFFD96570), HubCyan)
                            ),
                            shape = CircleShape
                        )
                )

                // Main Emblem Container
                Box(
                    modifier = Modifier
                        .size(86.dp)
                        .background(HubDarkCard, shape = RoundedCornerShape(26.dp))
                        .border(
                            width = 1.5.dp,
                            brush = Brush.linearGradient(
                                colors = listOf(HubCyan, HubIndigo, Color(0xFFD96570))
                            ),
                            shape = RoundedCornerShape(26.dp)
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        imageVector = Icons.Default.AutoAwesome,
                        contentDescription = "Personal AI Hub Logo",
                        tint = HubCyan,
                        modifier = Modifier
                            .size(42.dp)
                            .scale(if (pulseScale > 1.04f) 1.05f else 1f)
                    )
                }
            }

            Spacer(modifier = Modifier.height(28.dp))

            // Brand Titles & Tagline with smooth entrance fade
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                modifier = Modifier.alpha(contentAlpha)
            ) {
                Text(
                    text = "Personal AI Hub",
                    fontSize = 28.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color.White,
                    letterSpacing = 0.5.sp
                )

                Spacer(modifier = Modifier.height(8.dp))

                // Tagline Badge Pill
                Box(
                    modifier = Modifier
                        .background(
                            color = HubDarkSurface.copy(alpha = 0.7f),
                            shape = RoundedCornerShape(20.dp)
                        )
                        .border(
                            width = 1.dp,
                            color = Color.White.copy(alpha = 0.1f),
                            shape = RoundedCornerShape(20.dp)
                        )
                        .padding(horizontal = 14.dp, vertical = 6.dp)
                ) {
                    Text(
                        text = "UNIFIED AI ECOSYSTEM",
                        fontSize = 11.sp,
                        fontWeight = FontWeight.SemiBold,
                        color = Color(0xFF94A3B8),
                        letterSpacing = 1.2.sp
                    )
                }
            }

            Spacer(modifier = Modifier.weight(1f))

            // Modern Fluid Shimmer Capsule Loader
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                modifier = Modifier
                    .alpha(contentAlpha)
                    .padding(bottom = 16.dp)
            ) {
                Box(
                    modifier = Modifier
                        .width(130.dp)
                        .height(3.5.dp)
                        .background(
                            color = HubDarkSurface.copy(alpha = 0.8f),
                            shape = RoundedCornerShape(4.dp)
                        ),
                    contentAlignment = Alignment.CenterStart
                ) {
                    val progressFraction = ((shimmerOffset + 1f) / 2f).coerceIn(0f, 1f)
                    Box(
                        modifier = Modifier
                            .fillMaxHeight()
                            .width(50.dp)
                            .offset(x = (progressFraction * 80).dp)
                            .background(
                                brush = Brush.horizontalGradient(
                                    colors = listOf(
                                        Color.Transparent,
                                        HubCyan,
                                        HubIndigo,
                                        Color.Transparent
                                    )
                                ),
                                shape = RoundedCornerShape(4.dp)
                            )
                    )
                }

                Spacer(modifier = Modifier.height(12.dp))

                Text(
                    text = "INITIALIZING ECOSYSTEM...",
                    fontSize = 10.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color(0xFF64748B),
                    letterSpacing = 2.sp
                )
            }
        }
    }
}
