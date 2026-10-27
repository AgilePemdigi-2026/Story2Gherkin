# language: id
Fitur: Generate skenario Gherkin
  Sebagai system analyst
  Saya ingin mengubah user story menjadi skenario Gherkin
  Agar tidak menulis Given-When-Then manual

  Skenario: Generate sukses
    Dengan generator mengembalikan skenario valid
    Ketika saya generate skenario lengkap
    Maka status respons harus 200
    Dan generator dipanggil 1 kali

  Skenario: Generate field kosong
    Ketika saya generate tanpa user story
    Maka status respons harus 400
    Dan generator tidak dipanggil

  Skenario: Generator sibuk
    Dengan generator sedang sibuk
    Ketika saya generate skenario lengkap
    Maka status respons harus 503
