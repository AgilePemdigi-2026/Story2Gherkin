import os
from dataclasses import dataclass
from urllib.parse import urlparse

from dotenv import load_dotenv


@dataclass(frozen=True)
class Settings:
    ollama_url: str
    model: str
    service_key: str
    timeout_seconds: float = 120

    @classmethod
    def from_env(cls) -> "Settings":
        load_dotenv()
        settings = cls(
            ollama_url=os.getenv("OLLAMA_URL", "http://127.0.0.1:11434").rstrip("/"),
            model=os.getenv("OLLAMA_MODEL", "qwen3.5:4b").strip(),
            service_key=os.getenv("AI_SERVICE_KEY", ""),
            timeout_seconds=float(os.getenv("OLLAMA_TIMEOUT_SECONDS", "120")),
        )
        if len(settings.service_key.strip()) < 32 or settings.service_key.startswith("REPLACE_"):
            raise ValueError("Set AI_SERVICE_KEY to a random secret of at least 32 characters")
        url = urlparse(settings.ollama_url)
        if url.scheme not in {"http", "https"} or not url.hostname or url.username or url.password or url.query or url.fragment:
            raise ValueError("OLLAMA_URL must be an HTTP(S) base URL without credentials or query")
        if not settings.model or not 1 <= settings.timeout_seconds <= 300:
            raise ValueError("Invalid model or timeout configuration")
        return settings
