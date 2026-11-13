package cloud.eka_dev.eka_hub.feature.chat

import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Send
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.blur
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.eka_hub.core.deeplink.DeepLinkManager
import cloud.eka_dev.eka_hub.core.designsystem.*
import cloud.eka_dev.eka_hub.core.model.ChatMessage

@Composable
fun AiChatScreen(viewModel: AiChatViewModel) {
    val context = LocalContext.current
    val messages by viewModel.messages.collectAsState()
    val isLoading by viewModel.isLoading.collectAsState()
    var inputQuery by remember { mutableStateOf("") }
    val listState = rememberLazyListState()

    // Auto-scroll when new messages arrive or when streaming updates text
    LaunchedEffect(messages.size, messages.lastOrNull()?.content?.length, isLoading) {
        val totalItems = messages.size + if (isLoading) 1 else 0
        if (totalItems > 0) {
            listState.animateScrollToItem(totalItems - 1)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(HubDarkBg)
    ) {
        LazyColumn(
            state = listState,
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(14.dp)
        ) {
            items(messages, key = { it.id }) { msg ->
                val isUser = msg.role == "user"
                Column(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalAlignment = if (isUser) Alignment.End else Alignment.Start
                ) {
                    Surface(
                        color = if (isUser) HubIndigo else HubDarkCard,
                        shape = RoundedCornerShape(16.dp),
                        modifier = Modifier.border(
                            1.dp,
                            if (isUser) Color.Transparent else HubDarkBorder,
                            RoundedCornerShape(16.dp)
                        )
                    ) {
                        Text(
                            text = msg.content,
                            fontSize = 14.sp,
                            color = Color.White,
                            lineHeight = 21.sp,
                            modifier = Modifier.padding(14.dp)
                        )
                    }

                    if (msg.actions.isNotEmpty()) {
                        Spacer(modifier = Modifier.height(6.dp))
                        Row(
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                            modifier = Modifier.padding(top = 4.dp)
                        ) {
                            msg.actions.forEach { action ->
                                Surface(
                                    color = HubIndigo.copy(alpha = 0.2f),
                                    shape = RoundedCornerShape(10.dp),
                                    modifier = Modifier
                                        .border(1.dp, HubIndigo.copy(alpha = 0.4f), RoundedCornerShape(10.dp))
                                        .clickable {
                                            DeepLinkManager.openChildApp(context, action.deepLink)
                                        }
                                ) {
                                    Text(
                                        text = action.title + " ↗",
                                        fontSize = 12.sp,
                                        fontWeight = FontWeight.Medium,
                                        color = HubCyan,
                                        modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp)
                                    )
                                }
                            }
                        }
                    }
                }
            }

            if (isLoading) {
                item(key = "thinking_indicator") {
                    AnimatedThinkingIndicator()
                }
            }
        }

        // Input bar
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(HubDarkCard)
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            TextField(
                value = inputQuery,
                onValueChange = { inputQuery = it },
                placeholder = {
                    Text(
                        "Tanyakan pengeluaran atau tugas hari ini...",
                        color = Color.Gray,
                        fontSize = 13.sp
                    )
                },
                colors = TextFieldDefaults.colors(
                    focusedContainerColor = Color.Transparent,
                    unfocusedContainerColor = Color.Transparent,
                    focusedTextColor = Color.White,
                    unfocusedTextColor = Color.White,
                    focusedIndicatorColor = Color.Transparent,
                    unfocusedIndicatorColor = Color.Transparent
                ),
                modifier = Modifier
                    .weight(1f)
                    .background(Color(0xFF0B1120), RoundedCornerShape(20.dp))
                    .border(1.dp, HubDarkBorder, RoundedCornerShape(20.dp))
                    .padding(horizontal = 8.dp)
            )

            Spacer(modifier = Modifier.width(8.dp))

            IconButton(
                onClick = {
                    val q = inputQuery
                    inputQuery = ""
                    viewModel.sendMessage(q)
                },
                enabled = inputQuery.isNotBlank() && !isLoading,
                modifier = Modifier
                    .background(if (inputQuery.isNotBlank() && !isLoading) HubIndigo else HubIndigo.copy(alpha = 0.4f), CircleShape)
                    .size(44.dp)
            ) {
                Icon(
                    Icons.AutoMirrored.Filled.Send,
                    contentDescription = "Kirim",
                    tint = Color.White
                )
            }
        }
    }
}

@Composable
fun AnimatedThinkingIndicator() {
    val infiniteTransition = rememberInfiniteTransition(label = "thinking")

    // Pulsing aura animation
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 0.95f,
        targetValue = 1.15f,
        animationSpec = infiniteRepeatable(
            animation = tween(1200, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "pulse"
    )

    // Bouncing dots offset animations (3 staggered dots)
    val dot1Offset by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = -7f,
        animationSpec = infiniteRepeatable(
            animation = tween(450, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "dot1"
    )
    val dot2Offset by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = -7f,
        animationSpec = infiniteRepeatable(
            animation = tween(450, delayMillis = 150, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "dot2"
    )
    val dot3Offset by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = -7f,
        animationSpec = infiniteRepeatable(
            animation = tween(450, delayMillis = 300, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "dot3"
    )

    // Silky smooth continuous shimmer sweep across capsule (starts and ends completely offscreen)
    val shimmerOffset by infiniteTransition.animateFloat(
        initialValue = -45f,
        targetValue = 140f,
        animationSpec = infiniteRepeatable(
            animation = tween(1400, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "shimmer"
    )

    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.Start
    ) {
        Surface(
            color = HubDarkCard,
            shape = RoundedCornerShape(16.dp),
            modifier = Modifier.border(1.dp, HubDarkBorder, RoundedCornerShape(16.dp))
        ) {
            Column(
                modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(10.dp)
                ) {
                    // Glowing pulsating AI sparkle icon
                    Box(
                        contentAlignment = Alignment.Center,
                        modifier = Modifier.size(24.dp)
                    ) {
                        Box(
                            modifier = Modifier
                                .size(24.dp)
                                .scale(pulseScale)
                                .alpha(0.4f)
                                .blur(6.dp)
                                .background(
                                    brush = Brush.radialGradient(listOf(HubCyan, HubIndigo, Color.Transparent)),
                                    shape = CircleShape
                                )
                        )
                        Icon(
                            imageVector = Icons.Default.AutoAwesome,
                            contentDescription = "Thinking",
                            tint = HubCyan,
                            modifier = Modifier.size(16.dp)
                        )
                    }

                    Text(
                        text = "Sedang berpikir & menganalisis data...",
                        fontSize = 13.sp,
                        fontWeight = FontWeight.Medium,
                        color = Color(0xFFE2E8F0)
                    )

                    // 3 Animated Bouncing Dots
                    Row(
                        horizontalArrangement = Arrangement.spacedBy(4.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        modifier = Modifier.padding(start = 4.dp)
                    ) {
                        Box(
                            modifier = Modifier
                                .size(5.dp)
                                .offset(y = dot1Offset.dp)
                                .background(HubCyan, CircleShape)
                        )
                        Box(
                            modifier = Modifier
                                .size(5.dp)
                                .offset(y = dot2Offset.dp)
                                .background(HubIndigo, CircleShape)
                        )
                        Box(
                            modifier = Modifier
                                .size(5.dp)
                                .offset(y = dot3Offset.dp)
                                .background(Color(0xFFD96570), CircleShape)
                        )
                    }
                }

                Spacer(modifier = Modifier.height(8.dp))

                // Silky smooth continuous shimmer bar
                Box(
                    modifier = Modifier
                        .width(140.dp)
                        .height(3.5.dp)
                        .clip(RoundedCornerShape(2.dp))
                        .background(Color(0xFF334155).copy(alpha = 0.5f), RoundedCornerShape(2.dp)),
                    contentAlignment = Alignment.CenterStart
                ) {
                    Box(
                        modifier = Modifier
                            .fillMaxHeight()
                            .width(45.dp)
                            .offset(x = shimmerOffset.dp)
                            .background(
                                brush = Brush.horizontalGradient(
                                    colors = listOf(
                                        Color.Transparent,
                                        HubCyan,
                                        HubIndigo,
                                        Color.Transparent
                                    )
                                )
                            )
                    )
                }
            }
        }
    }
}
