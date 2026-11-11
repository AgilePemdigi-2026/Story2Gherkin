
# language: id
Fitur: Pengelolaan Golden Dataset
  Sebagai Quality Assurance
  Saya ingin mengunggah dan mengelola Golden Dataset
  Agar tersedia skenario Gherkin referensi untuk Run Evaluation

  Skenario: QA berhasil mengunggah Golden Dataset valid
    Dengan pengguna Golden Dataset login sebagai "QA"
    Dan file Golden Dataset berisi 2 item valid
    Dan validator Golden Dataset menyatakan semua Gherkin valid
    Ketika pengguna mengunggah Golden Dataset
    Maka status respons Golden Dataset harus 201
    Dan Golden Dataset tersimpan dengan 2 item
    Dan validator Golden Dataset dipanggil 2 kali

  Skenario: System Analyst tidak dapat mengunggah Golden Dataset
    Dengan pengguna Golden Dataset login sebagai "SA"
    Dan file Golden Dataset berisi 2 item valid
    Ketika pengguna mengunggah Golden Dataset
    Maka status respons Golden Dataset harus 403
    Dan tidak ada Golden Dataset baru yang tersimpan
    Dan validator Golden Dataset dipanggil 0 kali

  Skenario: Upload Golden Dataset memerlukan autentikasi
    Dengan pengguna Golden Dataset tidak login
    Dan file Golden Dataset berisi 2 item valid
    Ketika pengguna mengunggah Golden Dataset
    Maka status respons Golden Dataset harus 401
    Dan tidak ada Golden Dataset baru yang tersimpan

  Skenario: File JSON yang rusak ditolak
    Dengan pengguna Golden Dataset login sebagai "QA"
    Dan file Golden Dataset berisi JSON yang rusak
    Ketika pengguna mengunggah Golden Dataset
    Maka status respons Golden Dataset harus 422
    Dan tidak ada Golden Dataset baru yang tersimpan
    Dan validator Golden Dataset dipanggil 0 kali

  Skenario: Item dengan judul duplikat ditolak
    Dengan pengguna Golden Dataset login sebagai "QA"
    Dan file Golden Dataset memiliki judul item duplikat
    Ketika pengguna mengunggah Golden Dataset
    Maka status respons Golden Dataset harus 422
    Dan tidak ada Golden Dataset baru yang tersimpan

  Skenario: Gherkin referensi yang invalid ditolak
    Dengan pengguna Golden Dataset login sebagai "QA"
    Dan file Golden Dataset berisi 1 item valid
    Dan validator Golden Dataset menyatakan Gherkin invalid
    Ketika pengguna mengunggah Golden Dataset
    Maka status respons Golden Dataset harus 422
    Dan tidak ada Golden Dataset baru yang tersimpan

  Skenario: Kegagalan validator tidak menyimpan dataset
    Dengan pengguna Golden Dataset login sebagai "QA"
    Dan file Golden Dataset berisi 1 item valid
    Dan layanan validator Golden Dataset tidak tersedia
    Ketika pengguna mengunggah Golden Dataset
    Maka status respons Golden Dataset harus 503
    Dan tidak ada Golden Dataset baru yang tersimpan

  Skenario: Upload yang melebihi batas ukuran ditolak
    Dengan pengguna Golden Dataset login sebagai "QA"
    Dan file Golden Dataset melebihi batas 5 MB
    Ketika pengguna mengunggah Golden Dataset
    Maka status respons Golden Dataset harus 413
    Dan tidak ada Golden Dataset baru yang tersimpan

  Skenario: QA dapat melihat daftar Golden Dataset
    Dengan terdapat 1 Golden Dataset yang sudah tersimpan
    Dan pengguna Golden Dataset login sebagai "QA"
    Ketika pengguna meminta daftar Golden Dataset
    Maka status respons Golden Dataset harus 200
    Dan daftar Golden Dataset berisi 1 dataset

  Skenario: QA dapat melihat detail Golden Dataset
    Dengan terdapat 1 Golden Dataset yang sudah tersimpan
    Dan pengguna Golden Dataset login sebagai "QA"
    Ketika pengguna meminta detail Golden Dataset
    Maka status respons Golden Dataset harus 200
    Dan detail Golden Dataset memiliki 2 item

  Skenario: Dataset yang tidak ditemukan menghasilkan 404
    Dengan pengguna Golden Dataset login sebagai "QA"
    Ketika pengguna meminta detail Golden Dataset yang tidak tersedia
    Maka status respons Golden Dataset harus 404

  Garis-Besar Skenario: System Analyst tidak boleh melihat Golden Dataset
    Dengan terdapat 1 Golden Dataset yang sudah tersimpan
    Dan pengguna Golden Dataset login sebagai "SA"
    Ketika pengguna menjalankan operasi Golden Dataset "<operasi>"
    Maka status respons Golden Dataset harus 403

    Contoh:
      | operasi |
      | daftar  |
      | detail  |
