# Navigasi publik dan ruang kerja

| Kondisi | Menu utama | Tombol Home | Tujuan login default |
| --- | --- | --- | --- |
| Belum login | Home, About, Documentation | Mulai Menggunakan | Login |
| System Analyst | Home, Generate, About, Documentation | Mulai Generate | /generate |
| QA | Home, Evaluation, About, Documentation | Buka Evaluation | /evaluation |

Documentation dapat dibaca tanpa login dan memuat alur penggunaan serta penjelasan
validasi. Evaluation saat ini hanya halaman awal QA dengan status belum tersedia;
Golden Dataset, Run Evaluation, dan Evaluation Report belum diimplementasikan.

## Sesi dan akses

- Role dibaca dari claim `role` pada access JWT yang sudah diterbitkan backend.
  Token yang sudah ada tetap dikenali tanpa menyimpan salinan role terpisah.
- Claim hanya mengatur tampilan browser. Validasi tanda tangan JWT dan izin setiap
  operasi API tetap menjadi tanggung jawab backend.
- Login memakai kontrak backend aktual: `access_token`, `refresh_token`, `role`.
  Remember Me memakai localStorage; sesi biasa memakai sessionStorage. Login baru
  membersihkan token sesi sebelumnya dari kedua storage.
- Generate dan Review hanya tersedia untuk System Analyst; Evaluation hanya untuk
  QA. Pengunjung diarahkan ke Login, sedangkan role yang salah diarahkan ke ruang
  kerjanya sendiri. Redirect login dibatasi ke rute internal yang sesuai role.
- Token kedaluwarsa, tidak terbaca, atau role tidak dikenal menampilkan navigasi
  publik. Perubahan sesi, storage di tab lain, dan waktu kedaluwarsa memperbarui UI.
- Konfirmasi draft sebelum logout tetap berlaku. Logout yang dibatalkan tidak
  menghapus sesi atau draft; logout yang dikonfirmasi kembali ke Home.

## Verifikasi

Jalankan `npm run lint` dan `npm run build` dari frontend. Periksa di browser:

1. Tanpa login, Home/About/Documentation terbuka dan menu fitur tidak muncul.
2. Login sebagai SA membuka Generate; login QA membuka Evaluation.
3. Buka URL lintas role secara langsung; halaman fitur tidak terbuka dan API
   skenario tidak dipanggil oleh halaman yang ditolak.
4. Refresh halaman, gunakan Remember Me, dan logout. Menu serta tombol Home sesuai
   sesi yang aktif; token lama tidak tertinggal di storage lainnya.
5. Dari Review dengan draft, batalkan lalu konfirmasi logout.
6. Periksa menu ponsel, token kedaluwarsa, dan kedua role dari tab yang berbeda.

Tidak diperlukan perubahan database, backend, dependensi, atau konfigurasi .env.
