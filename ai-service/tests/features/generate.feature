# language: id
Fitur: Generate skenario Gherkin di AI-service
  Sebagai backend Story2Gherkin
  Saya ingin memanggil AI-service untuk mengubah user story menjadi skenario valid
  Agar hasil siap disimpan tanpa koreksi manual

  Skenario: Generate sukses di percobaan pertama
    Dengan model mengembalikan fitur valid
    Ketika saya generate skenario
    Maka hasilnya valid dengan nama fitur "Hapus User"
    Dan model tercatat dipakai

  Skenario: Output invalid diperbaiki lewat feedback lalu sukses
    Dengan model mengembalikan nama skenario duplikat lalu fitur valid
    Ketika saya generate skenario
    Maka hasilnya valid
    Dan model dipanggil 2 kali
    Dan feedback menyebut nama duplikat

  Skenario: Output selalu invalid hingga batas percobaan
    Dengan model selalu mengembalikan nama skenario duplikat
    Ketika saya generate skenario
    Maka generate gagal
    Dan model dipanggil 2 kali

  Skenario: Model tidak tersedia tidak di-retry
    Dengan model sedang down
    Ketika saya generate skenario
    Maka generate gagal karena model down
    Dan model dipanggil 1 kali

  Skenario: Referensi elemen GUI tak dikenal diperbaiki lalu sukses
    Dengan model mengembalikan gui_element_ids tak dikenal lalu fitur valid
    Ketika saya generate skenario
    Maka hasilnya valid
    Dan feedback menyebut referensi tak dikenal

  Skenario: Respons terpotong lalu dipersingkat dan sukses
    Dengan model mengembalikan respons terpotong lalu fitur valid
    Ketika saya generate skenario
    Maka hasilnya valid
    Dan feedback menyebut respons terpotong
