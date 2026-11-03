# Review & Koreksi Skenario

Halaman `/review` menggunakan skenario persisten dari PostgreSQL. Dari Generate,
tombol Review membuka `/review?scenario=<scenario_id>`; ID berasal dari respons
generate setelah record berhasil disimpan. Dropdown hanya memuat skenario milik
System Analyst yang sedang login. UUID ditampilkan apa adanya, bukan ID contoh.

## Alur

1. Backend memeriksa pengguna dan role melalui repository pengguna existing.
2. Detail skenario dan GUI elements dimuat dari generation request yang terkait.
3. Python memvalidasi teks asli dengan parser resmi Cucumber. Tidak ada panggilan
   LLM untuk review, perbaikan otomatis, atau pemotongan teks pengguna.
4. Editor read-only sampai Edit ditekan. Draft disimpan hanya di state halaman.
5. Simpan mengirim `gherkin` dan `updated_at` versi yang dibuka. Backend memvalidasi
   ulang, lalu memperbarui teks, judul, `source_type=CORRECTED`, `status=VALID`,
   coverage, dan `updated_at` dengan satu conditional UPDATE. Revisi dari tab lain
   menghasilkan 409 agar koreksi tidak tertimpa tanpa diketahui.
6. Validasi gagal menghasilkan 422 dengan daftar error bernomor baris. Tidak ada
   UPDATE; frontend mempertahankan draft. Coverage rendah adalah warning.
7. Download memuat dan memvalidasi ulang versi tersimpan. Respons berupa bytes
   teks yang sama, termasuk indentasi, akhir baris, dan trailing newline. Draft
   tidak pernah dikirim ke endpoint download.

Perpindahan skenario, link internal, back/forward, refresh/tutup tab, dan logout
memiliki perlindungan perubahan belum disimpan. Konfirmasi refresh/tutup tab
menggunakan dialog bawaan browser. Batalkan edit juga meminta konfirmasi jika
draft berubah. Navbar dan sidebar menggunakan komponen yang sama dengan Generate.

## API

Semua endpoint publik mengikuti envelope JSON existing, kecuali download.
Semua memerlukan Bearer token, role `system_analyst`, dan kepemilikan skenario.
Record yang tidak ada atau milik pengguna lain menghasilkan 404.

| Method | Path | Kegunaan |
| --- | --- | --- |
| GET | `/api/v1/scenarios` | Daftar ringkasan skenario pengguna |
| GET | `/api/v1/scenarios/:id` | Teks tersimpan, GUI elements, dan validasi aktual |
| PUT | `/api/v1/scenarios/:id` | Validasi dan simpan koreksi |
| GET | `/api/v1/scenarios/:id/download` | Download `.feature` versi valid tersimpan |
| POST | `/internal/scenarios/validate` | Validator Python; hanya dengan `X-AI-Service-Key` |

Body koreksi:

```json
{
  "gherkin": "Feature: Login\n  Scenario: Success\n    Given a page\n    When I click Login button\n    Then I see a dashboard\n",
  "updated_at": "<nilai updated_at persis dari respons detail>"
}
```

Respons detail/koreksi memuat `id`, `feature_name`, `source_type`, `status`,
`updated_at`, `gherkin`, `gui_elements`, dan `validation`. Validasi memuat
`valid`, `checks`, `issues` (message/line/column), `used_elements`,
`not_covered_elements`, dan `coverage_percent`. Pada 422, hasil validasi berada
di `data.validation`. Pada konflik 409, muat ulang versi server sebelum mencoba
menyimpan kembali; salin draft bila ingin mempertahankannya untuk digabung manual.

## Aturan SRS

Mengikuti Section 4.2 REQ-01–REQ-11 serta Section 6.2–6.3. Hanya keyword Feature,
Scenario, Given, When, Then, And yang diterima pada review. Nama wajib terisi;
nama Scenario unik (case-insensitive); setiap Scenario memerlukan Given/When/Then
dengan urutan yang benar. Siklus When–Then berulang mengikuti aturan tertulis
Section 6.2; tidak menambahkan status APPROVED atau kebijakan TBD lainnya.

Coverage hanya memeriksa teks langkah pada AST, bukan judul, komentar, tag,
deskripsi, doc string, atau tabel. Normalisasi menggunakan lowercase, membuang
tanda baca/kutip, dan merapikan whitespace; pencocokan substring, pembulatan dua
desimal. Tidak ada nilai coverage atau daftar elemen dari mockup di production.

Tidak diperlukan migrasi baru: memakai GherkinScenario, GenerationRequest, dan
User existing. API tidak memerlukan konfigurasi baru; tetap memakai
`AI_SERVICE_URL`, `AI_SERVICE_KEY`, dan `VITE_API_URL` yang sudah ada.

## Menjalankan dan menguji

Jalankan PostgreSQL, backend Go, Python service, dan frontend seperti biasa.
Python dan backend harus memakai service key yang sama. Login sebagai System
Analyst, generate skenario, lalu tekan Review Skenario. Ollama diperlukan untuk
generate, tetapi tidak diperlukan untuk memvalidasi atau menyimpan koreksi.

Frontend:

```sh
npm ci
npm run lint
npm run build
```

Backend (dari `backend/`):

```sh
go test -count=1 -timeout=5m ./...
go vet ./...
go build ./...
```

Gunakan `go test -race` di lingkungan dengan CGO dan compiler C yang sesuai.

Python (dari `ai-service/`, dependency `requirements-dev.txt` terpasang):

```sh
python -m pytest tests/unit/test_review_validation.py tests/api/test_review_api.py --noconftest -q -o addopts=
```

`--noconftest` diperlukan sementara karena conftest test lama mengimpor
`app.generator` yang sudah tidak ada setelah restrukturisasi service. Test review
berdiri sendiri dan tidak memakai fixture lama. Masalah suite lama terpisah dari
fitur ini; jangan menganggap command ini menguji seluruh AI service.

Job CI `review-validation` menjalankan test review yang sama pada Python 3.12,
tanpa Ollama atau kredensial provider. Job backend dan frontend tetap berjalan.

Test integrasi opsional `TestReviewPersistenceWithPostgresAndPython` memakai
`TEST_REVIEW_DATABASE_URL` dan `TEST_REVIEW_VALIDATOR_URL`. Jalankan hanya pada
database uji terpisah dengan izin membuat schema dan extension uuid-ossp. Test
membuat schema unik dan membersihkannya setelah selesai; tidak mengubah tabel
aplikasi. Service key validator uji adalah
`review-verification-service-key-1234567890` (khusus fixture, bukan production).

## Checklist browser

- Memuat daftar/detail, memilih hasil generate yang benar, dan menampilkan empty/error states.
- Save nonaktif sebelum edit atau bila teks belum berubah.
- Draft invalid tetap di editor; error mencantumkan baris/penyebab dan database tetap sama.
- Koreksi valid dengan coverage di bawah 100% tersimpan dan tetap ada setelah reload.
- Download saat ada draft mengunduh versi tersimpan; versi tersimpan invalid tidak dapat diekspor.
- Membatalkan konfirmasi perpindahan/keluar tidak menghapus draft atau sesi login.
- Konfirmasi meninggalkan draft, batal edit, akses QA/unauthenticated, dan retry API.
- Tampilan desktop dan mobile tanpa overflow horizontal.

## Daftar file perubahan

- `.github/workflows/ci.yml` — diubah.
- `ai-service/app/api/routes/review.py` — baru.
- `ai-service/app/main.py` — diubah.
- `ai-service/app/schemas/review.py` — baru.
- `ai-service/app/services/validation.py` — baru.
- `ai-service/tests/api/test_review_api.py` — baru.
- `ai-service/tests/unit/test_review_validation.py` — baru.
- `backend/modules/scenario/dto/review_dto.go` — baru.
- `backend/modules/scenario/dto/scenario_dto.go` — diubah.
- `backend/modules/scenario/handler/review_handler.go` — baru.
- `backend/modules/scenario/repository/review_repository.go` — baru.
- `backend/modules/scenario/routes.go` — diubah.
- `backend/modules/scenario/service/review_service.go` — baru.
- `backend/modules/scenario/service/scenario_service.go` — diubah.
- `backend/modules/scenario/service/validation_client.go` — baru.
- `backend/providers/core.go` — diubah.
- `backend/test/scenario/review/persistence_test.go` — baru.
- `backend/test/scenario/review/review_test.go` — baru.
- `docs/REVIEW_SCENARIO.md` — baru.
- `frontend/src/App.tsx` — diubah.
- `frontend/src/components/PublicHeader.tsx` — diubah.
- `frontend/src/components/review/GherkinEditor.tsx` — baru.
- `frontend/src/components/review/ScenarioSelector.tsx` — baru.
- `frontend/src/components/review/ValidationPanel.tsx` — baru.
- `frontend/src/components/ScenarioSidebar.tsx` — baru.
- `frontend/src/pages/GenerateScenarioPage.tsx` — diubah.
- `frontend/src/pages/ReviewScenarioPage.css` — baru.
- `frontend/src/pages/ReviewScenarioPage.tsx` — baru.
- `frontend/src/services/reviewApi.ts` — baru.
- `frontend/src/services/scenarioApi.ts` — diubah.
