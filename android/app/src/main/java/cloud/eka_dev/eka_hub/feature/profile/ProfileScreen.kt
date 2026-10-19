package cloud.eka_dev.eka_hub.feature.profile

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Person
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.eka_hub.core.designsystem.*

@Composable
fun ProfileScreen(onSignOut: () -> Unit = {}) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(HubDarkBg)
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Surface(
            color = HubIndigo,
            shape = CircleShape,
            modifier = Modifier.size(80.dp)
        ) {
            Box(contentAlignment = Alignment.Center) {
                Icon(Icons.Default.Person, contentDescription = "User", tint = Color.White, modifier = Modifier.size(48.dp))
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        Text("Eka Dev", fontSize = 20.sp, fontWeight = FontWeight.Bold, color = Color.White)
        Text("eka@eka-dev.cloud", fontSize = 14.sp, color = Color(0xFF94A3B8))

        Spacer(modifier = Modifier.height(32.dp))

        BentoCard(modifier = Modifier.fillMaxWidth()) {
            Text("Pengaturan Hub", fontWeight = FontWeight.SemiBold, color = Color.White)
            Spacer(modifier = Modifier.height(8.dp))
            Text("Versi Aplikasi: 1.0.0 (Build 1)", fontSize = 13.sp, color = Color.Gray)
            Text("Lingkungan: Production Cloud", fontSize = 13.sp, color = Color.Gray)
        }

        Spacer(modifier = Modifier.weight(1f))

        OutlinedButton(
            onClick = onSignOut,
            modifier = Modifier.fillMaxWidth(),
            colors = ButtonDefaults.outlinedButtonColors(contentColor = Color(0xFFEF4444))
        ) {
            Text("Keluar dari Akun")
        }
    }
}
