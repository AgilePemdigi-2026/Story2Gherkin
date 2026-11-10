# Backend BDD: Review Scenario

`features/scenario_review.feature` memuat perilaku Review dalam bahasa Indonesia.
Step definitions berada di `review_steps_test.go`, dengan adapter fixture pada
`review_support_test.go`. `TestReviewFeatures` menjalankan suite secara strict:
step yang belum diimplementasikan atau gagal akan menggagalkan tes.

## Menjalankan

Dari folder backend:

```powershell
go test -count=1 -timeout=90s ./test/bdd -run '^TestReviewFeatures$' -v
```

Seluruh BDD backend, termasuk Register dan Generate:

```powershell
go test -count=1 -timeout=90s ./test/bdd -v
```

Workflow CI yang sudah menjalankan `go test -race -count=1 -timeout=5m ./...`
otomatis menjalankan suite ini. Tidak perlu mengubah workflow atau dependensi.
Pada Windows lokal, gunakan perintah di atas tanpa `-race` jika compiler C/CGO
belum tersedia.

## Cakupan

- Daftar dan detail milik pengguna, kode SCN, serta UUID untuk lookup.
- Koreksi valid dengan coverage 100%, 75%, atau 0%; status CORRECTED/VALID,
  metadata coverage, byte teks, dan pemuatan kembali detail.
- Koreksi invalid: detail kesalahan diteruskan, data tersimpan tetap sama,
  download berikutnya memakai versi valid tersimpan.
- Download byte persis versi tersimpan beserta nama file dan header yang sesuai;
  versi tersimpan INVALID ditolak.
- Semua operasi menolak akses tanpa login; token rusak juga ditolak.
- QA tidak memperoleh akses, termasuk jika JWT masih mengklaim System Analyst
  sementara role pengguna yang tersimpan adalah QA.
- Pengguna lain tidak dapat mengakses detail, menyimpan, atau download skenario;
  daftar hanya memuat record milik pengguna dan mendukung hasil kosong.
- Versi koreksi lama ditolak sebelum validasi, tanpa perubahan data.
- Validator tidak tersedia, timeout, atau respons rusak menghasilkan status yang
  sesuai tanpa menulis ke penyimpanan.
- Payload tanpa updated_at ditolak sebelum validator dipanggil.

## Batas pengujian

Middleware JWT, penandatanganan/verifikasi token, handler HTTP, binding request,
dan ReviewService yang digunakan adalah implementasi produksi. Repository user,
repository skenario, dan laporan validator merupakan adapter terkontrol di memori.
Setiap skenario memiliki state baru, dan assertion memeriksa respons HTTP,
perubahan data, jumlah write, serta pemanggilan validator.

Suite ini menguji kontrak dan keputusan backend berdasarkan hasil validator,
bukan menjalankan parser Python atau query PostgreSQL sebenarnya. Kebenaran
parser, SQL kepemilikan, dan migrasi/sequence SCN tetap diperiksa oleh unit/API
Python serta tes integrasi pada `test/scenario/review`.

Tidak diperlukan PostgreSQL, Python, Ollama, layanan jaringan, atau `.env`.
Signing key khusus fixture dipasang sementara melalui `t.Setenv` dan tidak
mengubah konfigurasi aplikasi atau proses backend yang sedang berjalan.
