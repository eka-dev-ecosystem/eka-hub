package cloud.eka_dev.eka_hub.feature.dashboard

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.eka_hub.core.deeplink.DeepLinkManager
import cloud.eka_dev.eka_hub.core.designsystem.*
import cloud.eka_dev.eka_hub.core.model.AppSummaryResponse

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DashboardScreen(
    viewModel: DashboardViewModel,
    onNavigateToChat: () -> Unit = {}
) {
    val context = LocalContext.current
    val briefing by viewModel.dailyBriefing.collectAsState()
    val summaries by viewModel.summaries.collectAsState()

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(HubDarkBg)
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column {
                    Text(
                        text = "Personal AI Hub",
                        fontSize = 22.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color.White
                    )
                    Text(
                        text = "Life Aggregator & Intelligent Assistant",
                        fontSize = 12.sp,
                        color = Color(0xFF94A3B8)
                    )
                }

                IconButton(onClick = onNavigateToChat) {
                    Icon(
                        imageVector = Icons.Default.ChatBubble,
                        contentDescription = "Chat",
                        tint = HubIndigo
                    )
                }
            }
        }

        item {
            GradientAiBanner(
                title = "Daily AI Briefing",
                content = briefing
            )
        }

        item {
            Text(
                text = "Aplikasi Satelit Terhubung",
                fontSize = 16.sp,
                fontWeight = FontWeight.SemiBold,
                color = Color.White
            )
        }

        item {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                summaries.forEach { summary ->
                    BentoCard(
                        modifier = Modifier
                            .weight(1f)
                            .height(160.dp),
                        onClick = {
                            DeepLinkManager.openChildApp(context, summary.deepLink)
                        }
                    ) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Icon(
                                imageVector = if (summary.appId.contains("tracker")) Icons.Default.AccountBalanceWallet else Icons.Default.CheckCircle,
                                contentDescription = summary.appName,
                                tint = if (summary.appId.contains("tracker")) HubEmerald else HubIndigo,
                                modifier = Modifier.size(24.dp)
                            )
                            Icon(
                                imageVector = Icons.Default.ArrowOutward,
                                contentDescription = "Open",
                                tint = Color(0xFF64748B),
                                modifier = Modifier.size(16.dp)
                            )
                        }

                        Spacer(modifier = Modifier.weight(1f))

                        Text(
                            text = summary.appName,
                            fontSize = 12.sp,
                            color = Color(0xFF94A3B8)
                        )
                        Text(
                            text = summary.highlightMetric,
                            fontSize = 16.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color.White
                        )
                        Text(
                            text = summary.highlightLabel,
                            fontSize = 11.sp,
                            color = Color(0xFF64748B)
                        )
                    }
                }
            }
        }
    }
}
