Semua endpoint berada di bawah `/api/v1` dan membutuhkan header
`Authorization: Bearer <access_token>`. Terapkan middleware autentikasi
yang sudah ada dan verifikasi role QA berdasarkan data pengguna terbaru
di database.

Seluruh akun QA dapat melihat daftar, membuka detail, dan mengunduh
report dari seluruh run, tanpa pembatasan berdasarkan pemilik run.
Unduhan hanya tersedia untuk run berstatus `COMPLETED`.
Pengguna dengan role selain QA tidak memiliki akses ke endpoint Report.

`page` dimulai dari 1 dengan nilai default 1 dan nilai maksimum 1.000.000.
Default `page_size` adalah 10, dengan rentang yang diperbolehkan 1–100.
Parameter yang bukan bilangan bulat atau berada di luar rentang tersebut
menghasilkan HTTP 400.

Urutkan hasil berdasarkan `created_at DESC`, kemudian `run_id DESC`
sebagai pemecah seri. Daftar mencakup seluruh run yang dapat diakses QA,
tanpa filter pemilik.

Halaman kosong, termasuk halaman yang melampaui jumlah hasil,
mengembalikan HTTP 200 dengan `runs: []`. Jika belum ada run,
`total_pages` bernilai 0. Jumlah halaman dihitung dari total run dan
ukuran halaman yang diminta.

Jika `metrics` tersedia, properti `llm` dan `rule_based` masing-masing
berupa objek ringkasan. Setiap nilai `validation_percent` dan
`gui_coverage_percent` di dalamnya boleh `null`.

Pada setiap kasus, properti `llm` dan `rule_based` boleh `null` apabila
metode tersebut belum menghasilkan hasil. Jika hasil metode tersedia,
`status` harus bernilai `PASS`, `WARN`, atau `FAIL`, sedangkan
`gui_coverage_percent` tetap boleh `null`.

Field nullable tetap dikirim dalam JSON dan tidak dihilangkan.
Gunakan `cases: []` ketika hasil kasus belum tersedia.

Gunakan penulis CSV standar supaya koma, tanda kutip, dan Gherkin
multiline di-escape dengan benar. Satu record CSV mewakili satu kasus;
teks Gherkin multiline tetap menjadi satu sel.

Gunakan sel kosong untuk nilai `null`. Nilai coverage 0 ditulis sebagai
`0`, tanpa tanda persen. Kasus berstatus `WARN` atau `FAIL` tetap
disertakan beserta Golden Gherkin yang tersimpan.

Implementasi ekspor memberi awalan tab pada teks yang berpotensi
dianggap formula, termasuk ketika karakter pemicu didahului whitespace
atau karakter kontrol. Perlakuan ini hanya mengubah sel hasil ekspor,
tanpa mengubah data report tersimpan atau respons JSON.

CSV dibentuk sepenuhnya sebelum respons sukses dikirim. Jika pembuatan
CSV gagal, kembalikan HTTP 500 dalam envelope JSON tanpa header lampiran.