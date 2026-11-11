
# Backend BDD: Golden Dataset

`features/golden_dataset.feature` memuat perilaku pengelolaan Golden Dataset dalam bahasa Indonesia.

Step definitions berada di `golden_dataset_steps_test.go`, dengan adapter fixture pada `golden_dataset_support_test.go`.

`TestGoldenDatasetFeatures` menjalankan suite menggunakan strict mode. Step yang belum diimplementasikan atau gagal akan menggagalkan tes.

## Menjalankan

Dari folder `backend`:

```powershell
go test -count=1 -timeout=90s ./test/bdd -run '^TestGoldenDatasetFeatures$' -v
```

Seluruh BDD backend, termasuk Register, Generate, dan Review:

```powershell
go test -count=1 -timeout=90s ./test/bdd -v
```

Seluruh pengujian, pemeriksaan kode, dan build backend:

```powershell
go test ./...
go vet ./...
go build ./...
```

Workflow CI yang sudah menjalankan `go test -race -count=1 -timeout=5m ./...` otomatis menjalankan suite Golden Dataset. Tidak perlu mengubah workflow atau dependensi.

Pada Windows lokal, gunakan perintah tanpa `-race` jika compiler C/CGO belum tersedia.

## Cakupan

- Quality Assurance (QA) dapat mengunggah Golden Dataset dengan dua item valid. Respons menghasilkan HTTP 201, dataset tersimpan dengan jumlah item yang benar, dan validator dipanggil untuk setiap item.
- System Analyst tidak dapat mengunggah Golden Dataset. Respons menghasilkan HTTP 403 tanpa perubahan penyimpanan atau pemanggilan validator.
- Upload tanpa autentikasi ditolak dengan HTTP 401 dan tidak menyimpan dataset.
- File JSON yang rusak ditolak dengan HTTP 422 sebelum validator dipanggil.
- Item dengan judul duplikat ditolak dengan HTTP 422 tanpa menyimpan dataset.
- Gherkin referensi yang dinyatakan invalid oleh validator ditolak dengan HTTP 422 tanpa menyimpan dataset.
- Validator yang tidak tersedia menghasilkan HTTP 503 tanpa menyimpan dataset.
- Upload yang melebihi batas ukuran 5 MiB ditolak dengan HTTP 413 tanpa menyimpan dataset.
- QA dapat mengambil daftar Golden Dataset melalui GET List dengan HTTP 200.
- QA dapat mengambil detail Golden Dataset beserta item, GUI Elements, dan Expected Gherkin dengan HTTP 200.
- Dataset yang tidak ditemukan menghasilkan HTTP 404.
- System Analyst tidak dapat mengakses GET List maupun GET Detail, masing-masing menghasilkan HTTP 403.

## Mekanisme Pengujian

Middleware JWT, penandatanganan/verifikasi token, handler HTTP, binding multipart request, dan GoldenDatasetService yang digunakan adalah implementasi produksi.

Dependensi eksternal digantikan dengan adapter terkontrol di memori:

- `goldenBDDUsers` menyediakan pengguna QA dan System Analyst.
- `goldenBDDStore` mensimulasikan penyimpanan dan pengambilan Golden Dataset beserta item.
- `goldenBDDValidator` mensimulasikan hasil validasi Gherkin valid, invalid, dan layanan validator yang tidak tersedia.
- `goldenBDDRouter` menghubungkan middleware, handler, dan service asli untuk menerima request melalui `httptest`.

Setiap skenario memiliki state baru melalui reset fixture. Assertion memeriksa respons HTTP, jumlah dataset dan item tersimpan, jumlah write, serta jumlah pemanggilan validator.

File JSON disiapkan sebagai fixture dalam pengujian dan dikirim melalui multipart/form-data. Dengan demikian, binding dan batas ukuran upload pada handler tetap diuji.

## Hasil Pengujian

Pengujian dijalankan menggunakan:

```powershell
go test -count=1 -timeout=90s ./test/bdd -run '^TestGoldenDatasetFeatures$' -v
```

Hasil:

```text
13 scenarios (13 passed)
67 steps (67 passed)
--- PASS: TestGoldenDatasetFeatures
PASS
```

Pengujian seluruh suite BDD backend juga berhasil:

| Suite | Skenario | Hasil |
|---|---:|---|
| Register Service | 5 | Passed |
| Register HTTP | 3 | Passed |
| Golden Dataset | 13 | Passed |
| Review Scenario | 27 | Passed |
| Generate Scenario | 3 | Passed |
| **Total** | **51** | **Passed** |

Total seluruh suite adalah 315 langkah yang lulus.

Perintah `go test ./...`, `go vet ./...`, dan `go build ./...` juga berhasil dijalankan tanpa error pada pengujian tersebut.

## Batas Pengujian

Suite ini menguji kontrak HTTP dan keputusan backend berdasarkan hasil validator, bukan menjalankan parser Python atau query PostgreSQL sebenarnya.

Fake repository menyimulasikan penyimpanan dataset dan item, sehingga keberhasilan BDD belum membuktikan transaksi GORM dan rollback PostgreSQL.

Pengujian terhadap transaksi database, foreign key, serta kegagalan penyimpanan item perlu dilakukan melalui integration test menggunakan PostgreSQL.

Pengujian ini juga tidak membuktikan bahwa parser Python menerima seluruh contoh Expected Gherkin karena validator yang digunakan merupakan fake validator.

Integrasi upload dengan validator Python sebenarnya perlu diuji ketika AI Service sudah tersedia.

Tidak diperlukan PostgreSQL, Python, Ollama, layanan jaringan eksternal, atau `.env` untuk menjalankan suite BDD ini.

Signing key khusus fixture dipasang sementara melalui `t.Setenv` dan tidak mengubah konfigurasi aplikasi atau proses backend yang sedang berjalan.
