# Review & Koreksi Skenario

Halaman `/review` menggunakan skenario persisten dari PostgreSQL. Dari Generate,
tombol Review membuka `/review?scenario=<scenario_id>`; ID berasal dari respons
generate setelah record berhasil disimpan. Dropdown hanya memuat skenario milik
System Analyst yang sedang login. Dropdown menampilkan kode tetap seperti
`SCN-0001 — User Login`; UUID tetap digunakan untuk pemilihan dan endpoint API.

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

Respons daftar/detail/koreksi memuat `id`, `scenario_code`, `feature_name`, `source_type`, `status`,
`updated_at`, `gherkin`, `gui_elements`, dan `validation`. Validasi memuat
`valid`, `checks`, `issues` (message/line/column), `used_elements`,
`not_covered_elements`, `coverage_percent`, dan `coverage_available`. Pada 422, hasil validasi berada
di `data.validation`. Pada konflik 409, muat ulang versi server sebelum mencoba
menyimpan kembali; salin draft bila ingin mempertahankannya untuk digabung manual.

## Kode skenario

PostgreSQL memberikan `scenario_number` melalui sequence saat record dibuat.
Backend mengirimkannya sebagai `scenario_code` dengan format SCN dan minimal
empat digit: `SCN-0001`, `SCN-0042`, `SCN-10000`. Nomor bersifat global dan unik,
bukan posisi dropdown atau nomor per pengguna. Urutan dropdown tetap berdasarkan
waktu pembuatan terbaru. Nomor dapat memiliki jeda akibat penghapusan, transaksi
batal, record invalid, atau skenario pengguna lain.

Edit judul, simpan koreksi, pengurutan daftar, dan restart tidak mengganti nomor.
Nomor yang dihapus tidak dipakai kembali oleh sequence. API tetap menerima UUID
untuk lookup dan tetap memeriksa role serta kepemilikan. Frontend memakai UUID
sebagai value pilihan, tetapi menampilkan kode dan judul. Jika frontend diperbarui
sebelum backend, dropdown sementara menampilkan judul saja.

Restart backend untuk menjalankan AutoMigrate sebelum mencoba kode skenario.
Tidak perlu konfigurasi baru atau migrasi SQL manual. Tes PostgreSQL baru memakai
`TEST_DATABASE_URL` dan schema disposable; CI backend sudah menyediakan variabel
tersebut. Tes mencakup backfill, migrasi berulang, create bersamaan, kepemilikan,
kode setelah rename/hapus, dan format melebihi empat digit.

## Diagnostik validasi

Feature/Scenario yang sudah terisi tetap ditandai hijau ketika langkah lain salah.
Given/When/Then/And tanpa deskripsi mendapatkan pesan spesifik beserta nomor baris,
misalnya "Given belum memiliki deskripsi langkah". Pesan token internal parser tidak
ditampilkan. Pemeriksaan sumber saat parser gagal hanya membantu menunjukkan
bagian yang terdeteksi; validitas tetap memerlukan parser Cucumber atas teks asli
dan seluruh aturan semantik. Teks tidak diperbaiki otomatis.

Jika parser gagal atau terdapat langkah tanpa deskripsi, `coverage_available=false`:
persentase numerik bernilai 0 sebagai placeholder, kedua daftar elemen kosong, dan
UI menampilkan "Belum dapat dihitung" tanpa persentase maupun daftar Not Covered.
Ini berbeda dari coverage 0% yang benar-benar dihitung dari langkah AST. Struktur
yang terbaca tetapi urutan langkahnya salah tetap dapat memiliki coverage aktual.
Klien lama yang tidak mengirim field ini tetap didukung (field absen berarti tersedia).

Pesan gagal simpan muncul sekali di panel Validation Result. Draft tetap tersedia
dan download memakai versi valid terakhir yang tersimpan. Setelah perubahan
validator/kontrak ini, restart AI service dan backend lalu refresh frontend.

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

Migrasi otomatis menambahkan `scenario_number` (PostgreSQL BIGSERIAL, unik dan
tidak boleh NULL) pada GherkinScenario. Data lama mendapat nomor otomatis saat
backend dijalankan kembali; UUID, teks, dan timestamp lama tetap dipertahankan. API tidak memerlukan konfigurasi baru; tetap memakai
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
- `backend/test/scenario/review/scenario_code_test.go` — baru.
- `backend/test/scenario/review/validation_diagnostics_test.go` — baru.
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
