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
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.History
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Warning
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
    val currentSessionId by viewModel.currentSessionId.collectAsState()
    val sessions by viewModel.sessions.collectAsState()
    val isHistoryLoading by viewModel.isHistoryLoading.collectAsState()
    val lastFailedPrompt by viewModel.lastFailedPrompt.collectAsState()
    var showHistorySheet by remember { mutableStateOf(false) }

    var inputQuery by remember { mutableStateOf("") }
    val listState = rememberLazyListState()

    // Auto-scroll when new messages arrive or when streaming updates text
    LaunchedEffect(messages.size, messages.lastOrNull()?.content?.length, isLoading) {
        val totalItems = messages.size + if (isLoading) 1 else 0
        if (totalItems > 0) {
            listState.animateScrollToItem(totalItems - 1)
        }
    }

    if (showHistorySheet) {
        ChatHistorySheet(
            sessions = sessions,
            currentSessionId = currentSessionId,
            isLoading = isHistoryLoading,
            onSelectSession = { sid -> viewModel.selectSession(sid) },
            onNewChat = { viewModel.startNewChat() },
            onDeleteSession = { sid -> viewModel.deleteSession(sid) },
            onDismiss = { showHistorySheet = false }
        )
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(HubDarkBg)
    ) {
        // Gemini-Grade Top Bar
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(HubDarkCard.copy(alpha = 0.95f))
                .border(1.dp, HubDarkBorder, RoundedCornerShape(0.dp))
                .padding(horizontal = 14.dp, vertical = 10.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            IconButton(
                onClick = {
                    viewModel.loadSessions()
                    showHistorySheet = true
                },
                modifier = Modifier.size(36.dp)
            ) {
                Icon(
                    imageVector = Icons.Default.History,
                    contentDescription = "Riwayat Obrolan",
                    tint = Color.White,
                    modifier = Modifier.size(22.dp)
                )
            }

            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Icon(
                    imageVector = Icons.Default.AutoAwesome,
                    contentDescription = null,
                    tint = HubCyan,
                    modifier = Modifier.size(18.dp)
                )
                Text(
                    text = "Eka-Hub AI",
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
            }

            IconButton(
                onClick = { viewModel.startNewChat() },
                modifier = Modifier.size(36.dp)
            ) {
                Icon(
                    imageVector = Icons.Default.Add,
                    contentDescription = "Chat Baru",
                    tint = Color.White,
                    modifier = Modifier.size(22.dp)
                )
            }
        }

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
                val isLatest = msg.id == messages.lastOrNull()?.id
                Column(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalAlignment = if (isUser) Alignment.End else Alignment.Start
                ) {
                    Surface(
                        color = when {
                            isUser && msg.isFailed -> Color(0xFF7F1D1D).copy(alpha = 0.55f)
                            isUser -> HubIndigo
                            else -> HubDarkCard
                        },
                        shape = RoundedCornerShape(16.dp),
                        modifier = Modifier.border(
                            1.dp,
                            when {
                                isUser && msg.isFailed -> Color(0xFFEF4444).copy(alpha = 0.6f)
                                isUser -> Color.Transparent
                                else -> HubDarkBorder
                            },
                            RoundedCornerShape(16.dp)
                        )
                    ) {
                        if (isUser) {
                            Text(
                                text = msg.content,
                                fontSize = 14.sp,
                                color = Color.White,
                                lineHeight = 21.sp,
                                modifier = Modifier.padding(14.dp)
                            )
                        } else {
                            Box(modifier = Modifier.padding(14.dp)) {
                                GeminiMarkdownContent(
                                    content = msg.content,
                                    textColor = Color(0xFFF1F5F9)
                                )
                            }
                        }
                    }

                    if (isUser && msg.isFailed) {
                        Spacer(modifier = Modifier.height(6.dp))
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                            modifier = Modifier.padding(top = 2.dp)
                        ) {
                            Icon(
                                imageVector = Icons.Default.Warning,
                                contentDescription = null,
                                tint = Color(0xFFF87171),
                                modifier = Modifier.size(13.dp)
                            )
                            Text(
                                text = msg.errorMessage ?: "Gagal terkirim",
                                fontSize = 11.sp,
                                color = Color(0xFFF87171)
                            )
                            Surface(
                                color = Color(0xFFEF4444).copy(alpha = 0.2f),
                                shape = RoundedCornerShape(8.dp),
                                modifier = Modifier
                                    .border(1.dp, Color(0xFFEF4444).copy(alpha = 0.5f), RoundedCornerShape(8.dp))
                                    .alpha(if (isLoading) 0.45f else 1f)
                                    .clickable(enabled = !isLoading) {
                                        viewModel.retryMessage(msg.id)
                                    }
                            ) {
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(4.dp),
                                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp)
                                ) {
                                    Icon(
                                        imageVector = Icons.Default.Refresh,
                                        contentDescription = null,
                                        tint = Color.White,
                                        modifier = Modifier.size(11.dp)
                                    )
                                    Text(
                                        text = "Coba Lagi",
                                        fontSize = 11.sp,
                                        fontWeight = FontWeight.SemiBold,
                                        color = Color.White
                                    )
                                }
                            }
                            Text(
                                text = "Salin",
                                fontSize = 11.sp,
                                fontWeight = FontWeight.SemiBold,
                                color = HubCyan,
                                modifier = Modifier
                                    .clickable {
                                        inputQuery = msg.content
                                    }
                                    .padding(horizontal = 4.dp, vertical = 2.dp)
                            )
                        }
                    }

                    if (isLatest && !isLoading && msg.quickReplies.isNotEmpty() && !msg.isStreaming) {
                        Spacer(modifier = Modifier.height(8.dp))
                        Row(
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                            modifier = Modifier.padding(top = 2.dp)
                        ) {
                            msg.quickReplies.forEach { reply ->
                                val isPositive = reply.contains("Ya", ignoreCase = true) || reply.contains("✅")
                                Surface(
                                    color = if (isPositive) Color(0xFF10B981).copy(alpha = 0.2f) else Color(0xFFEF4444).copy(alpha = 0.2f),
                                    shape = RoundedCornerShape(12.dp),
                                    modifier = Modifier
                                        .border(
                                            1.dp,
                                            if (isPositive) Color(0xFF10B981).copy(alpha = 0.6f) else Color(0xFFEF4444).copy(alpha = 0.5f),
                                            RoundedCornerShape(12.dp)
                                        )
                                        .alpha(if (isLoading) 0.45f else 1f)
                                        .clickable(enabled = !isLoading) {
                                            viewModel.sendMessage(reply)
                                        }
                                ) {
                                    Text(
                                        text = reply,
                                        fontSize = 13.sp,
                                        fontWeight = FontWeight.SemiBold,
                                        color = if (isPositive) Color(0xFF34D399) else Color(0xFFF87171),
                                        modifier = Modifier.padding(horizontal = 14.dp, vertical = 8.dp)
                                    )
                                }
                            }
                        }
                    }

                    if (msg.actions.isNotEmpty()) {
                        Spacer(modifier = Modifier.height(6.dp))
                        Column(
                            verticalArrangement = Arrangement.spacedBy(6.dp),
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
                                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp)
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

        if (lastFailedPrompt != null) {
            Surface(
                color = Color(0xFF1F1315),
                shape = RoundedCornerShape(12.dp),
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 4.dp)
                    .border(1.dp, Color(0xFFEF4444).copy(alpha = 0.4f), RoundedCornerShape(12.dp))
            ) {
                Row(
                    modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        imageVector = Icons.Default.Warning,
                        contentDescription = null,
                        tint = Color(0xFFF87171),
                        modifier = Modifier.size(15.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "Pesan gagal dikirim (Koneksi terputus)",
                        fontSize = 12.sp,
                        color = Color(0xFFFCA5A5),
                        modifier = Modifier.weight(1f)
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Surface(
                        color = Color(0xFFEF4444),
                        shape = RoundedCornerShape(8.dp),
                        modifier = Modifier
                            .alpha(if (isLoading) 0.5f else 1f)
                            .clickable(enabled = !isLoading) {
                                viewModel.retryLastMessage()
                            }
                    ) {
                        Text(
                            text = "Coba Lagi",
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color.White,
                            modifier = Modifier.padding(horizontal = 10.dp, vertical = 5.dp)
                        )
                    }
                    Spacer(modifier = Modifier.width(6.dp))
                    Icon(
                        imageVector = Icons.Default.Close,
                        contentDescription = "Tutup",
                        tint = Color.Gray,
                        modifier = Modifier
                            .size(16.dp)
                            .clickable {
                                viewModel.clearFailedPrompt()
                            }
                    )
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
                    if (!isLoading && inputQuery.isNotBlank()) {
                        val q = inputQuery
                        inputQuery = ""
                        viewModel.sendMessage(q)
                    }
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

    // Fluid carousel shimmer sweep with full edge exit behavior
    val shimmerProgress by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(1250, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "shimmerProgress"
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

                // Fluid Carousel Shimmer Capsule Loader (full edge exit, zero pop)
                Box(
                    modifier = Modifier
                        .width(140.dp)
                        .height(3.5.dp)
                        .clip(RoundedCornerShape(2.dp))
                        .background(Color(0xFF334155).copy(alpha = 0.5f), RoundedCornerShape(2.dp)),
                    contentAlignment = Alignment.CenterStart
                ) {
                    val beamBrush = Brush.horizontalGradient(
                        colors = listOf(
                            Color.Transparent,
                            HubCyan.copy(alpha = 0.85f),
                            HubIndigo.copy(alpha = 0.85f),
                            Color.Transparent
                        )
                    )
                    // Carousel progression: track width = 140dp, beam width = 56dp, total cycle = 196dp
                    val currentOffset = shimmerProgress * 196f
                    val primaryX = -56f + currentOffset
                    val followX = primaryX - 196f
                    val exitX = primaryX + 196f

                    listOf(primaryX, followX, exitX).forEach { posX ->
                        Box(
                            modifier = Modifier
                                .width(56.dp)
                                .fillMaxHeight()
                                .offset(x = posX.dp)
                                .background(beamBrush)
                        )
                    }
                }
            }
        }
    }
}
