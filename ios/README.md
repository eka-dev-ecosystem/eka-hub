# 🍏 Personal AI Hub - iOS Native App

Aplikasi iOS murni (*Pure Native*) berbasis **Swift 5.10 / Swift 6**, **SwiftUI (iOS 17+)**, Observation framework (`@Observable`), AppLogger, type-safe L10n localization, dan arsitektur modular Core & Features.

## 🚀 Fitur Utama
- **Modern SwiftUI 17+**: Menggunakan `@Observable` macro dan Swift Concurrency (`async/await`).
- **100% Native Design System**: Dark theme modern, Bento Cards, Glassmorphism card styles, dan gradient banner.
- **DeepLinkGuard & DeepLinkManager**: Navigasi mulus ke satelit (F-Tracker & Task Management) dengan verifikasi `expected_uid`.
- **Daily AI Briefing & Chat**: Ringkasan cerdas Gemini dan eksekusi aksi satu ketukan.
- **Structured AppLogger**: Pencatatan log komprehensif via `os.Logger` dan styled console badges.

## 📂 Struktur Modular
```text
ios/EkaHub/
├── EkaHubApp.swift           # Application entrypoint & AppDelegate
├── MainTabView.swift         # Root tab navigation & deep link coordinator
├── Info.plist                # Schemas & queries
├── EkaHub.entitlements       # Keychain & App Groups
├── Core/
│   ├── Config/               # AppConfig.swift (EnvReader, AppEnvironment)
│   ├── DesignSystem/         # Theme.swift, Components.swift
│   ├── Localization/         # I18n.swift, L10n.swift
│   ├── Logger/               # AppLogger.swift
│   ├── Models/               # Models.swift
│   ├── Network/              # APIClient.swift, NetworkLogger.swift, SSEClient.swift
│   ├── Notification/         # PushNotificationManager.swift
│   ├── Security/             # KeychainManager.swift
│   ├── Shared/               # DateTimeHelper.swift
│   ├── DeepLink/             # DeepLinkManager.swift
│   └── Guard/                # ChildAppDeepLinkGuard.swift
├── Features/
│   ├── Auth/                 # AuthView.swift, AuthViewModel.swift
│   ├── Dashboard/            # HubDashboardView.swift, HubDashboardViewModel.swift
│   ├── Chat/                 # AiChatView.swift, AiChatViewModel.swift
│   ├── ConnectedApps/        # ConnectedAppsView.swift, ConnectedAppsViewModel.swift
│   └── Profile/              # ProfileView.swift, ProfileViewModel.swift
└── Resources/
    ├── Assets.xcassets/
    └── Localizable.xcstrings
```
