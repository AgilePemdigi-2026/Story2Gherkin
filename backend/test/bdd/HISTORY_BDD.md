
# Backend BDD: Generation History

`features/generation_history.feature` memuat perilaku Generation History dalam bahasa Indonesia.

Step definitions berada di `history_steps_test.go`, dengan adapter fixture pada `history_support_test.go`.

`TestHistoryFeatures` menjalankan suite menggunakan strict mode. Step yang belum diimplementasikan atau gagal akan menggagalkan tes.

## Menjalankan

Dari folder `backend`:

```powershell
go test -count=1 -timeout=90s ./test/bdd -run '^TestHistoryFeatures$' -v
```

Seluruh BDD backend, termasuk Register, Generate, Review, dan Golden Dataset:

```powershell
go test -count=1 -timeout=90s ./test/bdd -v
```

Seluruh pengujian, pemeriksaan kode, dan build backend:

```powershell
go test ./...
go vet ./...
go build ./...
```

Workflow CI yang sudah menjalankan `go test -race -count=1 -timeout=5m ./...` akan otomatis menjalankan suite History.

Pada Windows lokal, gunakan perintah tanpa `-race` apabila compiler C/CGO belum tersedia.

## Cakupan

- Daftar History hanya menampilkan Generation Request milik System Analyst yang sedang login, baik yang berstatus SUCCESS maupun FAILED.
- Pengguna tanpa riwayat menerima daftar kosong dengan HTTP 200.
- Pagination mendukung pengaturan page dan page_size, termasuk halaman yang tidak memiliki data.
- Parameter pagination yang tidak valid ditolak dengan HTTP 400.
- Detail History SUCCESS memuat User Story, GUI Elements, Free Scenario, dan Generated Gherkin yang sesuai dengan data tersimpan.
- Detail History FAILED tetap menampilkan input asli dan error message, sementara scenario bernilai null.
- History milik pengguna lain tidak dapat diakses dan menghasilkan HTTP 404 tanpa membocorkan isi data.
- UUID tidak valid dan History yang tidak ditemukan menghasilkan HTTP 404.
- Endpoint List dan Detail membutuhkan autentikasi JWT. Request tanpa login maupun dengan token rusak menghasilkan HTTP 401.
- Quality Assurance (QA) tidak dapat mengakses Generation History dan menerima HTTP 403.
- Hak akses mengikuti role pengguna yang tersimpan di database, bukan hanya klaim role dalam JWT.
- Setiap skenario memiliki fixture terisolasi agar data pengujian tidak saling memengaruhi.

## Mekanisme Pengujian

Middleware JWT, penandatanganan/verifikasi token, HistoryHandler, dan HistoryService yang digunakan merupakan implementasi produksi.

Dependensi eksternal digantikan dengan adapter terkontrol di memori:

- `historyBDDUsers` menyediakan akun System Analyst pemilik, System Analyst lain, dan Quality Assurance.
- `historyBDDStore` menyimulasikan penyimpanan Generation Request dan Generated Gherkin.
- `seedSuccess()` menyediakan request SUCCESS dengan hasil Gherkin lengkap.
- `seedFailed()` menyediakan request FAILED tanpa Generated Gherkin.
- `seedOther()` menyediakan History milik System Analyst lain.
- `historyBDDRouter()` menghubungkan middleware JWT, Handler, dan Service asli menggunakan httptest.

Setiap skenario dimulai dengan state baru melalui reset fixture.

Assertion tidak hanya memeriksa respons HTTP, tetapi juga memeriksa User Story, GUI Elements, Free Scenario, Generated Gherkin, metadata skenario, status, pagination, serta tidak adanya kebocoran data pengguna lain.

## Hasil Pengujian

Jalankan:

```powershell
go test -count=1 -timeout=90s ./test/bdd -run '^TestHistoryFeatures$' -v
```

File feature dirancang mencakup 19 eksekusi skenario, termasuk setiap baris Scenario Outline.

Hasil pengujian dicatat setelah suite berhasil dijalankan. Pastikan seluruh skenario dan langkah berstatus passed, tanpa undefined step atau failed assertion.

Pengujian regresi dilakukan melalui:

```powershell
go test -count=1 -timeout=90s ./test/bdd -v
go test ./...
go vet ./...
go build ./...
```

## Batas Pengujian

Suite ini menguji kontrak HTTP dan keputusan backend menggunakan repository pengguna dan History tiruan, bukan menjalankan query PostgreSQL sebenarnya.

Ketepatan query kepemilikan data pada PostgreSQL perlu diperiksa melalui pengujian integrasi atau pengujian API terhadap database pengujian.

BDD ini tidak menjalankan LLM Generator atau Gherkin Validator Python karena endpoint History hanya membaca data yang sebelumnya sudah tersimpan.

Data SUCCESS dan FAILED dalam fixture merupakan simulasi, bukan hasil Generate dari AI Service sebenarnya.

Tidak diperlukan PostgreSQL, Python, Ollama, layanan jaringan eksternal, atau .env untuk menjalankan BDD ini.

Signing key khusus fixture dipasang sementara melalui t.Setenv dan tidak mengubah konfigurasi backend yang sedang berjalan.
