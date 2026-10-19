# Agent Operational Guidelines & Proactive Task Management

## 1. Project Identity & Configuration
- **Project Name**: Eka-Hub (Personal AI Hub / Life Aggregator)
- **Target Project ID**: `11`
- **App Package / Bundle Identifier**:
  - **iOS (`CFBundleIdentifier`)**: `cloud.eka-dev.eka-hub`
  - **Android (`applicationId` & `namespace`)**: `cloud.eka_dev.eka_hub`
- **Ecosystem**:
  - `backend`: Golang (Fiber v2) REST API, PostgreSQL, Redis, Gemini Function Calling
  - `android`: Native Android (Kotlin 2.1.0, Jetpack Compose, Material 3)
  - `ios`: Native iOS (Swift 6, SwiftUI)
  - `integration`: DeepLinkGuards & standard `/summary` contracts for satellite apps (F-Tracker, Task Management)

---

## 2. Autonomous Task Lifecycle (`task-management` MCP)

Setiap kali pengguna meminta pengerjaan fitur, perbaikan bug, refactoring, atau enhancement pada repository **Eka-Hub** ini, jalankan siklus task management berikut **secara mandiri dan proaktif tanpa perlu diingatkan**:

### Phase 1: Search & Discovery
- Sebelum memulai implementasi apapun, periksa selalu task aktif menggunakan tool MCP `task-management` -> `list_tasks` dengan parameter:
  - `project_id`: `11`
  - `status`: `"in_progress"` (atau kata kunci terkait).
- Cegah pembuatan task duplikat jika task untuk pekerjaan tersebut sudah pernah dibuat sebelumnya.

### Phase 2: Adopt or Auto-Create
- **Adopt Existing**: Jika task yang cocok sudah ada di pipeline, gunakan task tersebut dan pastikan statusnya `in_progress` via `update_task_status`.
- **Auto-Create New**: Jika belum ada task yang sesuai:
  - Otomatis buat task baru menggunakan tool MCP `create_task`:
    - `project_id`: `11`
    - `title`: Judul deskriptif dan ringkas dengan prefix emoji (contoh: `✨ Feat: ...`, `🐛 Fix: ...`, `♻️ Refactor: ...`).
    - `description`: Rincian teknis pengerjaan dan ruang lingkup perubahan.
    - `status`: `"in_progress"`
    - `priority`: Sesuai urgensi (`urgent`, `high`, `medium`, `low`).
    - `subtasks`: Rincian checklist langkah implementasi & verifikasi.
    - `is_focus`: `true`
    - `tags`: Tag domain yang relevan (contoh: `["android", "ios", "backend", "hub"]`).

### Phase 3: Auto-Complete Upon Verification & Push
- Setelah pekerjaan selesai diimplementasikan, seluruh quality gates lolos, dan kode sudah di-commit/push ke remote repository:
  - Otomatis panggil `log_task_progress` untuk task terkait:
    - `task_id`: ID task terkait.
    - `completed_subtask_ids`: Array ID subtask yang telah diselesaikan.
    - `progress_note`: Rangkuman ringkas fitur/bugfix yang diselesaikan serta hasil verifikasi quality gate.
    - `status`: `"done"`
- Siklus ini harus berjalan secara otomatis dan mandiri (self-managed) tanpa menunggu instruksi eksplisit dari pengguna.

---

## 3. Quality Gates
Selalu lakukan verifikasi mandiri sebelum melakukan commit dan menutup task:
- **Backend (`backend/`)**: Jalankan `go test ./...` dan pastikan seluruh test 100% pass.
- **Android (`android/`)**: Pastikan manifest, resource XML, dan Compose kode bebas dari compilation/lint error (`./gradlew assembleDebug` atau build check).
- **iOS (`ios/`)**: Pastikan Swift syntax, bundle ID `cloud.eka-dev.eka-hub`, dan SwiftUI views valid.
