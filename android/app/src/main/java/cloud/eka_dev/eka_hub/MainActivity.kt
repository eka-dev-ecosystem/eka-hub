package cloud.eka_dev.eka_hub

import android.content.Intent
import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.compose.*
import cloud.eka_dev.eka_hub.core.deeplink.DeepLinkManager
import cloud.eka_dev.eka_hub.core.designsystem.*
import cloud.eka_dev.eka_hub.core.guard.AccountMismatchBottomSheet
import cloud.eka_dev.eka_hub.core.guard.ChildAppDeepLinkGuard
import cloud.eka_dev.eka_hub.core.model.AccountMismatchState
import cloud.eka_dev.eka_hub.feature.auth.AuthScreen
import cloud.eka_dev.eka_hub.feature.chat.AiChatScreen
import cloud.eka_dev.eka_hub.feature.chat.AiChatViewModel
import cloud.eka_dev.eka_hub.feature.connectedapps.ConnectedAppsScreen
import cloud.eka_dev.eka_hub.feature.dashboard.DashboardScreen
import cloud.eka_dev.eka_hub.feature.dashboard.DashboardViewModel
import cloud.eka_dev.eka_hub.feature.profile.ProfileScreen
import cloud.eka_dev.eka_hub.feature.splash.SplashScreen
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {

    private var mismatchState by mutableStateOf<AccountMismatchState?>(null)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val app = application as EkaHubApp

        val dashboardViewModel = DashboardViewModel(app.apiClient)
        val aiChatViewModel = AiChatViewModel(app.apiClient)

        handleIncomingIntent(intent)

        setContent {
            EkaHubTheme {
                val navController = rememberNavController()
                val navBackStackEntry by navController.currentBackStackEntryAsState()
                val currentRoute = navBackStackEntry?.destination?.route
                val coroutineScope = rememberCoroutineScope()

                val showBottomBar = currentRoute in listOf("dashboard", "chat", "apps", "profile")

                Scaffold(
                    containerColor = HubDarkBg,
                    bottomBar = {
                        if (showBottomBar) {
                            NavigationBar(
                                containerColor = HubDarkSurface,
                                contentColor = Color.White
                            ) {
                                NavigationBarItem(
                                    selected = currentRoute == "dashboard",
                                    onClick = {
                                        navController.navigate("dashboard") {
                                            popUpTo(navController.graph.findStartDestination().id) { saveState = true }
                                            launchSingleTop = true
                                            restoreState = true
                                        }
                                    },
                                    icon = { Icon(Icons.Default.Dashboard, contentDescription = "Dashboard") },
                                    label = { Text("Dashboard") },
                                    colors = NavigationBarItemDefaults.colors(
                                        selectedIconColor = HubCyan,
                                        selectedTextColor = HubCyan,
                                        indicatorColor = HubIndigo.copy(alpha = 0.2f)
                                    )
                                )
                                NavigationBarItem(
                                    selected = currentRoute == "chat",
                                    onClick = {
                                        navController.navigate("chat") {
                                            popUpTo(navController.graph.findStartDestination().id) { saveState = true }
                                            launchSingleTop = true
                                            restoreState = true
                                        }
                                    },
                                    icon = { Icon(Icons.Default.ChatBubble, contentDescription = "Chat") },
                                    label = { Text("AI Assistant") },
                                    colors = NavigationBarItemDefaults.colors(
                                        selectedIconColor = HubCyan,
                                        selectedTextColor = HubCyan,
                                        indicatorColor = HubIndigo.copy(alpha = 0.2f)
                                    )
                                )
                                NavigationBarItem(
                                    selected = currentRoute == "apps",
                                    onClick = {
                                        navController.navigate("apps") {
                                            popUpTo(navController.graph.findStartDestination().id) { saveState = true }
                                            launchSingleTop = true
                                            restoreState = true
                                        }
                                    },
                                    icon = { Icon(Icons.Default.Apps, contentDescription = "Apps") },
                                    label = { Text("Satellites") },
                                    colors = NavigationBarItemDefaults.colors(
                                        selectedIconColor = HubCyan,
                                        selectedTextColor = HubCyan,
                                        indicatorColor = HubIndigo.copy(alpha = 0.2f)
                                    )
                                )
                                NavigationBarItem(
                                    selected = currentRoute == "profile",
                                    onClick = {
                                        navController.navigate("profile") {
                                            popUpTo(navController.graph.findStartDestination().id) { saveState = true }
                                            launchSingleTop = true
                                            restoreState = true
                                        }
                                    },
                                    icon = { Icon(Icons.Default.Person, contentDescription = "Profile") },
                                    label = { Text("Profile") },
                                    colors = NavigationBarItemDefaults.colors(
                                        selectedIconColor = HubCyan,
                                        selectedTextColor = HubCyan,
                                        indicatorColor = HubIndigo.copy(alpha = 0.2f)
                                    )
                                )
                            }
                        }
                    }
                ) { innerPadding ->
                    NavHost(
                        navController = navController,
                        startDestination = "splash",
                        modifier = Modifier.padding(if (showBottomBar) innerPadding else PaddingValues(0.dp))
                    ) {
                        composable("splash") {
                            SplashScreen(
                                tokenManager = app.tokenManager,
                                onSplashFinished = { isLoggedIn ->
                                    val targetRoute = if (isLoggedIn) "dashboard" else "auth"
                                    navController.navigate(targetRoute) {
                                        popUpTo("splash") { inclusive = true }
                                    }
                                }
                            )
                        }

                        composable("auth") {
                            AuthScreen(
                                tokenManager = app.tokenManager,
                                onLoginSuccess = {
                                    navController.navigate("dashboard") {
                                        popUpTo("auth") { inclusive = true }
                                    }
                                }
                            )
                        }

                        composable("dashboard") {
                            DashboardScreen(
                                viewModel = dashboardViewModel,
                                onNavigateToChat = { navController.navigate("chat") }
                            )
                        }

                        composable("chat") {
                            AiChatScreen(viewModel = aiChatViewModel)
                        }

                        composable("apps") {
                            ConnectedAppsScreen()
                        }

                        composable("profile") {
                            ProfileScreen(
                                onSignOut = {
                                    coroutineScope.launch {
                                        app.tokenManager.clearSession()
                                        navController.navigate("auth") {
                                            popUpTo(navController.graph.findStartDestination().id) { inclusive = true }
                                        }
                                    }
                                }
                            )
                        }
                    }

                    // Account Mismatch Dialog
                    mismatchState?.let { state ->
                        AccountMismatchBottomSheet(
                            state = state,
                            onSwitchAccount = {
                                mismatchState = null
                                Toast.makeText(this@MainActivity, "Mengarahkan ke login akun...", Toast.LENGTH_SHORT).show()
                            },
                            onUpdateHubConnection = {
                                mismatchState = null
                                Toast.makeText(this@MainActivity, "Koneksi Hub diperbarui!", Toast.LENGTH_SHORT).show()
                            },
                            onCancel = {
                                mismatchState = null
                            }
                        )
                    }
                }
            }
        }
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        handleIncomingIntent(intent)
    }

    private fun handleIncomingIntent(intent: Intent?) {
        val uri = intent?.data ?: return
        if (uri.scheme == "ekahub" && uri.host == "link-callback") {
            val status = uri.getQueryParameter("status") ?: "success"
            val appId = uri.getQueryParameter("app_id") ?: "app"
            Toast.makeText(this, "Berhasil menghubungkan $appId ke Personal AI Hub ($status)!", Toast.LENGTH_LONG).show()
        }
    }
}
