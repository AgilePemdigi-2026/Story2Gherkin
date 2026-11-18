# language: id
Fitur: Menjalankan Evaluasi
  Sebagai Quality Assurance
  Saya ingin menjalankan evaluasi pada Golden Dataset
  Agar hasil LLM dan Rule-Based dapat dibandingkan lalu dibaca pada Evaluation Report

  Skenario: QA berhasil menjalankan evaluasi
    Dengan pengguna evaluasi login sebagai "QA"
    Dan terdapat Golden Dataset evaluasi berisi 2 item
    Dan evaluator menghasilkan Gherkin valid dengan cakupan 100 persen
    Ketika pengguna menjalankan evaluasi
    Maka status respons evaluasi harus 201
    Dan respons evaluasi berisi 2 kasus
    Dan evaluasi berstatus "COMPLETED" dengan 2 kasus diproses
    Dan metrik evaluasi menyimpan dua metode

  Skenario: System Analyst tidak dapat menjalankan evaluasi
    Dengan pengguna evaluasi login sebagai "SA"
    Dan terdapat Golden Dataset evaluasi berisi 2 item
    Ketika pengguna menjalankan evaluasi
    Maka status respons evaluasi harus 403
    Dan tidak ada evaluasi yang tersimpan

  Skenario: Run Evaluation memerlukan autentikasi
    Dengan pengguna evaluasi tidak login
    Dan terdapat Golden Dataset evaluasi berisi 2 item
    Ketika pengguna menjalankan evaluasi
    Maka status respons evaluasi harus 401
    Dan tidak ada evaluasi yang tersimpan

  Skenario: Evaluasi ditolak ketika masih ada run aktif
    Dengan pengguna evaluasi login sebagai "QA"
    Dan terdapat Golden Dataset evaluasi berisi 2 item
    Dan terdapat evaluasi aktif yang tersimpan
    Ketika pengguna menjalankan evaluasi
    Maka status respons evaluasi harus 409
    Dan tidak ada evaluasi baru yang tersimpan

  Skenario: Golden Dataset tanpa item ditolak
    Dengan pengguna evaluasi login sebagai "QA"
    Dan terdapat Golden Dataset evaluasi tanpa item
    Ketika pengguna menjalankan evaluasi
    Maka status respons evaluasi harus 400
    Dan tidak ada evaluasi yang tersimpan

  Skenario: Golden Dataset yang tidak tersedia ditolak
    Dengan pengguna evaluasi login sebagai "QA"
    Ketika pengguna menjalankan evaluasi pada dataset yang tidak tersedia
    Maka status respons evaluasi harus 404
    Dan tidak ada evaluasi yang tersimpan

  Skenario: Cakupan di bawah 100 persen menghasilkan WARN
    Dengan pengguna evaluasi login sebagai "QA"
    Dan terdapat Golden Dataset evaluasi berisi 1 item
    Dan evaluator menghasilkan Gherkin valid dengan cakupan 50 persen
    Ketika pengguna menjalankan evaluasi
    Maka evaluasi berstatus "COMPLETED" dengan 1 kasus diproses
    Dan kasus evaluasi berstatus "WARN" untuk metode "rule_based"

  Skenario: Gherkin tidak valid menghasilkan FAIL
    Dengan pengguna evaluasi login sebagai "QA"
    Dan terdapat Golden Dataset evaluasi berisi 1 item
    Dan evaluator menghasilkan Gherkin tidak valid
    Ketika pengguna menjalankan evaluasi
    Maka evaluasi berstatus "COMPLETED" dengan 1 kasus diproses
    Dan kasus evaluasi berstatus "FAIL" untuk metode "llm"

  Skenario: Progres evaluasi dapat dipantau
    Dengan pengguna evaluasi login sebagai "QA"
    Dan terdapat evaluasi yang sedang berjalan dengan 3 kasus dan 1 kasus diproses
    Ketika pengguna meminta progress evaluasi
    Maka status respons evaluasi harus 200
    Dan progress evaluasi menunjukkan 1 dari 3 kasus

  Skenario: Progres evaluasi yang tidak tersedia
    Dengan pengguna evaluasi login sebagai "QA"
    Ketika pengguna meminta progress evaluasi yang tidak tersedia
    Maka status respons evaluasi harus 404
