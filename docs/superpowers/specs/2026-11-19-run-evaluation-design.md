# Run Evaluation — Design

Status: approved (chat, 2026-11-19). QA-only.

## Scope

QA memilih Golden Dataset, menjalankan evaluasi head-to-head LLM vs Rule-Based per
kasus, dan melihat progres sampai selesai. Hasil disimpan ke `evaluation_runs` dan
`evaluation_run_cases` yang sudah ada lalu dibaca oleh Evaluation Report.

Tidak dibangun: skor kemiripan terhadap golden, `winner`, Overall Score, antrean
worker terpisah, status `PENDING` (tetap `QUEUED`). Teks Gherkin hasil generate tidak
disimpan.

Revisi implementasi: kolom progres baru tetap dipakai, yaitu
`evaluation_run_cases.processed` (boolean), karena kasus yang gagal dihasilkan oleh
kedua metode menyimpan hasil `null` sehingga jumlah kasus terisi tidak dapat
menghitung progres secara andal.

## Alur

1. `POST /api/v1/evaluations` body `{golden_dataset_id}` (QA) → tolak 409 bila ada run
   `RUNNING`.
2. Snapshot `evaluation_run_cases` (`case_id`, `position`, `title`, `expected_gherkin`)
   dari item dataset, set `used_at` dataset, status `QUEUED`.
3. Goroutine: status `RUNNING`. Per kasus, kedua generator menerima input identik
   (`user_story`, `gui_elements`, `free_scenario`):
   - LLM: ai-service `POST /internal/scenarios/generate`, lalu `.../validate`.
   - Rule-Based: ai-service `POST /internal/scenarios/generate-rule-based` (deterministik),
     lalu validator yang sama.
   - Simpan hasil kasus segera setelah selesai; kegagalan satu metode = hasil `null`,
     kegagalan satu kasus tidak menghentikan run.
4. Selesai: hitung metrik, `COMPLETED`. Error fatal (dataset hilang, DB gagal) → `FAILED`
   dengan `error_message` publik.
5. `GET /api/v1/evaluation-runs/:run_id/progress` (QA) → status +
   `processed_cases`/`total_cases` (dihitung dari `processed = true`), untuk polling UI.

## Aturan status dan metrik

- `PASS`: Gherkin valid dan GUI coverage 100.
- `WARN`: Gherkin valid, coverage < 100.
- `FAIL`: Gherkin tidak valid.
- Metode tanpa hasil (generate/validate gagal) → hasil `null`, dihitung tidak valid pada
  agregat.
- `validation_percent` = valid/total×100; `gui_coverage_percent` = rata-rata coverage
  kasus yang punya hasil. Keduanya dibulatkan 2 desimal. Aturan ini ditambahkan ke
  `docs/EVALUATION_REPORT_API.md`.

## Penempatan generator

Rule-Based tinggal di ai-service (Python) sebagai implementasi kontrak generate yang
sama, karena generator adalah subjek uji dan tidak boleh berada di dalam komponen yang
menghitung metrik (pola evaluation harness). Backend Go tetap orkestrator.

## File

- ai-service: `app/services/rule_based.py`, route di `app/api/routes/generate.py`,
  schema response memakai `GenerateResponse`.
- backend: `modules/scenario/service/evaluation_run_service.go`,
  `modules/scenario/repository/evaluation_run_repository.go`,
  `modules/scenario/handler/evaluation_run_handler.go`,
  `modules/scenario/dto/evaluation_run_dto.go`, `evaluation_run_routes.go`,
  wiring di `providers/core.go`.
- frontend: `pages/EvaluationPage.tsx` (route `/evaluation` sudah QA-guarded),
  `services/evaluationRunApi.ts`.

## Verifikasi

- ai-service: pytest untuk rule-based (deterministik, output lolos `validate_gherkin`).
- backend: unit test service dengan fake generator/validator (metrik, status, kegagalan
  per kasus, satu run aktif), handler test, tes repo.
- frontend: `npm run build`; uji manual QA (jalankan, progres, tautan report).
