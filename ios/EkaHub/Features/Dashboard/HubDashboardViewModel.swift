import Foundation
import SwiftUI

@MainActor
@Observable
public final class HubDashboardViewModel {
    public var dailyBriefingText: String = ""
    public var summaryCards: [AppSummaryResponse] = []
    public var suggestedActions: [SuggestedAction] = []
    public var isLoading: Bool = false
    public var errorMessage: String?

    private let apiClient = APIClient.shared

    public init() {}

    public func loadDashboard() async {
        isLoading = true
        errorMessage = nil

        do {
            let briefing = try await apiClient.fetchDailyBriefing()
            dailyBriefingText = briefing.dailyAiBriefing
            summaryCards = briefing.miniSummaryCards ?? []
            suggestedActions = briefing.suggestedActions ?? []
        } catch {
            errorMessage = error.localizedDescription
            summaryCards = []
        }

        isLoading = false
    }
}
