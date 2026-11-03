import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import DocumentationCode from '../components/documentation/DocumentationCode'

export interface DocumentationArticle {
  slug: string
  title: string
  group: string
  description: string
  keywords: string
  sections: { id: string; title: string; content: ReactNode }[]
}

const loginExample = `Feature: User login

  Scenario: Successful login
    Given the Email field and Password field are filled
    When the user clicks the Login button
    Then the Dashboard heading is displayed`


export const documentationArticles: DocumentationArticle[] = [
  {
    slug: 'introduction', title: 'Pengenalan', group: 'overview',
    description: 'Pelajari Story2Gherkin, alur kerja pengguna, dan dasar skenario Gherkin.', keywords: 'introduction overview tentang bdd',
    sections: [
      { id: 'tentang', title: 'Apa itu Story2Gherkin?', content: <><p>Story2Gherkin membantu System Analyst mengubah kebutuhan pengguna menjadi skenario Gherkin, meninjau hasilnya, dan menyimpan koreksi. Ruang kerja QA disiapkan untuk evaluasi kualitas skenario.</p><p>Dokumentasi ini dapat dibaca tanpa login. Mulailah dari panduan penggunaan atau pelajari dasar Gherkin sebelum membuat skenario pertama.</p></> },
      { id: 'alur-kerja', title: 'Bagaimana alur kerjanya?', content: <ol><li><strong>Masukkan kebutuhan.</strong> Isi User Story, GUI Elements, dan Free Scenario.</li><li><strong>Generate dan review.</strong> Periksa hasil Gherkin, struktur langkah, dan cakupan elemen GUI.</li><li><strong>Simpan dan unduh.</strong> Simpan koreksi yang valid, lalu download file <code>.feature</code>.</li></ol> },
      { id: 'mengenal-gherkin', title: 'Apa itu Gherkin?', content: <><p>Gherkin memberi struktur pada contoh perilaku menggunakan teks yang mudah dibaca. File <code>.feature</code> mengelompokkan skenario dan langkah-langkahnya.</p><p>Untuk menjalankannya sebagai pengujian otomatis, langkah Gherkin perlu dihubungkan ke kode melalui <em>step definitions</em> dan dijalankan dengan alat pengujian. Story2Gherkin saat ini membantu menyusun dan memvalidasi teks skenario.</p></> },
      { id: 'contoh-skenario', title: 'Contoh skenario', content: <><p>Contoh berikut menggambarkan login berhasil. Nama elemen GUI ditulis konsisten agar mudah diperiksa pada saat review.</p><DocumentationCode code={loginExample} filename="login.feature" /></> },
      { id: 'ketersediaan', title: 'Fitur yang tersedia', content: <><p>Generate Scenario dan Review Scenario sudah dapat digunakan oleh System Analyst.</p><aside className="docs-note"><strong>Evaluation dalam pengembangan</strong><p>Golden Dataset, Run Evaluation, dan Evaluation Report belum tersedia pada versi ini. Panduan QA menjelaskan status ruang kerjanya.</p></aside></> },
    ],
  },
  {
    slug: 'getting-started', title: 'Mulai Menggunakan', group: 'guides',
    description: 'Buat akun, login, dan temukan ruang kerja sesuai peran Anda.', keywords: 'login sign up register akun role sa qa mulai',
    sections: [
      { id: 'mulai', title: 'Sebelum login', content: <p>Home, About, dan Documentation terbuka untuk semua pengunjung. Tombol <strong>Mulai Menggunakan</strong> membawa Anda ke halaman Login, tanpa menentukan fitur yang akan digunakan terlebih dahulu.</p> },
      { id: 'membuat-akun', title: 'Membuat akun', content: <ol><li>Buka <Link to="/register">Sign Up</Link>.</li><li>Isi nama, email, dan password.</li><li>Pilih peran System Analyst atau Quality Assurance sesuai tugas Anda.</li><li>Selesaikan pendaftaran, lalu <Link to="/login">login</Link> menggunakan akun tersebut.</li></ol> },
      { id: 'memilih-ruang-kerja', title: 'Ruang kerja sesuai peran', content: <div className="docs-table-wrap"><table><thead><tr><th>Peran</th><th>Halaman setelah login</th><th>Kegunaan</th></tr></thead><tbody><tr><td>System Analyst</td><td>Generate</td><td>Membuat dan meninjau skenario.</td></tr><tr><td>Quality Assurance</td><td>Evaluation</td><td>Ruang kerja evaluasi; fitur inti masih dalam pengembangan.</td></tr></tbody></table></div> },
      { id: 'sesi-login', title: 'Sesi dan logout', content: <><p>Gunakan Remember me jika ingin mempertahankan sesi sesuai masa berlaku login. Jika sesi berakhir, login kembali untuk mengakses fitur.</p><p>Logout mengembalikan Anda ke Home. Jika ada draft review yang belum disimpan, aplikasi meminta konfirmasi sebelum meninggalkan perubahan.</p></> },
    ],
  },
  {
    slug: 'generate', title: 'Generate Scenario', group: 'guides',
    description: 'Ubah user story dan informasi GUI menjadi draft skenario Gherkin.', keywords: 'system analyst sa user story gui elements free scenario generate',
    sections: [
      { id: 'system-analyst', title: 'Untuk System Analyst', content: <p>Login sebagai System Analyst dan buka menu Generate. Siapkan satu kebutuhan yang jelas beserta alur pengguna yang ingin diuji.</p> },
      { id: 'input-kebutuhan', title: 'Isi kebutuhan skenario', content: <><dl className="docs-definitions"><dt>User Story</dt><dd>Jelaskan siapa pengguna, tujuan yang ingin dicapai, dan alasannya.</dd><dt>GUI Elements</dt><dd>Masukkan satu nama elemen per kolom. Gunakan Tambah Elemen untuk menambah kolom.</dd><dt>Free Scenario</dt><dd>Uraikan kondisi awal, tindakan pengguna, dan hasil yang diharapkan.</dd></dl><aside className="docs-note"><strong>Contoh input login</strong><p>User Story: As a user, I want to log in so that I can access my dashboard.</p><p>GUI Elements: Email field, Password field, Login button, Dashboard heading.</p><p>Free Scenario: The user fills the email and password, clicks Login button, then sees Dashboard heading.</p></aside></> },
      { id: 'menjalankan-generate', title: 'Jalankan generate', content: <ol><li>Pastikan semua kolom wajib terisi.</li><li>Klik <strong>Generate Skenario</strong> dan tunggu proses selesai.</li><li>Baca hasilnya dan periksa apakah sesuai kebutuhan.</li><li>Jika hasil sudah tersimpan, klik <strong>Review Skenario</strong> untuk meninjaunya lebih lanjut.</li></ol> },
      { id: 'hasil-generate', title: 'Periksa hasil', content: <><p>Hasil AI perlu ditinjau kembali. Periksa nama fitur, urutan langkah, hasil yang diharapkan, dan konsistensi nama elemen GUI.</p><DocumentationCode code={loginExample} filename="login.feature" /></> },
    ],
  },
  {
    slug: 'review', title: 'Review Scenario', group: 'guides',
    description: 'Periksa, koreksi, simpan, dan unduh skenario milik Anda.', keywords: 'edit koreksi simpan download feature draft system analyst',
    sections: [
      { id: 'memilih-skenario', title: 'Pilih skenario', content: <p>Buka Review Scenario dari sidebar Generate, lalu pilih skenario pada dropdown. Anda juga dapat membuka hasil tertentu langsung dari tombol Review Skenario setelah generate.</p> },
      { id: 'mengedit', title: 'Edit skenario', content: <p>Klik <strong>Edit Skenario</strong> untuk mengubah teks. Pastikan Feature dan Scenario memiliki nama, serta setiap Given, When, dan Then menjelaskan langkah yang dimaksud.</p> },
      { id: 'menyimpan', title: 'Simpan perubahan', content: <><p>Klik <strong>Simpan Perubahan</strong> untuk menjalankan validasi. Koreksi yang valid akan menjadi versi tersimpan. Jika invalid, perbaiki kesalahan sesuai nomor baris; draft tetap tersedia dan versi tersimpan tidak berubah.</p><p>Coverage di bawah 100% merupakan peringatan kelengkapan. Koreksi tetap dapat disimpan jika struktur skenario valid.</p></> },
      { id: 'mengunduh', title: 'Download .feature', content: <p>Tombol Download .feature selalu mengunduh versi valid terakhir yang tersimpan. Jika Anda sedang mengedit, simpan koreksi terlebih dahulu agar perubahan tersebut ikut diunduh.</p> },
      { id: 'draft-dan-konflik', title: 'Draft dan perubahan di tab lain', content: <><p>Perpindahan halaman, pergantian skenario, dan logout meminta konfirmasi saat ada perubahan yang belum disimpan.</p><p>Jika skenario telah diperbarui dari tab lain, salin draft yang ingin dipertahankan, lalu muat ulang versi terbaru sebelum menggabungkan koreksi.</p></> },
    ],
  },
  {
    slug: 'evaluation', title: 'Evaluation untuk QA', group: 'guides',
    description: 'Kenali ruang kerja Quality Assurance dan status fitur evaluasi.', keywords: 'qa quality assurance golden dataset run evaluation report',
    sections: [
      { id: 'qa', title: 'Ruang kerja Quality Assurance', content: <p>Setelah login sebagai QA, Anda diarahkan ke Evaluation. Menu ini disiapkan untuk pekerjaan evaluasi kualitas skenario.</p> },
      { id: 'status-evaluasi', title: 'Status ketersediaan', content: <aside className="docs-note"><strong>Belum tersedia pada versi ini</strong><p>Golden Dataset, Run Evaluation, dan Evaluation Report masih dalam pengembangan. Halaman awal Evaluation menampilkan status ini dan belum menjalankan perhitungan evaluasi.</p></aside> },
      { id: 'persiapan-qa', title: 'Yang dapat dipelajari sekarang', content: <p>Pelajari struktur Gherkin, baca alur Generate dan Review, lalu samakan pemahaman tim tentang kebutuhan, skenario yang diharapkan, dan nama elemen GUI sebelum fitur evaluasi digunakan.</p> },
    ],
  },
  {
    slug: 'gherkin', title: 'Dasar Gherkin', group: 'reference',
    description: 'Pahami Feature, Scenario, Given, When, Then, dan And.', keywords: 'gherkin syntax keyword feature scenario given when then and bdd',
    sections: [
      { id: 'struktur-file', title: 'Struktur file .feature', content: <><p><code>Feature:</code> menamai fitur dan mengelompokkan skenario. <code>Scenario:</code> menggambarkan satu contoh perilaku yang konkret. Keduanya diikuti titik dua dan judul.</p><DocumentationCode code={loginExample} filename="login.feature" /></> },
      { id: 'keyword-langkah', title: 'Keyword langkah', content: <div className="docs-table-wrap"><table><thead><tr><th>Keyword</th><th>Makna</th></tr></thead><tbody><tr><td><code>Given</code></td><td>Kondisi awal sebelum tindakan.</td></tr><tr><td><code>When</code></td><td>Tindakan atau kejadian yang diuji.</td></tr><tr><td><code>Then</code></td><td>Hasil yang dapat diamati.</td></tr><tr><td><code>And</code></td><td>Melanjutkan langkah sebelumnya.</td></tr></tbody></table></div> },
      { id: 'menulis-langkah', title: 'Menulis langkah yang jelas', content: <><p>Isi deskripsi setelah keyword. Tulis hasil yang dapat diamati pada Then. Gunakan indentasi yang konsisten, misalnya dua spasi per tingkat.</p><p>Contoh: <code>When the user clicks the Login button</code> menjelaskan tindakan, sedangkan <code>When</code> saja belum menjelaskan langkah.</p></> },
      { id: 'dukungan-story2gherkin', title: 'Dukungan Story2Gherkin', content: <aside className="docs-note"><strong>Aturan review pada aplikasi ini</strong><p>Review menggunakan keyword Inggris Feature, Scenario, Given, When, Then, dan And. Setiap Scenario harus memiliki Given, When, dan Then dengan deskripsi serta urutan yang sesuai.</p><p>Background, Rule, But, Scenario Outline, dan Examples belum didukung oleh validator review Story2Gherkin.</p></aside> },
    ],
  },
  {
    slug: 'validation', title: 'Validation & GUI Coverage', group: 'reference',
    description: 'Baca status struktur, pesan kesalahan, dan cakupan elemen GUI.', keywords: 'valid invalid warning error struktur kosong deskripsi validasi coverage covered',
    sections: [
      { id: 'validasi', title: 'Struktur Gherkin', content: <p>Panel menunjukkan pemeriksaan Feature, Scenario, Given, When, dan Then. Bagian yang sudah lengkap tetap ditandai hijau meskipun langkah lain salah. Status VALID menunjukkan teks lolos aturan review aplikasi, bukan bukti bahwa aplikasi yang diuji bekerja dengan benar.</p> },
      { id: 'kesalahan-langkah', title: 'Langkah kosong atau tidak sesuai', content: <><p>Pesan kesalahan menyebut nomor baris dan penyebabnya. Contohnya, <strong>Given belum memiliki deskripsi langkah</strong> meminta Anda menambahkan kondisi awal setelah keyword Given.</p><DocumentationCode code={'Feature: User login\n\n  Scenario: Successful login\n    Given\n    When\n    Then'} filename="contoh-invalid.feature" /><p>Lengkapi tiga langkah tersebut, lalu simpan kembali untuk memvalidasi koreksi.</p></> },
      { id: 'gui-coverage', title: 'GUI Coverage', content: <><p>Story2Gherkin menghitung persentase elemen GUI yang ditemukan pada teks langkah. Judul, komentar, dan isi tabel tidak dihitung sebagai penggunaan elemen.</p><p><code>Coverage = elemen GUI yang digunakan / total elemen GUI × 100%</code></p><p>Jika 3 dari 4 elemen disebut, coverage adalah 75%. USED berisi elemen yang ditemukan; NOT COVERED berisi yang belum ditemukan. Angka ini membantu memeriksa kelengkapan penyebutan, bukan keberhasilan pengujian otomatis.</p></> },
      { id: 'belum-dihitung', title: 'Belum dapat dihitung', content: <p>Jika struktur tidak dapat dibaca atau ada langkah tanpa deskripsi, perbaiki teks terlebih dahulu. Panel menampilkan Belum dapat dihitung agar kondisi ini tidak tertukar dengan coverage 0% yang benar-benar dihitung.</p> },
      { id: 'coverage-dan-simpan', title: 'Pengaruh terhadap penyimpanan', content: <p>Coverage di bawah 100% tidak menghalangi penyimpanan jika struktur valid. Struktur invalid tetap ditolak; draft dapat diperbaiki, sedangkan versi valid terakhir tetap tersimpan.</p> },
    ],
  },
  {
    slug: 'faq', title: 'Pertanyaan Umum', group: 'help',
    description: 'Jawaban singkat untuk pertanyaan saat menggunakan Story2Gherkin.', keywords: 'faq bantuan masalah login download error generate evaluasi',
    sections: [
      { id: 'akun-dan-akses', title: 'Akun dan akses', content: <div className="docs-faq"><details><summary>Apakah membuka web harus langsung login?</summary><p>Tidak. Home, About, dan Documentation terbuka untuk semua pengunjung. Login diperlukan saat mengakses ruang kerja.</p></details><details><summary>Mengapa menu Generate atau Evaluation tidak terlihat?</summary><p>Menu mengikuti role akun. System Analyst melihat Generate dan Review; QA melihat Evaluation. Jika sesi berakhir, login kembali.</p></details></div> },
      { id: 'hasil-dan-koreksi', title: 'Hasil dan koreksi', content: <div className="docs-faq"><details><summary>Mengapa perubahan gagal disimpan?</summary><p>Periksa pesan di Validation Result. Lengkapi deskripsi langkah dan perbaiki urutannya. Jika versi telah berubah di tab lain, pertahankan salinan draft sebelum memuat ulang.</p></details><details><summary>Mengapa download tidak memuat edit terbaru?</summary><p>Download menggunakan versi tersimpan. Klik Simpan Perubahan dan pastikan berhasil sebelum mengunduh lagi.</p></details><details><summary>Apakah hasil VALID berarti pengujian sudah dijalankan?</summary><p>Tidak. Validasi review memeriksa teks skenario. Pengujian otomatis membutuhkan step definitions dan alat pengujian yang dikonfigurasi untuk sistem yang diuji.</p></details></div> },
    ],
  },

]

export const documentationGroups = [
  { id: 'overview', label: 'Memulai' },
  { id: 'guides', label: 'Panduan Penggunaan' },
  { id: 'reference', label: 'Gherkin & Validasi' },
  { id: 'help', label: 'Bantuan' },
]

const legacyTopics: Record<string, string> = { '#mulai': 'getting-started', '#system-analyst': 'generate', '#qa': 'evaluation', '#validasi': 'validation' }

export function resolveDocumentationTopic(topic: string | null, hash: string) {
  const selected = topic || legacyTopics[hash] || 'introduction'
  return documentationArticles.find(article => article.slug === selected) || documentationArticles[0]
}
