import SwiftUI

@MainActor
@Observable
public final class HubDashboardViewModel {
    public var briefingText: String = """
    Halo Eka! Berdasarkan sinkronisasi aplikasi Anda hari ini:

    💰 F-Tracker: Total pengeluaran tercatat Rp 4.250.000 (sisa budget 57% aman).
    📋 Task Management: 5 tugas menunggu, dengan prioritas utama "Deploy Eka-Hub to Staging".

    Semua sistem berjalan normal. Klik kartu bento di bawah untuk membuka aplikasi asli.
    """

    public var summaries: [AppSummaryResponse] = [
        AppSummaryResponse(
            appId: "ftracker",
            appName: "F-Tracker",
            connectedUid: "usr_ftracker_123",
            connectedEmail: "eka@finance.app",
            highlightMetric: "Rp 4.250.000",
            highlightLabel: "Pengeluaran Bulan Ini",
            deepLink: "ftracker://expenses?category=food&expected_uid=usr_ftracker_123"
        ),
        AppSummaryResponse(
            appId: "taskmanagement",
            appName: "Task Management",
            connectedUid: "usr_task_456",
            connectedEmail: "eka@taskhub.app",
            highlightMetric: "5 Tugas Pending",
            highlightLabel: "Deadline Hari Ini",
            deepLink: "taskhub://tasks?filter=today&expected_uid=usr_task_456"
        )
    ]

    public init() {}

    public func refresh() async {
        do {
            let fetched = try await APIClient.shared.fetchSummaries()
            if !fetched.isEmpty {
                self.summaries = fetched
            }
        } catch {
            AppLogger.warn("Failed to fetch fresh summaries from backend; using cached data", category: .hub)
        }
    }
}
