package cloud.eka_dev.eka_hub.core.designsystem

import android.app.Activity
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.SideEffect
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.platform.LocalView
import androidx.core.view.WindowCompat

val HubIndigo = Color(0xFF6366F1)
val HubIndigoDark = Color(0xFF4F46E5)
val HubCyan = Color(0xFF06B6D4)
val HubEmerald = Color(0xFF10B981)
val HubAmber = Color(0xFFF59E0B)
val HubDarkBg = Color(0xFF0F172A)
val HubDarkSurface = Color(0xFF1E293B)
val HubDarkCard = Color(0xFF161E2E)
val HubDarkBorder = Color(0xFF334155)

private val DarkColorScheme = darkColorScheme(
    primary = HubIndigo,
    secondary = HubCyan,
    tertiary = HubEmerald,
    background = HubDarkBg,
    surface = HubDarkSurface,
    onPrimary = Color.White,
    onSecondary = Color.White,
    onBackground = Color.White,
    onSurface = Color.White
)

private val LightColorScheme = lightColorScheme(
    primary = HubIndigoDark,
    secondary = HubCyan,
    tertiary = HubEmerald,
    background = Color(0xFFF8FAFC),
    surface = Color.White,
    onPrimary = Color.White,
    onSecondary = Color.Black,
    onBackground = Color(0xFF0F172A),
    onSurface = Color(0xFF0F172A)
)

@Composable
fun EkaHubTheme(
    darkTheme: Boolean = true, // Default dark modern aesthetics
    content: @Composable () -> Unit
) {
    val colorScheme = if (darkTheme) DarkColorScheme else LightColorScheme
    val view = LocalView.current
    if (!view.isInEditMode) {
        SideEffect {
            val window = (view.context as Activity).window
            window.statusBarColor = colorScheme.background.toArgb()
            WindowCompat.getInsetsController(window, view).isAppearanceLightStatusBars = !darkTheme
        }
    }

    MaterialTheme(
        colorScheme = colorScheme,
        content = content
    )
}
