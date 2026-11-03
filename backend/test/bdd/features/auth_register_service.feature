# language: id
Fitur: Registrasi pengguna via service
  Sebagai calon pengguna
  Saya ingin mendaftar akun lewat service
  Agar nama/email dinormalisasi dan password ter-hash

  Skenario: Register sukses normalisasi dan hash
    Dengan tidak ada pengguna beremail "rizal@example.com"
    Ketika saya register service dengan nama "  Rizal  ", email "  RIZAL@Example.COM ", password "rahasia123", role "qa"
    Maka register service berhasil dengan nama "Rizal", email "rizal@example.com", role "qa" dan password ter-hash

  Skenario: Register email duplikat
    Dengan sudah ada pengguna beremail "rizal@example.com"
    Ketika saya register service dengan nama "Rizal", email "rizal@example.com", password "rahasia123", role "qa"
    Maka register service gagal email sudah ada dan repository Register tidak dipanggil

  Skenario: Register gagal saat cek email
    Dengan cek email error "koneksi database putus"
    Ketika saya register service dengan nama "Rizal", email "rizal@example.com", password "rahasia123", role "qa"
    Maka register service gagal dengan error "koneksi database putus"

  Skenario: Register gagal saat insert
    Dengan tidak ada pengguna beremail "rizal@example.com" dan insert error "insert gagal"
    Ketika saya register service dengan nama "Rizal", email "rizal@example.com", password "rahasia123", role "qa"
    Maka register service gagal dengan error "insert gagal"

  Skenario: Register password terlalu panjang
    Dengan tidak ada pengguna beremail "rizal@example.com"
    Ketika saya register service dengan nama "Rizal", email "rizal@example.com", password sepanjang 73 karakter, role "qa"
    Maka register service gagal dan repository Register tidak dipanggil
