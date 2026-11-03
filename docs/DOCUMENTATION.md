# Documentation

Halaman publik `/documentation` memakai sidebar topik, area artikel, breadcrumb,
dan daftar isi per artikel dengan identitas visual Story2Gherkin. Judul bagian
berupa teks biasa tanpa tautan langsung.

Konten aplikasi berada di `frontend/src/content/documentation.tsx`. Setiap artikel
memiliki slug, kelompok navigasi, deskripsi, kata pencarian, dan section dengan ID
stabil. Tambahkan topik melalui data tersebut agar sidebar, daftar isi, dan
navigasi sebelumnya/selanjutnya ikut diperbarui.

## Navigasi

- `/documentation` membuka Pengenalan.
- `/documentation?topic=review` membuka Review Scenario.
- `/documentation?topic=review#menyimpan` membuka bagian Simpan Perubahan.
- Tautan lama `#mulai`, `#system-analyst`, `#qa`, dan `#validasi` tetap didukung.
- Slug yang tidak dikenal menampilkan Pengenalan.
- Pencarian memfilter judul, deskripsi, dan kata kunci topik, bukan isi penuh
  artikel. Grup navigasi dapat dibuka/ditutup.
- Pada layar kecil, topik dibuka melalui tombol Topik dokumentasi dan daftar isi
  ditampilkan sebagai panel yang dapat dibuka di dalam artikel.

Contoh Gherkin memiliki tombol Salin. Jika izin clipboard ditolak, tampilkan
petunjuk menyalin manual. Tidak ada panggilan API, login wajib, atau layanan
eksternal yang diperlukan untuk membaca dokumentasi.

## Cakupan isi

Topik meliputi Pengenalan, Mulai Menggunakan, Generate, Review, Evaluation untuk
QA, Dasar Gherkin, Validation & GUI Coverage, dan Pertanyaan Umum.
Panduan menjelaskan aturan review yang didukung Story2Gherkin. Status fitur
Evaluation tetap ditulis sebagai dalam pengembangan.

CSS `documentation-card`, `documentation-action`, `documentation-icon`, dan
`documentation-status` masih dipakai halaman awal Evaluation. Periksa halaman QA
jika mengubah kelas tersebut.

## Pemeriksaan

Jalankan lint dan build frontend. Periksa perpindahan seluruh topik, back/forward,
tautan langsung dan fragment lama, pencarian kosong/tanpa hasil, salin contoh
dengan/tanpa izin, FAQ, dan viewport desktop/tablet/ponsel. Dokumentasi harus
dapat dibuka sebagai pengunjung dan tautan Panduan QA tetap menuju artikel QA.
