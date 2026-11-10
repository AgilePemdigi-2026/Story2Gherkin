# language: id
Fitur: Review dan koreksi skenario Gherkin
  Sebagai System Analyst
  Saya ingin meninjau dan mengoreksi skenario milik saya
  Agar versi valid dapat disimpan dan diunduh tanpa kehilangan data sebelumnya

  Latar Belakang:
    Dengan ada skenario review tersimpan dengan kode "SCN-0001"

  Skenario: Daftar hanya menampilkan skenario milik pengguna
    Dengan pengguna review login sebagai "pemilik"
    Ketika pengguna menjalankan operasi review "daftar"
    Maka status respons review harus 200
    Dan daftar review hanya memuat skenario miliknya dengan kode "SCN-0001"
    Dan validator review dipanggil 0 kali

  Skenario: Detail menampilkan kode tetap dengan UUID untuk lookup
    Dengan pengguna review login sebagai "pemilik"
    Ketika pengguna menjalankan operasi review "detail"
    Maka status respons review harus 200
    Dan respons detail review memuat kode "SCN-0001" dan UUID semula
    Dan penyimpanan review dilakukan 0 kali

  Garis-Besar Skenario: Koreksi valid dapat disimpan termasuk saat coverage rendah
    Dengan pengguna review login sebagai "pemilik"
    Dan validator menyatakan draft review valid dengan coverage <coverage> persen
    Ketika pengguna menjalankan operasi review "simpan"
    Maka status respons review harus 200
    Dan draft review tersimpan persis sebagai CORRECTED dan VALID
    Dan coverage review tersimpan sebesar <coverage> persen
    Dan respons detail review memuat kode "SCN-0001" dan UUID semula
    Dan penyimpanan review dilakukan 1 kali
    Ketika pengguna menjalankan operasi review "detail"
    Maka status respons review harus 200
    Dan respons detail review berisi koreksi yang sudah disimpan

    Contoh:
      | coverage |
      | 100      |
      | 75       |
      | 0        |

  Skenario: Koreksi invalid mengembalikan detail kesalahan tanpa mengganti versi tersimpan
    Dengan pengguna review login sebagai "pemilik"
    Dan validator menolak draft review berikut karena Given kosong:
      """
      Feature: User Login
        Scenario: Successful login
          Given
          When the user clicks the Login button
          Then the Dashboard heading is displayed
      """
    Ketika pengguna menjalankan operasi review "simpan"
    Maka status respons review harus 422
    Dan respons review menjelaskan kesalahan deskripsi Given pada baris 3
    Dan data review tersimpan tidak berubah
    Dan penyimpanan review dilakukan 0 kali
    Ketika pengguna menjalankan operasi review "download"
    Maka status respons review harus 200
    Dan file feature berisi versi tersimpan secara persis

  Skenario: Download memuat koreksi valid terakhir
    Dengan pengguna review login sebagai "pemilik"
    Dan validator menyatakan draft review valid dengan coverage 75 persen
    Ketika pengguna menjalankan operasi review "simpan"
    Maka status respons review harus 200
    Ketika pengguna menjalankan operasi review "download"
    Maka status respons review harus 200
    Dan file feature berisi versi tersimpan secara persis
    Dan draft review tersimpan persis sebagai CORRECTED dan VALID
    Dan penyimpanan review dilakukan 1 kali

  Garis-Besar Skenario: Endpoint review memerlukan login
    Dengan pengguna review "<sesi>"
    Ketika pengguna menjalankan operasi review "<operasi>"
    Maka status respons review harus 401
    Dan validator review dipanggil 0 kali
    Dan data review tersimpan tidak berubah
    Dan penyimpanan review dilakukan 0 kali

    Contoh:
      | sesi                 | operasi  |
      | tanpa login          | daftar   |
      | tanpa login          | detail   |
      | tanpa login          | simpan   |
      | tanpa login          | download |
      | memakai token rusak  | detail   |

  Garis-Besar Skenario: QA tidak dapat menggunakan endpoint review
    Dengan pengguna review login sebagai "QA"
    Ketika pengguna menjalankan operasi review "<operasi>"
    Maka status respons review harus 403
    Dan validator review dipanggil 0 kali
    Dan data review tersimpan tidak berubah
    Dan penyimpanan review dilakukan 0 kali

    Contoh:
      | operasi  |
      | daftar   |
      | detail   |
      | simpan   |
      | download |

  Skenario: Hak akses mengikuti role pengguna tersimpan
    Dengan pengguna review login sebagai "QA"
    Dan token review mengklaim role System Analyst
    Ketika pengguna menjalankan operasi review "detail"
    Maka status respons review harus 403
    Dan validator review dipanggil 0 kali

  Garis-Besar Skenario: Skenario pengguna lain tidak dapat diakses
    Dengan pengguna review login sebagai "analis lain"
    Ketika pengguna menjalankan operasi review "<operasi>"
    Maka status respons review harus 404
    Dan respons review tidak membocorkan isi skenario
    Dan validator review dipanggil 0 kali
    Dan data review tersimpan tidak berubah
    Dan penyimpanan review dilakukan 0 kali

    Contoh:
      | operasi  |
      | detail   |
      | simpan   |
      | download |

  Skenario: Pengguna tanpa skenario menerima daftar kosong
    Dengan pengguna review login sebagai "analis tanpa skenario"
    Ketika pengguna menjalankan operasi review "daftar"
    Maka status respons review harus 200
    Dan daftar review kosong

  Skenario: Versi lama tidak dapat menimpa koreksi terbaru
    Dengan pengguna review login sebagai "pemilik"
    Dan validator menyatakan draft review valid dengan coverage 100 persen
    Dan pengguna membawa versi review yang sudah kedaluwarsa
    Ketika pengguna menjalankan operasi review "simpan"
    Maka status respons review harus 409
    Dan validator review dipanggil 0 kali
    Dan data review tersimpan tidak berubah
    Dan penyimpanan review dilakukan 0 kali

  Garis-Besar Skenario: Kegagalan validator tidak merusak versi tersimpan
    Dengan pengguna review login sebagai "pemilik"
    Dan validator menyatakan draft review valid dengan coverage 100 persen
    Dan layanan validator review mengalami "<gangguan>"
    Ketika pengguna menjalankan operasi review "simpan"
    Maka status respons review harus <status>
    Dan data review tersimpan tidak berubah
    Dan penyimpanan review dilakukan 0 kali

    Contoh:
      | gangguan        | status |
      | tidak tersedia  | 503    |
      | timeout         | 504    |
      | respons rusak   | 502    |

  Skenario: Versi tersimpan invalid tidak boleh diunduh
    Dengan pengguna review login sebagai "pemilik"
    Dan versi review tersimpan berstatus INVALID
    Ketika pengguna menjalankan operasi review "download"
    Maka status respons review harus 422
    Dan data review tersimpan tidak berubah
    Dan penyimpanan review dilakukan 0 kali

  Skenario: Payload tanpa versi ditolak sebelum validasi
    Dengan pengguna review login sebagai "pemilik"
    Ketika pengguna mengirim koreksi review tanpa updated_at
    Maka status respons review harus 400
    Dan validator review dipanggil 0 kali
    Dan data review tersimpan tidak berubah
    Dan penyimpanan review dilakukan 0 kali
