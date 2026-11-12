DELETE FROM golden_dataset_items WHERE golden_dataset_id IN (
  'a1000000-0000-4000-8000-000000000001',
  'a1000000-0000-4000-8000-000000000002',
  'a1000000-0000-4000-8000-000000000003'
);

DELETE FROM golden_datasets WHERE id IN (
  'a1000000-0000-4000-8000-000000000001',
  'a1000000-0000-4000-8000-000000000002',
  'a1000000-0000-4000-8000-000000000003'
);

INSERT INTO golden_datasets (id, name, description, created_by, item_count, used_at) VALUES
  ('a1000000-0000-4000-8000-000000000001', 'Golden Dataset - Login', 'Kasus uji autentikasi dasar', (SELECT coalesce((SELECT id::text FROM users WHERE role = 'qa' ORDER BY created_at LIMIT 1), 'seed')), 3, NULL),
  ('a1000000-0000-4000-8000-000000000002', 'Golden Dataset - Checkout', 'Kasus uji alur pembelian', (SELECT coalesce((SELECT id::text FROM users WHERE role = 'qa' ORDER BY created_at LIMIT 1), 'seed')), 2, NULL),
  ('a1000000-0000-4000-8000-000000000003', 'Golden Dataset - Arsip', 'Sudah dipakai evaluasi, terkunci', (SELECT coalesce((SELECT id::text FROM users WHERE role = 'qa' ORDER BY created_at LIMIT 1), 'seed')), 1, now() - interval '7 days');

INSERT INTO golden_dataset_items (golden_dataset_id, item_title, user_story_text, gui_elements, free_scenario_text, expected_gherkin) VALUES
  ('a1000000-0000-4000-8000-000000000001', 'Login berhasil',
   'Sebagai pengguna, saya ingin masuk dengan email dan password agar dapat mengakses dashboard.',
   '["Kolom Email", "Kolom Password", "Tombol Login"]',
   'Pengguna mengisi email dan password yang valid lalu menekan tombol Login.',
   'Feature: Login
  Scenario: Login berhasil dengan kredensial valid
    Given pengguna berada di halaman login
    When pengguna mengisi kolom Email dengan "user@example.com"
    And pengguna mengisi kolom Password dengan "rahasia123"
    And pengguna menekan tombol Login
    Then pengguna diarahkan ke halaman dashboard'),

  ('a1000000-0000-4000-8000-000000000001', 'Login password salah',
   'Sebagai pengguna, saya ingin mendapat pesan kesalahan saat password salah agar tahu penyebab gagal masuk.',
   '["Kolom Email", "Kolom Password", "Tombol Login", "Pesan Error"]',
   'Pengguna mengisi password salah lalu menekan tombol Login.',
   'Feature: Login
  Scenario: Login gagal karena password salah
    Given pengguna berada di halaman login
    When pengguna mengisi kolom Email dengan "user@example.com"
    And pengguna mengisi kolom Password dengan "salah"
    And pengguna menekan tombol Login
    Then pesan Error menampilkan "Email atau password salah"
    And pengguna tetap berada di halaman login'),

  ('a1000000-0000-4000-8000-000000000001', 'Login kolom kosong',
   'Sebagai pengguna, saya ingin diminta mengisi kolom yang kosong agar tidak mengirim form tidak lengkap.',
   '["Kolom Email", "Kolom Password", "Tombol Login", "Pesan Error"]',
   'Pengguna menekan Login tanpa mengisi email dan password.',
   'Feature: Login
  Scenario: Login ditolak saat kolom kosong
    Given pengguna berada di halaman login
    When pengguna mengosongkan kolom Email
    And pengguna mengosongkan kolom Password
    And pengguna menekan tombol Login
    Then pesan Error menampilkan "Email wajib diisi"
    And permintaan login tidak dikirim'),

  ('a1000000-0000-4000-8000-000000000002', 'Checkout berhasil',
   'Sebagai pembeli, saya ingin menyelesaikan checkout agar pesanan diproses.',
   '["Keranjang", "Tombol Checkout", "Alamat Pengiriman", "Metode Pembayaran", "Tombol Konfirmasi"]',
   'Pembeli membuka keranjang, mengisi alamat, memilih pembayaran, lalu konfirmasi.',
   'Feature: Checkout
  Scenario: Checkout pesanan berhasil
    Given pembeli memiliki 1 produk di keranjang
    When pembeli menekan tombol Checkout
    And pembeli mengisi Alamat Pengiriman dengan "Jl. Merdeka 1"
    And pembeli memilih Metode Pembayaran "Transfer Bank"
    And pembeli menekan tombol Konfirmasi
    Then pesanan tercatat dengan status "Menunggu Pembayaran"'),

  ('a1000000-0000-4000-8000-000000000002', 'Checkout keranjang kosong',
   'Sebagai pembeli, saya ingin dicegah checkout saat keranjang kosong agar tidak membuat pesanan kosong.',
   '["Keranjang", "Tombol Checkout", "Pesan Error"]',
   'Pembeli menekan Checkout dengan keranjang kosong.',
   'Feature: Checkout
  Scenario: Checkout ditolak saat keranjang kosong
    Given keranjang pembeli kosong
    When pembeli menekan tombol Checkout
    Then pesan Error menampilkan "Keranjang masih kosong"
    And tidak ada pesanan yang dibuat'),

  ('a1000000-0000-4000-8000-000000000003', 'Logout',
   'Sebagai pengguna, saya ingin keluar dari akun agar sesi berakhir dengan aman.',
   '["Menu Profil", "Tombol Logout"]',
   'Pengguna membuka menu profil lalu menekan tombol Logout.',
   'Feature: Logout
  Scenario: Pengguna berhasil logout
    Given pengguna sudah login
    When pengguna membuka menu Profil
    And pengguna menekan tombol Logout
    Then sesi pengguna berakhir
    And pengguna diarahkan ke halaman login');
