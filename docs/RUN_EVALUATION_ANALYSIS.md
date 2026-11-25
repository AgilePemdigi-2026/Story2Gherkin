# Analisis proyek dan perbaikan Run Evaluation

Tanggal: 19 November 2026. Basis: kode lokal pada HEAD `f9bf192` beserta perbaikan
yang belum di-commit. Production tidak diakses atau diubah.

## Konteks aplikasi saat ini

Story2Gherkin mengubah kebutuhan menjadi draft Gherkin. React/Vite menyediakan UI;
Go/Gin menangani autentikasi, otorisasi, orkestrasi, dan penyimpanan PostgreSQL;
Python/FastAPI menangani generator LLM, generator rule-based, dan validator.

| Bagian | Alur dan sumber utama |
| --- | --- |
| Identitas | `backend/modules/auth`, middleware JWT, `frontend/src/services/authSession.ts`, `RequireRole.tsx` |
| System Analyst | Generate → review/koreksi/validasi → history dan unduhan; `scenario_service.go`, `review_service.go`, `history_service.go` |
| QA | Upload/kelola Golden Dataset → Run Evaluation → report dan CSV; service dan repository `golden_dataset`, `evaluation_run`, `evaluation_report` |
| AI | `ai-service/app/api/routes`, `services/generator.py`, `services/rule_based.py`, `services/validation.py`; provider Ollama/OpenRouter |
| Penyimpanan | Entity GORM di `backend/database/entities`, migrasi startup, wiring `backend/providers/core.go` |
| Deployment | Docker Compose dan volume PostgreSQL; CI menguji backend/frontend; CD memicu Coolify setelah push `main` dengan CI sukses |

Kontrak generate saat ini sudah memakai `user_story`, `gui_elements`, dan
`free_scenario` pada Go dan Python. Pernyataan README lama tentang frontend scaffold
dan perbedaan kontrak sudah tidak menggambarkan kode saat ini.

Penelusuran mencakup entry point, dependency injection, routing/role, kontrak antar
service, entity/migrasi, repository, alur UI, pengujian, serta CI/CD. Graphify
berhasil membangun graph kode: 1.535 node dan 3.607 edge. Parser SQL opsional belum
terpasang, sehingga graph tersebut tidak mencakup isi SQL secara lengkap.

## Alur evaluasi

1. QA memilih dataset. `POST /api/v1/evaluation-runs` memeriksa role, dataset,
   item, dan keberadaan run aktif.
2. Backend menyimpan run `QUEUED`, snapshot referensi kasus, dan `used_at` dataset.
3. Goroutine mengubah status menjadi `RUNNING`. Setiap kasus dikirim berurutan ke
   generator LLM dan rule-based, lalu ke validator yang sama.
4. Hasil setiap metode disimpan sebagai JSONB; `processed=true` menandai kasus
   selesai diproses, termasuk ketika kedua metode gagal menghasilkan hasil.
5. Backend menghitung metrik dan menyimpan `COMPLETED`, atau `FAILED` bila terjadi
   kegagalan database.
6. Browser membaca progres setiap 1,5 detik. Report membaca hasil tersimpan;
   membukanya tidak menjalankan evaluasi ulang.

`PASS` berarti sintaks valid dan coverage 100%; `WARN` berarti valid dengan
coverage kurang dari 100%; `FAIL` berarti tidak valid. Kegagalan generate/validate
membuat hasil metode `null`, tanpa menghentikan seluruh evaluasi.

Validation percent memakai seluruh kasus sebagai penyebut. Coverage adalah
rata-rata hasil metode yang tersedia. Golden Gherkin disimpan sebagai referensi;
belum ada skor kesamaan semantik dengan Golden. Hal ini sesuai desain yang
disetujui di `docs/superpowers/specs/2026-11-19-run-evaluation-design.md`, tetapi
teks UI “Compare with Golden” bisa menimbulkan ekspektasi berbeda.

## Akar masalah yang direproduksi

`backend/config/database.go` memakai `PreferSimpleProtocol: true`.
`SaveCaseResult` dan `Finish` sebelumnya mengirim pointer struct langsung melalui
`Updates(map[string]any{...})` untuk kolom JSONB.

Pada GORM versi yang dipakai proyek, cabang update map memasukkan nilai map
langsung sebagai parameter SQL, melewati serializer field. pgx simple protocol
tidak mengetahui tipe JSONB parameter tersebut dan gagal meng-encode struct:

```text
unable to encode &entities.EvaluationCaseMethodResult{...}
into text format for unknown type (OID 0): cannot find encode plan

unable to encode &entities.EvaluationRunMetrics{...}
into text format for unknown type (OID 0): cannot find encode plan
```

Kedua error terbukti pada tes PostgreSQL 16 terisolasi sebelum perbaikan. Service
sebelumnya mengabaikan error penyimpanan kasus. Penyimpanan metrik akhir juga
gagal, lalu jalur fallback menyimpan `FAILED` dengan pesan publik internal error.

HTTP 200 pada `GET .../progress` tetap benar: pembacaan status berhasil, sementara
`data.status` berisi `FAILED`. Keberhasilan pekerjaan harus dibaca dari status
domain, bukan status HTTP polling.

Tes lama memakai repository tiruan untuk Run Evaluation. Helper integrasi report
juga memakai protokol default, sehingga belum mewakili konfigurasi production.

## Perubahan

- Repository memakai update struct dengan `Select` eksplisit. Serializer JSON
  dijalankan dan nilai nil tetap ditulis sebagai SQL NULL.
- Service menghentikan run saat hasil kasus gagal disimpan, tanpa membuat metrik
  sukses dari data yang belum tersimpan.
- Log backend mencatat `run_id`, tahap kegagalan, dan error. Kegagalan menyimpan
  status FAILED juga dicatat; pesan publik tetap tidak memuat detail database.
- Tombol evaluasi dapat digunakan lagi setelah FAILED/COMPLETED. Retry menghapus
  referensi run sebelumnya dan respons polling yang telah dibatalkan diabaikan.
- Tes regresi PostgreSQL memakai simple protocol yang sama seperti aplikasi.

## Verifikasi

- Sebelum perubahan: dua tes repository gagal dengan error encoder di atas;
  tes service gagal karena error save masih menghasilkan COMPLETED.
- Setelah perubahan: tes PostgreSQL menyimpan dan membaca ulang kedua metode,
  progres, metrik pecahan/nol/null, serta transisi FAILED. Dua metode nil juga
  diverifikasi sebagai SQL NULL dengan `processed=true`.
- `go test -race -count=1 -timeout=5m ./...` dengan `TEST_DATABASE_URL` ke database
  PostgreSQL sementara: lulus. Tes yang memerlukan environment integrasi lain
  masih mengikuti skip bawaan masing-masing.
- `go vet ./...`: lulus.
- `npm run lint` dan `npm run build`: lulus.
- Playwright/Chrome dengan API tiruan: masalah tombol terkunci direproduksi
  sebelum perubahan. Setelah perubahan, retry setelah FAILED dan COMPLETED,
  penolakan retry oleh server, penghentian polling run lama, dan link report baru
  berhasil diverifikasi. Ini bukan pengujian terhadap production atau model nyata.
- Tes unit Python rule-based dan review validation: 36 lulus; tes API review:
  1 lulus. Kode Python tidak diubah; kualitas generasi LLM nyata belum diukur
  pada pekerjaan ini.

Tes PostgreSQL dapat dijalankan dari `backend`:

```bash
TEST_DATABASE_URL='<DSN database test>' go test -count=1 -v \
  ./test/scenario/evaluation_report -run TestEvaluationRunPostgres
```

Helper hanya menerima nama database `story2gherkin_test` atau
`story2gherkin_migration_test`, memakai schema sementara, lalu membersihkannya.
CI backend sudah menyediakan `TEST_DATABASE_URL`, sehingga tes baru ikut berjalan
tanpa menambahkan konfigurasi workflow.

## Report yang masih terlihat setelah dihapus

Pengguna menyatakan yang dihapus adalah evaluation run/report. Belum ada bukti
langsung untuk menetapkan penyebab perilaku production tersebut.

Dalam kode saat ini, daftar/detail report dibaca dari PostgreSQL dan fetch browser
memakai `cache: 'no-store'`. Tidak ditemukan cache report di service backend.
Komponen tetap menyimpan respons terakhir sampai dimuat ulang atau dinavigasi.
Menghapus `evaluation_runs` secara permanen tidak memiliki jalur otomatis untuk
menciptakan ulang row itu. Menghapus case saja tidak menghapus parent run.

Langkah diagnosis tanpa mengubah data:

1. Muat ulang halaman; periksa URL API dan `run_id` pada respons network.
2. Cocokkan run ID, bukan hanya report code/nama dataset, dengan database yang
   dipakai instance backend production.
3. Pada koneksi database tersebut, jalankan query baca berikut:

```sql
SELECT current_database(), current_schema(), inet_server_addr(), inet_server_port();

SELECT id, run_number, status, total_cases, created_at, updated_at
FROM evaluation_runs
ORDER BY created_at DESC
LIMIT 20;

SELECT run_id, count(*) AS cases,
       count(*) FILTER (WHERE processed) AS processed_cases
FROM evaluation_run_cases
GROUP BY run_id;
```

Jika API masih mengembalikan ID yang tidak ada pada database yang diperiksa,
periksa host/database/schema, target API frontend/proxy, dan instance deployment.
Jika ID berbeda, mungkin yang terlihat adalah run baru; itu perlu dibuktikan
dengan ID dan waktu. Rebuild aplikasi tidak membersihkan volume PostgreSQL.

Perubahan lokal ini belum di-deploy. Backend dan frontend perlu dibangun ulang
dan dirilis; run FAILED lama tetap berupa riwayat dan tidak otomatis dihitung ulang.
Perbaikan ini tidak memerlukan migrasi schema atau penghapusan data production.

## Risiko lanjutan di luar perbaikan

- `HasActive` dan create belum atomik: POST bersamaan berpotensi membuat dua run.
- Worker berada dalam proses API; restart dapat meninggalkan QUEUED/RUNNING
  tanpa mekanisme pemulihan.
- Halaman belum menemukan kembali run aktif setelah reload/navigasi.
- Polling interval dapat tumpang tindih jika respons lebih lama dari interval.
- Pilihan dataset hanya mengambil halaman pertama, maksimal 100 dataset.

## Sumber yang diperiksa

- Kode proyek dan tes yang disebutkan di atas; implementasi dependency lokal
  `gorm.io/gorm@v1.31.2/callbacks/update.go` dan encoder pgx v5.
- [GORM serializer](https://gorm.io/docs/serializer.html) dan
  [GORM updates](https://gorm.io/docs/update.html), diambil melalui Context7.
- [React: synchronizing with Effects](https://react.dev/learn/synchronizing-with-effects),
  diambil melalui Context7 untuk penanganan respons fetch yang sudah tidak relevan.

Riset berfokus pada implementasi dan dokumentasi primer teknis. Tidak ada klaim
hasil benchmark, riset ilmiah, akses production, atau kualitas model yang belum diuji.
