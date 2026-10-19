package cloud.eka_dev.eka_hub.feature.chat

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.eka_hub.core.deeplink.DeepLinkManager
import cloud.eka_dev.eka_hub.core.designsystem.*
import cloud.eka_dev.eka_hub.core.model.ChatMessage

@Composable
fun AiChatScreen(viewModel: AiChatViewModel) {
    val context = LocalContext.current
    val messages by viewModel.messages.collectAsState()
    var inputQuery by remember { mutableStateOf("") }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(HubDarkBg)
    ) {
        LazyColumn(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            items(messages) { msg ->
                val isUser = msg.role == "user"
                Column(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalAlignment = if (isUser) Alignment.End else Alignment.Start
                ) {
                    Surface(
                        color = if (isUser) HubIndigo else HubDarkCard,
                        shape = RoundedCornerShape(16.dp),
                        modifier = Modifier.border(1.dp, if (isUser) Color.Transparent else HubDarkBorder, RoundedCornerShape(16.dp))
                    ) {
                        Text(
                            text = msg.content,
                            fontSize = 14.sp,
                            color = Color.White,
                            modifier = Modifier.padding(14.dp)
                        )
                    }

                    if (msg.actions.isNotEmpty()) {
                        Spacer(modifier = Modifier.height(6.dp))
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
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
                                        color = HubCyan,
                                        modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp)
                                    )
                                }
                            }
                        }
                    }
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
                placeholder = { Text("Tanyakan pengeluaran atau tugas...", color = Color.Gray, fontSize = 13.sp) },
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
                    viewModel.sendMessage(inputQuery)
                    inputQuery = ""
                },
                modifier = Modifier
                    .background(HubIndigo, CircleShape)
                    .size(44.dp)
            ) {
                Icon(Icons.Default.Send, contentDescription = "Kirim", tint = Color.White)
            }
        }
    }
}
