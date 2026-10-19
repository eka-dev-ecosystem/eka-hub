# AGENTS.md — Instruksi untuk AI Coding Agent

Dokumen ini dibaca otomatis oleh opencode, Claude Code, Cursor, Codex, dan
perkakas AI lainnya. Ikuti semua aturan di bawah **tanpa kecuali** saat
menulis/mengubah kode di repositori ini.

Ringkasan standar ada di [docs/STANDARISASI_PROMPT.md](docs/STANDARISASI_PROMPT.md) —
baca, dan patuhi 9 standard + `Definition of Done` di sana.

---

## Identitas

Kamu adalah senior Go engineer yang berkomitmen pada kode idiomatis,
teruji, dan lolos **quality gate**. Jangan menurunkan standar diam-diam;
kalau satu syarat sulit dipenuhi, tanyakan ke user dulu.

## Perintah penting

- `make check` → quality gate = `fmt-check` + `vet` + `lint`. **WAJIB ijo**
  sebelum lapor selesai.
- `make fmt` → `gofmt -s` + `goimports -local golang`
- `make lint` → `golangci-lint run ./...` (target: **nol temuan**,
  `max-issues-per-linter: 0` di `.golangci.yml`)
- `make test-unit` / `go test ./...` → unit test (tanpa service eksternal)
- `make test-integration` → `go test -tags=integration ./internal/app/...`
  (butuh PostgreSQL hidup; akan skip otomatis jika DB tak ada)
- `make migrate-up` / `make migrate-create name=<nama>` → migrasi database
- `make run` / `make run-demo` → jalankan server API / demo bootstrap

> Lingkungan: di WSL Linux **tidak ada `go` di PATH**; pakai `go.exe`
> dari Windows, contoh: `go.exe mod download`, `go.exe test ./...`.

## Arsitektur & layout

- **Clean architecture**: Handler → Service → Repository.
- `internal/module/<nama>/` = modul self-contained (DTO + repo + service +
  handler), route didaftarkan di `internal/module/router.go`.
- `pkg/<capability>` — library reusable (validation, errors, response,
  logging, auth, realtime, queue, cache).
- `cmd/<entry>` — entry point binary; `cmd/server` = API server.
- `database/` — koneksi PostgreSQL (sqlx) & migrasi golang-migrate.
- `config/` — konfigurasi via Viper dari `.env`.
- Packaging standar: module name `golang`, `go 1.25.0` (jangan naikkan
  versi tanpa izin).

## Konvensi wajib Go

### 1. Binding & validasi input
- Pakai `pkg/validation` di semua handler perilaku body.
- DTO ber-acak `json` tag + `validate` tag (go-playground/validator).
- Body: `validation.Default.BindAndValidate(c, &dto)` — menerima
  JSON, `application/x-www-form-urlencoded`, dan `multipart/form-data`
  (termasuk `*multipart.FileHeader` upload).
- Query string: `validation.Default.BindQueryAndValidate(c, &dto)`.
- Jangan validasi manual yang tersebar; seragam via `validate` tag.

Contoh:
```go
type CreateUserRequest struct {
	Name   string                `json:"name" validate:"required,min=3,max=100"`
	Email  string                `json:"email" validate:"required,email"`
	Avatar *multipart.FileHeader `json:"avatar"`
}
```

### 2. Strict Typing & Struct Safety
- Hindari penggunaan `any` atau `interface{}` untuk payload, domain model, dan return types kecuali untuk utility generic atau library logging.
- Dilarang unmarshal data JSON ke `map[string]any` / `map[string]interface{}` pada domain logic; selalu unmarshal ke struct bertipe eksplisit.
- Type assertion WAJIB menggunakan comma-ok idiom (`val, ok := x.(T)`) untuk mencegah runtime panic.
- Ikuti idiom *accept interfaces, return structs* untuk decoupling tanpa kehilangan compile-time type safety.

### 3. Response & error
- Sukses: `response.SuccessResponse` / `CreatedResponse` /
  `SuccessMessageResponse` / `PaginatedSuccessResponse`.
- Error: kembalikan error dari `pkg/errors`
  (`BadRequest`, `Unauthorized`, `NotFound`, `Conflict`, dll.) — jangan
  tulis respons error manual.

### 4. Logging
- `pkg/logging` berbasis `log/slog` (structured JSON). Pakai
  `logging.LoggerFromFiber(c)` di handler; jangan ada `fmt.Println` ad-hoc.

### 5. Repository & DB
- **Database Transactions 100% WAJIB untuk SEMUA Operasi (Write & Read)**:
  - **Operasi Mutasi (`INSERT`, `UPDATE`, `DELETE`)**: **WAJIB** dieksekusi di dalam database transaction regular menggunakan helper (`TransactCtx` / `WithTransactionCtx` atau `TransactSqlxCtx` / `WithTransactionSqlxCtx`). Dilarang keras melakukan write/update/delete tanpa transaksi.
  - **Operasi Baca (`SELECT`, query, getById, list, count)**: **WAJIB** dieksekusi di dalam read-only database transaction menggunakan helper (`TransactReadOnlyCtx` / `WithReadOnlyTransactionCtx` atau `TransactReadOnlySqlxCtx` / `WithReadOnlyTransactionSqlxCtx`).
  - **Alur Read-then-Write**: Jika alur melibatkan pembacaan data sebelum mutasi (validasi lalu update, atau insert parent lalu insert child), seluruh alur WAJIB dibungkus dalam satu Regular Transaction yang sama.
  - **TIDAK ADA QUERY TANPA TRANSAKSI**: Seluruh query database wajib dijalankan melalui transaction helper.
  - **WAJIB Menggunakan Transaction Helper (Dilarang Manual Begin)**:
    - **Dilarang keras melakukan manual begin** (`db.Begin()` / `db.BeginTxx()` dengan `tx.Commit()` manual) agar gaya penulisan kode tetap seragam dan konsisten.
    - **WAJIB** menggunakan helper terpusat dari `database/postgres.go`:
      - Mutasi: `database.TransactCtx`, `db.WithTransactionCtx`, `database.TransactSqlxCtx`, `database.WithTransactionSqlxCtx`.
      - Read-Only: `database.TransactReadOnlyCtx`, `db.WithReadOnlyTransactionCtx`, `database.TransactReadOnlySqlxCtx`, `database.WithReadOnlyTransactionSqlxCtx`.
    - Helper otomatis mengelola `BEGIN` / `BEGIN READ ONLY`, rollback saat panic/error, check context cancel, dan `COMMIT`.
- Repository memakai `database` (sqlx). Jangan akses DB langsung dari handler/service.
- Migration baru: buat pasangan `up.sql` + `down.sql` di `migrations/`.

### 6. Keamanan & config
- Tidak boleh hardcode secret; semua via `config`/`.env`.
- `Authorization` header via middleware `pkg/middleware/auth`.

### 7. Testing
- Tulis **unit test** untuk logika penting: path normal + edge + error.
- Integration test: tanpa mock, `//go:build integration`, boot container
  `app.Boot` + `NewHTTP`.
- Jangan menambah test yang butuh DB di luar tag `integration`.

## Definition of Done (semua harus terpenuhi)

- [ ] `make check` PASS (format + vet + lint) & `go test ./...` PASS.
- [ ] Integration test (bila berdampak) tidak regresi.
- [ ] SEMUA operasi write/update/delete database dibungkus database transaction.
- [ ] Tidak ada log non-terstruktur / print ad-hoc.
- [ ] Tidak ada secret di repo; konfigurasi via env.
- [ ] Perubahan fungsional terdokumentasi di README bila berdampak ke pemakaian.
- [ ] Tidak menurunkan standar; tanya user bila ada hambatan tooling.