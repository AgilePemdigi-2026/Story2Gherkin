# Story2Gherkin AI service

AI service menerima tiga input: `user_story`, `gui_elements` (daftar string), dan `free_scenario`. Backend Go memanggil AI service; AI service kemudian meminta Ollama menghasilkan satu feature Gherkin, mem-parsingnya, lalu mengembalikan hasil terstruktur.

## Jalankan dengan Ollama lokal

1. Jalankan Ollama dan unduh model yang dipilih di `OLLAMA_MODEL` (contoh dari `.env.example`):

   ```sh
   ollama serve
   ollama pull qwen2.5:3b
   ollama list
   ```

2. Siapkan konfigurasi dan secret bersama backend:

   ```sh
   cp .env.example .env
   openssl rand -hex 32
   ```

   Simpan nilai acak itu pada `AI_SERVICE_KEY` di `ai-service/.env` dan `backend/.env`. Pertahankan `LLM_PROVIDER=ollama`, lalu pastikan `OLLAMA_MODEL` sama dengan nama model hasil `ollama list`.

3. Instal dependency dan jalankan service dari direktori `ai-service`:

   ```sh
   python -m venv .venv
   . .venv/bin/activate
   pip install -r requirements.txt
   uvicorn app.main:app --host 0.0.0.0 --port 8001
   ```

4. Atur `AI_SERVICE_URL` di `backend/.env` ke alamat yang bisa dijangkau backend, biasanya `http://127.0.0.1:8001` bila keduanya berjalan langsung pada host yang sama. `OLLAMA_URL` harus bisa dijangkau dari AI service. Bila salah satunya berjalan di container, `127.0.0.1` menunjuk ke container itu sendiri; gunakan alamat host/service Ollama yang tersedia di jaringan container.

5. Jalankan backend dan frontend sesuai konfigurasi masing-masing. Endpoint liveness AI service adalah `GET /health`; endpoint generate internal memerlukan header `X-AI-Service-Key` dan dipanggil backend.

## Kontrak input generator

Contoh body internal yang dikirim backend:

```json
{
  "user_story": "As an admin, I want to delete an account so that obsolete accounts can be removed.",
  "gui_elements": ["Delete button", "Confirmation dialog"],
  "free_scenario": "The admin opens the account page, selects delete, and confirms the action."
}
```

`OLLAMA_NUM_CTX` mengatur context window yang diminta dan `OLLAMA_NUM_PREDICT` membatasi token keluaran. Keduanya perlu disesuaikan dengan model dan memori yang tersedia. Service meminta satu respons JSON dari endpoint Ollama `/api/chat` (`stream: false`) dan memvalidasi Gherkin sebelum mengembalikannya.

OpenRouter masih tersedia untuk lingkungan yang memilih `LLM_PROVIDER=openrouter`; konfigurasi normal sekarang memilih Ollama.
