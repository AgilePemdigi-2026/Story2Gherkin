import os
from dataclasses import dataclass
from pathlib import Path
from urllib.parse import urlparse

from dotenv import load_dotenv

ENV_FILE = Path(__file__).resolve().parent.parent.parent / ".env"


@dataclass(frozen=True)
class Settings:
    provider: str = "ollama"
    base_url: str = "https://openrouter.ai/api/v1"
    api_key: str = ""
    model: str = ""
    fallbacks: tuple[str, ...] = ()
    max_tokens: int = 4096
    reasoning_effort: str = "none"
    provider_order: tuple[str, ...] = ()
    ollama_url: str = "http://127.0.0.1:11434"
    ollama_model: str = ""
    ollama_num_ctx: int = 8192
    ollama_num_predict: int = 4096
    service_key: str = ""
    timeout_seconds: float = 120

    @classmethod
    def from_env(cls) -> "Settings":
        load_dotenv(ENV_FILE)
        raw_fallbacks = os.getenv("OPENROUTER_FALLBACKS", "")
        settings = cls(
            provider=os.getenv("LLM_PROVIDER", "ollama").strip().lower(),
            base_url=os.getenv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1").rstrip("/"),
            api_key=os.getenv("OPENROUTER_API_KEY", ""),
            model=os.getenv("OPENROUTER_MODEL", "").strip(),
            fallbacks=tuple(m.strip() for m in raw_fallbacks.split(",") if m.strip()),
            max_tokens=int(os.getenv("OPENROUTER_MAX_TOKENS", "8192")),
            reasoning_effort=os.getenv("OPENROUTER_REASONING_EFFORT", "none").strip().lower(),
            provider_order=tuple(
                p.strip() for p in os.getenv("OPENROUTER_PROVIDER_ORDER", "").split(",") if p.strip()
            ),
            ollama_url=os.getenv("OLLAMA_URL", "http://127.0.0.1:11434").rstrip("/"),
            ollama_model=os.getenv("OLLAMA_MODEL", "").strip(),
            ollama_num_ctx=int(os.getenv("OLLAMA_NUM_CTX", "4096")),
            ollama_num_predict=int(os.getenv("OLLAMA_NUM_PREDICT", "2048")),
            service_key=os.getenv("AI_SERVICE_KEY", ""),
            timeout_seconds=float(os.getenv("LLM_TIMEOUT_SECONDS", os.getenv("OPENROUTER_TIMEOUT_SECONDS", "120"))),
        )
        if len(settings.service_key.strip()) < 32 or settings.service_key.startswith("REPLACE_"):
            raise ValueError("Set AI_SERVICE_KEY to a random secret of at least 32 characters")
        if not 1 <= settings.timeout_seconds <= 300:
            raise ValueError("Invalid timeout configuration")
        if settings.provider == "ollama":
            url = urlparse(settings.ollama_url)
            if url.scheme not in {"http", "https"} or not url.hostname:
                raise ValueError("OLLAMA_URL must be an HTTP(S) base URL")
            if not settings.ollama_model:
                raise ValueError("Set OLLAMA_MODEL, e.g. qwen2.5:3b")
            if settings.ollama_num_ctx < 512 or settings.ollama_num_predict < 1:
                raise ValueError("OLLAMA_NUM_CTX must be at least 512 and OLLAMA_NUM_PREDICT must be positive")
        elif settings.provider == "openrouter":
            if not settings.api_key or settings.api_key.startswith("REPLACE_"):
                raise ValueError("Set OPENROUTER_API_KEY to your OpenRouter key")
            url = urlparse(settings.base_url)
            if url.scheme not in {"http", "https"} or not url.hostname or url.username or url.password or url.query or url.fragment:
                raise ValueError("OPENROUTER_BASE_URL must be an HTTP(S) base URL without credentials or query")
            if not settings.model:
                raise ValueError("Set OPENROUTER_MODEL to a model slug, e.g. vendor/name:free")
            if settings.max_tokens < 512:
                raise ValueError("OPENROUTER_MAX_TOKENS too small for scenario output")
            if settings.reasoning_effort not in {"max", "xhigh", "high", "medium", "low", "minimal", "none"}:
                raise ValueError("OPENROUTER_REASONING_EFFORT must be one of max/xhigh/high/medium/low/minimal/none")
        else:
            raise ValueError("LLM_PROVIDER must be openrouter or ollama")
        return settings
