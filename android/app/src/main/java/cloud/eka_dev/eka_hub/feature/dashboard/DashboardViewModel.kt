package cloud.eka_dev.eka_hub.feature.dashboard

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import cloud.eka_dev.eka_hub.core.model.AppSummaryResponse
import cloud.eka_dev.eka_hub.core.network.ApiClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class DashboardViewModel(private val apiClient: ApiClient) : ViewModel() {

    private val _dailyBriefing = MutableStateFlow(
        """Halo Eka! Berdasarkan sinkronisasi aplikasi Anda hari ini:

💰 F-Tracker: Total pengeluaran tercatat Rp 4.250.000 (sisa budget 57% aman).
📋 Task Management: 5 tugas menunggu, dengan prioritas utama "Deploy Eka-Hub to Staging".

Semua sistem berjalan normal. Klik kartu bento di bawah untuk membuka aplikasi asli."""
    )
    val dailyBriefing: StateFlow<String> = _dailyBriefing.asStateFlow()

    private val _summaries = MutableStateFlow<List<AppSummaryResponse>>(
        listOf(
            AppSummaryResponse(
                appId = "ftracker",
                appName = "F-Tracker",
                connectedUid = "usr_ftracker_123",
                connectedEmail = "eka@finance.app",
                highlightMetric = "Rp 4.250.000",
                highlightLabel = "Pengeluaran Bulan Ini",
                deepLink = "ftracker://expenses?category=food&expected_uid=usr_ftracker_123"
            ),
            AppSummaryResponse(
                appId = "taskmanagement",
                appName = "Task Management",
                connectedUid = "usr_task_456",
                connectedEmail = "eka@taskhub.app",
                highlightMetric = "5 Tugas Pending",
                highlightLabel = "Deadline Hari Ini",
                deepLink = "taskhub://tasks?filter=today&expected_uid=usr_task_456"
            )
        )
    )
    val summaries: StateFlow<List<AppSummaryResponse>> = _summaries.asStateFlow()

    fun refreshData() {
        viewModelScope.launch {
            try {
                val res = apiClient.apiService.getAppSummaries()
                if (res.success && res.data != null) {
                    _summaries.value = res.data
                }
            } catch (_: Exception) {
                // Keep cached data
            }
        }
    }
}
