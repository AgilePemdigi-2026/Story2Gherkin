# Migrasi AI service ke OpenRouter Nemotron

Provider AI berubah dari Ollama lokal menjadi OpenRouter.
Tidak perlu mengunduh model, menjalankan Ollama, melatih model, atau memiliki GPU lokal.
Komputer tetap memerlukan internet untuk request generate asli.

## Perubahan kode

| File | Perubahan |
| --- | --- |
| `app/config.py` | Environment OpenRouter, validasi key, model gratis, token budget, timeout |
| `app/llm_client.py` | `OpenRouterClient`, Bearer API key, endpoint chat completions, pembacaan `choices` |
| `app/main.py` | Wiring client baru dan penanganan rate limit/configuration error |
| `.env.example` | Contoh konfigurasi baru tanpa secret asli |
| `tests/` | Mock OpenRouter, konfigurasi baru, format JSON provider, BDD error tambahan |

Model dipilih lewat environment. Kode tidak menetapkan varian Nemotron tertentu karena ID pilihan
pengguna belum diberikan. Untuk mencegah pemakaian model berbayar tidak sengaja, konfigurasi
menerima ID NVIDIA Nemotron yang diakhiri `:free`. Tidak ada fallback model berbayar di request.
Validasi nama lokal tidak membuktikan ID itu masih tersedia pada OpenRouter.

## Isi environment lokal

Edit `ai-service/.env`. Pertahankan nilai internal `AI_SERVICE_KEY` yang sudah sama dengan backend:

```dotenv
OPENROUTER_API_KEY=ISI_API_KEY_OPENROUTER_KAMU
OPENROUTER_MODEL=ISI_ID_MODEL_NEMOTRON_YANG_KAMU_PILIH
OPENROUTER_RESPONSE_FORMAT=prompt
OPENROUTER_MAX_TOKENS=4096
LLM_TIMEOUT_SECONDS=120
AI_SERVICE_KEY=SECRET_INTERNAL_YANG_SAMA_DENGAN_BACKEND
```

`OPENROUTER_API_KEY` adalah key provider, sedangkan `AI_SERVICE_KEY` adalah secret komunikasi Go–Python.
Keduanya wajib berbeda. API key provider hanya dikirim dari Python ke endpoint HTTPS OpenRouter.
API key tidak dikirim ke frontend dan tidak diteruskan ke Go.

Contoh ID chat model yang diverifikasi pada halaman resmi saat revisi:
`nvidia/nemotron-3-super-120b-a12b:free`.
Gunakan ID persis dari halaman model yang kamu pilih; contoh ini bukan penggantian otomatis atas pilihanmu.
[Halaman model](https://openrouter.ai/nvidia/nemotron-3-super-120b-a12b:free).

Variable `OLLAMA_URL`, `OLLAMA_MODEL`, dan `OLLAMA_TIMEOUT_SECONDS` tidak digunakan oleh versi ini.
Jangan commit `.env`. File `.env.example` tetap berisi placeholder.

## Format output

Default `OPENROUTER_RESPONSE_FORMAT=prompt`:

- JSON Schema dimasukkan ke system prompt.
- Provider diminta mengembalikan JSON.
- JSON tetap diperiksa dengan Pydantic, validasi referensi, dan parser Gherkin.
- Model bisa gagal mengikuti format; kegagalan tersebut menjadi HTTP 502.

Jika model dan provider mendukung structured outputs, gunakan `json_schema`.
Mode tersebut mengirim `response_format` dengan JSON Schema dan meminta provider mendukung parameter itu.
Kode tidak melakukan fallback diam-diam jika konfigurasi tidak didukung.

Respons memakai `choices[0].message.content`. Output yang kosong, terpotong
(misalnya `finish_reason=length`), refused, atau tidak cocok schema ditolak.
Konten reasoning tidak digunakan sebagai hasil Gherkin.

## Jalankan

Dari folder `ai-service` dan environment Python aktif:

```powershell
python -m pip install -r requirements-dev.txt
python -m pytest -q
python -m uvicorn app.main:app --host 127.0.0.1 --port 8001
```

Stop lalu restart Uvicorn setelah mengubah kode atau environment.
Tidak ada dependency SDK provider baru; implementasi tetap menggunakan HTTPX.
Jangan menjalankan server dengan nilai placeholder pada environment.

## Error provider

| Kondisi OpenRouter | Respons AI service |
| --- | --- |
| Rate limit HTTP 429 | HTTP 503 dengan pesan rate limit |
| API key salah, model tidak ditemukan, parameter tidak didukung | HTTP 502 dengan pesan konfigurasi umum |
| Provider tidak tersedia/koneksi gagal | HTTP 503 |
| Timeout | HTTP 504 |
| Respons bukan JSON/schema salah/output terpotong | HTTP 502 |

Body error provider dan key tidak dikirim ke caller. Free endpoint tetap memiliki batas
dan ketersediaan; error tidak memicu pemilihan model berbayar.
Python melakukan satu panggilan provider per request. Retry yang sudah ada berada di Go.

## Pengujian

```powershell
python -m pytest tests/unit -v
python -m pytest tests/api -v
python -m pytest tests/bdd -v --gherkin-terminal-reporter
python -m pytest tests --cov=app --cov-branch --cov-report=term-missing --cov-report=html
```

Verifikasi revisi: **169 test PASS**, **100% statement dan branch coverage** pada snapshot revisi.
Test menggunakan mock HTTP, tidak memakai API key asli, kuota, maupun model sungguhan.
Ada satu deprecation warning dari Starlette TestClient/HTTPX yang tidak menggagalkan test.

## Kontrak Go–Python

Revisi ini hanya mengganti provider model. Python masih menerima `acceptance_criteria` dan
mengembalikan `criterion_ids`. Go yang diperiksa menggunakan `gui_elements`, `free_scenario`,
dan `gui_element_ids`. Kontrak tersebut perlu diselaraskan sebelum menyatakan integrasi generate selesai.

## Referensi

- [OpenRouter quickstart](https://openrouter.ai/docs/quickstart)
- [OpenRouter structured outputs](https://openrouter.ai/docs/guides/features/structured-outputs)
