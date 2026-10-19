package cloud.eka_dev.eka_hub.feature.connectedapps

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.eka_hub.core.deeplink.DeepLinkManager
import cloud.eka_dev.eka_hub.core.designsystem.*

@Composable
fun ConnectedAppsScreen() {
    val context = LocalContext.current

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(HubDarkBg)
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            Text(
                text = "Aplikasi Satelit Terhubung",
                fontSize = 20.sp,
                fontWeight = FontWeight.Bold,
                color = Color.White
            )
            Text(
                text = "Kelola koneksi OAuth & sinkronisasi data antar aplikasi",
                fontSize = 12.sp,
                color = Color(0xFF94A3B8)
            )
        }

        item {
            BentoCard(
                modifier = Modifier.fillMaxWidth()
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        imageVector = Icons.Default.AccountBalanceWallet,
                        contentDescription = "F-Tracker",
                        tint = HubEmerald,
                        modifier = Modifier.size(32.dp)
                    )
                    Spacer(modifier = Modifier.width(12.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text("F-Tracker", fontWeight = FontWeight.Bold, color = Color.White)
                        Text("Terhubung: eka@finance.app", fontSize = 12.sp, color = HubCyan)
                    }
                    Button(
                        onClick = {
                            val sessionCode = "sess_${System.currentTimeMillis()}"
                            DeepLinkManager.initiateLinkFlow(context, "ftracker", sessionCode)
                        },
                        colors = ButtonDefaults.buttonColors(containerColor = HubIndigo.copy(alpha = 0.3f))
                    ) {
                        Text("Sync", color = Color.White, fontSize = 12.sp)
                    }
                }
            }
        }

        item {
            BentoCard(
                modifier = Modifier.fillMaxWidth()
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        imageVector = Icons.Default.CheckCircle,
                        contentDescription = "Task Management",
                        tint = HubIndigo,
                        modifier = Modifier.size(32.dp)
                    )
                    Spacer(modifier = Modifier.width(12.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text("Task Management", fontWeight = FontWeight.Bold, color = Color.White)
                        Text("Terhubung: eka@taskhub.app", fontSize = 12.sp, color = HubCyan)
                    }
                    Button(
                        onClick = {
                            val sessionCode = "sess_${System.currentTimeMillis()}"
                            DeepLinkManager.initiateLinkFlow(context, "taskhub", sessionCode)
                        },
                        colors = ButtonDefaults.buttonColors(containerColor = HubIndigo.copy(alpha = 0.3f))
                    ) {
                        Text("Sync", color = Color.White, fontSize = 12.sp)
                    }
                }
            }
        }
    }
}
