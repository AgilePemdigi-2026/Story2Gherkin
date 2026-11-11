
# language: id

Fitur: Riwayat Generate Scenario
  Sebagai System Analyst
  Saya ingin melihat daftar dan detail riwayat Generate
  Agar dapat meninjau input dan hasil Generate sebelumnya
  tanpa bergantung pada AI Service

  Skenario: Daftar History hanya menampilkan Generate SUCCESS milik pengguna
    Dengan tersedia History SUCCESS dan FAILED milik SA
    Dan tersedia History milik SA lain
    Dan pengguna History login sebagai "pemilik"
    Ketika pengguna meminta daftar History
    Maka status respons History harus 200
    Dan daftar History hanya berisi 1 riwayat milik pengguna
    Dan daftar History tidak memuat riwayat pengguna lain
    Dan daftar History tidak memuat request FAILED

  Skenario: System Analyst tanpa History menerima daftar kosong
    Dengan tidak ada History tersimpan
    Dan pengguna History login sebagai "pemilik"
    Ketika pengguna meminta daftar History
    Maka status respons History harus 200
    Dan daftar History kosong

  Skenario: Pagination History menampilkan jumlah halaman yang benar
    Dengan tersedia 2 History milik SA
    Dan pengguna History login sebagai "pemilik"
    Ketika pengguna meminta daftar History halaman 1 dengan ukuran 1
    Maka status respons History harus 200
    Dan pagination History memiliki total 2 dan 2 halaman
    Dan daftar History berisi 1 item

  Skenario: Halaman History di luar jumlah data menghasilkan daftar kosong
    Dengan tersedia 2 History milik SA
    Dan pengguna History login sebagai "pemilik"
    Ketika pengguna meminta daftar History halaman 3 dengan ukuran 1
    Maka status respons History harus 200
    Dan daftar History kosong

  Garis-Besar Skenario: Parameter pagination History tidak valid
    Dengan pengguna History login sebagai "pemilik"
    Ketika pengguna meminta daftar History dengan parameter "<parameter>"
    Maka status respons History harus 400

    Contoh:
      | parameter             |
      | page=0                |
      | page_size=0           |
      | page_size=101         |
      | page=abc              |

  Skenario: Detail History SUCCESS menampilkan input dan Generated Gherkin
    Dengan tersedia History SUCCESS milik SA
    Dan pengguna History login sebagai "pemilik"
    Ketika pengguna meminta detail History SUCCESS
    Maka status respons History harus 200
    Dan detail History memuat User Story asli
    Dan detail History memuat GUI Elements asli
    Dan detail History memuat Free Scenario asli
    Dan detail History memuat Generated Gherkin
    Dan status History adalah "SUCCESS"

  Skenario: Detail History FAILED tidak dapat diakses
    Dengan tersedia History FAILED milik SA tanpa skenario
    Dan pengguna History login sebagai "pemilik"
    Ketika pengguna meminta detail History FAILED
    Maka status respons History harus 404
    Dan respons History tidak membocorkan data FAILED

  Skenario: System Analyst tidak dapat melihat History pengguna lain
    Dengan tersedia History milik SA lain
    Dan pengguna History login sebagai "pemilik"
    Ketika pengguna meminta detail History milik SA lain
    Maka status respons History harus 404
    Dan respons History tidak membocorkan data pengguna lain

  Skenario: History yang tidak ditemukan menghasilkan 404
    Dengan pengguna History login sebagai "pemilik"
    Ketika pengguna meminta detail History yang tidak tersedia
    Maka status respons History harus 404

  Skenario: UUID History yang tidak valid ditolak
    Dengan pengguna History login sebagai "pemilik"
    Ketika pengguna meminta detail History dengan UUID tidak valid
    Maka status respons History harus 404

  Garis-Besar Skenario: Endpoint History memerlukan login
    Dengan pengguna History tidak login
    Ketika pengguna menjalankan operasi History "<operasi>"
    Maka status respons History harus 401

    Contoh:
      | operasi |
      | daftar  |
      | detail  |

  Skenario: Token rusak tidak dapat mengakses History
    Dengan pengguna History menggunakan token rusak
    Ketika pengguna meminta daftar History
    Maka status respons History harus 401

  Garis-Besar Skenario: QA tidak dapat mengakses History
    Dengan pengguna History login sebagai "QA"
    Ketika pengguna menjalankan operasi History "<operasi>"
    Maka status respons History harus 403

    Contoh:
      | operasi |
      | daftar  |
      | detail  |

  Skenario: Otorisasi mengikuti role pengguna yang tersimpan
    Dengan pengguna History login sebagai "QA"
    Dan JWT History mengklaim role System Analyst
    Ketika pengguna meminta daftar History
    Maka status respons History harus 403
