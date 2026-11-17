# Kontrak Backend Evaluation Report

Frontend QA menampilkan riwayat, ringkasan metrik LLM dan Rule-Based, hasil
per kasus, serta unduhan CSV. Dokumen ini adalah kontrak yang dipakai frontend
dan perlu disepakati bersama PIC Run Evaluation.

## Pembagian pekerjaan

| PIC | Tanggung jawab |
| --- | --- |
| Run Evaluation | Menjalankan evaluasi, menentukan dan menghitung metrik, menyimpan status serta hasil per metode, dan mengembalikan `run_id`. |
| Backend Evaluation Report | Membaca hasil yang tersimpan, menerapkan otorisasi QA, menyediakan API daftar dan detail, serta menghasilkan CSV. |
| Frontend Evaluation Report | Mengambil data dari API, menampilkan metrik dan hasil per kasus, lalu mengunduh CSV. |

Run Evaluation dan Report memakai penyimpanan run yang sama. Simpan nama
dataset dan Golden Gherkin yang dipakai saat run dijalankan agar riwayat
tetap dapat dibaca apabila dataset berubah. Pertahankan urutan kasus
yang konsisten dengan urutan evaluasi dan gunakan ID kasus yang unik di
tiap run. API Report membaca data tersimpan tanpa menjalankan ulang model.

## Endpoint

Semua endpoint berada di bawah `/api/v1` dan membutuhkan header
`Authorization: Bearer <access_token>`. Terapkan middleware autentikasi
yang sudah ada dan verifikasi role QA di backend. Untuk pembagian akses
antar akun QA, usulan awalnya adalah hanya pemilik run yang dapat
melihat dan mengunduh report. Sepakati aturan ini sebelum menentukan query.

| Metode | Path | Respons sukses |
| --- | --- | --- |
| GET | `/evaluation-runs?page=1&page_size=10` | JSON berisi riwayat, terbaru terlebih dahulu. |
| GET | `/evaluation-runs/:run_id` | JSON berisi ringkasan run dan hasil per kasus. |
| GET | `/evaluation-runs/:run_id/download` | File CSV untuk run yang berstatus `COMPLETED`. |

Gunakan envelope JSON yang sudah dipakai proyek, dengan `status`
bertipe boolean dan payload sukses pada `data`.

### Daftar run

`page` dimulai dari 1. Default `page_size` adalah 10. Batasi ukuran
halaman di backend dan urutkan berdasarkan `created_at DESC`, dengan
`run_id` sebagai pemecah seri. Halaman kosong mengembalikan array
`runs: []`. Jika belum ada run, `total_pages` bernilai 0.

```json
{
  "status": true,
  "message": "Evaluation runs loaded.",
  "data": {
    "runs": [
      {
        "run_id": "6b308226-0b1f-4efb-92c6-9b4066132b74",
        "dataset_name": "Golden Dataset v1",
        "created_at": "2026-11-17T02:00:00Z",
        "total_cases": 1,
        "status": "COMPLETED"
      }
    ],
    "total_pages": 1
  }
}
```

### Detail satu run

`total_cases` adalah jumlah kasus yang direncanakan untuk run.
Status run yang diterima frontend adalah `QUEUED`, `RUNNING`,
`COMPLETED`, dan `FAILED`. `COMPLETED` berarti proses evaluasi selesai.
Kasus individual tetap bisa berstatus `WARN` atau `FAIL`.

Nilai metrik berbentuk angka 0–100, tanpa tanda persen. Gunakan `null`
jika metrik belum tersedia. Nilai `0` berarti hasil hitungnya nol.
Gunakan `metrics: null` sebelum ringkasan tersedia dan `cases: []`
sebelum hasil kasus tersedia. Hasil masing-masing metode boleh `null`
jika belum dievaluasi. Tetapkan arti Validation Metric dan ambang
`PASS`/`WARN`/`FAIL` bersama PIC Run Evaluation. Frontend hanya
menghitung selisih LLM dikurangi Rule-Based untuk tampilan, dalam poin
persentase.

```json
{
  "status": true,
  "message": "Evaluation report loaded.",
  "data": {
    "run_id": "6b308226-0b1f-4efb-92c6-9b4066132b74",
    "dataset_name": "Golden Dataset v1",
    "created_at": "2026-11-17T02:00:00Z",
    "total_cases": 1,
    "status": "COMPLETED",
    "error_message": null,
    "metrics": {
      "llm": { "validation_percent": 100, "gui_coverage_percent": 90 },
      "rule_based": { "validation_percent": 0, "gui_coverage_percent": 60 }
    },
    "cases": [
      {
        "case_id": "GC-002",
        "title": "Registrasi user",
        "expected_gherkin": "Feature: Register\n  Scenario: Valid registration\n    Given the registration form is open\n    When the visitor submits valid details\n    Then the account is created",
        "llm": { "status": "PASS", "gui_coverage_percent": 90 },
        "rule_based": { "status": "FAIL", "gui_coverage_percent": 60 }
      }
    ]
  }
}
```

Untuk status `FAILED`, gunakan `error_message` yang bisa ditampilkan
ke QA tanpa stack trace atau informasi rahasia. Gunakan `null` bila
tidak ada pesan. Halaman QA untuk hasil ini adalah
`/evaluation/reports/:runId`. Tombol **Lihat Report** di Run Evaluation
bisa mengarah ke sana setelah menerima `run_id`.

### Unduhan CSV

Kembalikan CSV UTF-8 dengan `Content-Type: text/csv; charset=utf-8` dan
`Content-Disposition: attachment; filename="evaluation-report-<run_id>.csv"`.
Satu baris mewakili satu kasus. Kolom minimum:

```text
case_id,title,expected_gherkin,llm_status,llm_gui_coverage_percent,rule_based_status,rule_based_gui_coverage_percent
```

Gunakan penulis CSV standar supaya koma, kutip, dan Gherkin multiline
di-escape dengan benar. Tangani teks yang dapat dianggap formula oleh
spreadsheet, dan gunakan sel kosong untuk nilai `null`.

### Error dan kriteria penerimaan

| HTTP | Kondisi |
| --- | --- |
| 400 | ID atau parameter paginasi tidak valid. |
| 401 | Bearer token hilang atau tidak valid. |
| 403 | Pengguna login dengan role selain QA. |
| 404 | Run tidak ditemukan atau berada di luar cakupan akses pengguna. |
| 409 | Run belum berstatus `COMPLETED` saat diminta unduh. |
| 500 | Penyimpanan atau pembuatan laporan gagal. |

- Daftar kosong, urutan hasil, dan paginasi diuji.
- Detail run mengembalikan Golden Gherkin yang digunakan ketika run
  dijalankan, termasuk pada kasus yang hasilnya `FAIL`.
- Run yang sedang berjalan atau gagal tetap dapat dibaca. Metrik yang
  belum dihitung tetap `null`.
- Pengguna tanpa izin tidak dapat membuka daftar, detail, atau unduhan.
- CSV hanya tersedia untuk run selesai dan sesuai dengan data per kasus
  di respons detail.

Untuk melihat UI sebelum API siap, login sebagai QA di Vite dev server
dan buka `/evaluation/reports/demo`. Halaman itu memakai data contoh.
Rute biasa meminta data ke API, dan contoh tidak aktif pada build production.
