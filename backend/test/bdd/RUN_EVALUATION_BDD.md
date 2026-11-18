# Backend BDD: Run Evaluation

`features/run_evaluation.feature` memuat perilaku Run Evaluation dalam bahasa Indonesia.
Step definitions berada di `run_evaluation_steps_test.go`, dengan adapter fixture pada
`run_evaluation_support_test.go`. `TestRunEvaluationFeatures` menjalankan suite dalam
strict mode, sehingga step yang belum diimplementasikan atau gagal menggagalkan tes.

## Menjalankan

Dari folder `backend`:

```powershell
go test -count=1 -timeout=90s ./test/bdd -run '^TestRunEvaluationFeatures$' -v
```

## Cakupan

- QA memulai evaluasi: HTTP 201, `total_cases` sesuai dataset, run selesai `COMPLETED`
  dengan seluruh kasus diproses dan metrik kedua metode tersimpan.
- System Analyst dan pengguna tanpa token ditolak (403 dan 401) tanpa menyimpan run.
- Run aktif (`QUEUED`/`RUNNING`) membuat permintaan baru ditolak dengan 409.
- Golden Dataset tanpa item ditolak 400; dataset yang tidak ada menghasilkan 404.
- Cakupan di bawah 100 persen menghasilkan kasus `WARN`; Gherkin tidak valid
  menghasilkan `FAIL`.
- Progress menampilkan `processed_cases`/`total_cases`; run yang tidak ada
  menghasilkan 404.

## Mekanisme

Middleware JWT, `EvaluationRunHandler`, dan `EvaluationRunService` adalah implementasi
produksi. Dependensi eksternal diganti adapter di memori:

- `goldenBDDStore` (dipakai ulang dari suite Golden Dataset) menyediakan dataset dan item.
- `evalBDDRunRepo` menyimulasikan tabel `evaluation_runs` dan `evaluation_run_cases`.
- `evalBDDClient` menyimulasikan generator LLM, generator Rule-Based, dan validator Gherkin.
- `evalBDDRouter()` merangkai middleware JWT dan route produksi melalui httptest.

Karena run dieksekusi di latar belakang, assertion akhir menunggu status `COMPLETED`
atau `FAILED` maksimal 2 detik sebelum memeriksa metrik dan hasil kasus.

## Batas Pengujian

Suite ini menguji kontrak HTTP dan orkestrasi memakai repository tiruan, bukan query
PostgreSQL sebenarnya. Query `HasActive`, penyimpanan hasil, dan perhitungan progres
perlu diverifikasi melalui pengujian integrasi basis data. Python, Ollama, dan layanan
jaringan tidak diperlukan.
