---
name: "[Backend] - Register"
about: Fitur Register
title: ''
labels: ''
assignees: ''

---

---
title: "[BE]: Registrasi Pengguna"
labels: task
---

## Problem

Pengguna belum dapat membuat akun secara mandiri melalui aplikasi. Fitur registrasi diperlukan agar pengguna baru dapat mendaftarkan akun dan mengakses layanan yang tersedia.

## Solution

Mengembangkan fitur registrasi pengguna yang memungkinkan pengguna membuat akun menggunakan nama, email, dan kata sandi.

Proses registrasi harus mencakup validasi input, pencegahan penggunaan email yang sudah terdaftar, serta penyimpanan kredensial pengguna secara aman.

## API Contract

### `POST /api/v1/auth/register`

**Catatan**
- Seluruh input harus divalidasi di backend.
- Kata sandi harus disimpan menggunakan algoritma hashing yang aman.
- Kata sandi dan hasil hashing-nya tidak boleh dikembalikan dalam respons API.
- `password_confirmation` hanya digunakan untuk validasi dan tidak perlu disimpan ke database.
- Format respons dan kode status harus mengikuti konvensi API yang digunakan dalam proyek.

## Acceptance Criteria

- [ ] Sistem menolak registrasi jika email sudah terdaftar.
- [ ] Konfirmasi kata sandi harus sesuai dengan kata sandi.
- [ ] Kata sandi di-hash secara aman sebelum disimpan ke database.
- [ ] Registrasi dengan input valid berhasil membuat akun pengguna.
- [ ] API mengembalikan respons dan kode status sesuai kontrak yang disepakati.
- [ ] Input tidak valid menghasilkan respons kesalahan yang sesuai.
- [ ] Pengujian otomatis mencakup registrasi berhasil dan berbagai kondisi kegagalan validasi.
