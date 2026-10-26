# language: id
Fitur: Registrasi pengguna
  Sebagai calon pengguna
  Saya ingin mendaftar akun
  Agar bisa masuk ke sistem

  Skenario: Register sukses
    Dengan tidak ada pengguna beremail "bdd@example.com"
    Ketika saya mendaftar dengan nama "Budi", email "bdd@example.com", password "rahasia123", role "qa"
    Maka status respons harus 201
    Dan respons berisi email "bdd@example.com"

  Skenario: Register email duplikat
    Dengan sudah ada pengguna beremail "duplikat@example.com"
    Ketika saya mendaftar dengan nama "Budi", email "duplikat@example.com", password "rahasia123", role "qa"
    Maka status respons harus 409

  Skenario: Register body tidak valid
    Ketika saya mendaftar dengan nama "B", email "bukan-email", password "pendek", role "qa"
    Maka status respons harus 400
