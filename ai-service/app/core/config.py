import os
from dataclasses import dataclass
from urllib.parse import urlparse

from dotenv import load_dotenv


@dataclass(frozen=True)
class Settings:
    base_url: str
    api_key: str
    model: str
    fallbacks: tuple[str, ...] = ()
    max_tokens: int = 4096
    reasoning_effort: str = "none"
    service_key: str = ""
    timeout_seconds: float = 120

    @classmethod
    def from_env(cls) -> "Settings":
        load_dotenv()
        raw_fallbacks = os.getenv("OPENROUTER_FALLBACKS", "")
        settings = cls(
            base_url=os.getenv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1").rstrip("/"),
            api_key=os.getenv("OPENROUTER_API_KEY", ""),
            model=os.getenv("OPENROUTER_MODEL", "").strip(),
            fallbacks=tuple(m.strip() for m in raw_fallbacks.split(",") if m.strip()),
            max_tokens=int(os.getenv("OPENROUTER_MAX_TOKENS", "4096")),
            reasoning_effort=os.getenv("OPENROUTER_REASONING_EFFORT", "none").strip().lower(),
            service_key=os.getenv("AI_SERVICE_KEY", ""),
            timeout_seconds=float(os.getenv("OPENROUTER_TIMEOUT_SECONDS", "120")),
        )
        if not settings.api_key or settings.api_key.startswith("REPLACE_"):
            raise ValueError("Set OPENROUTER_API_KEY to your OpenRouter key")
        if len(settings.service_key.strip()) < 32 or settings.service_key.startswith("REPLACE_"):
            raise ValueError("Set AI_SERVICE_KEY to a random secret of at least 32 characters")
        url = urlparse(settings.base_url)
        if url.scheme not in {"http", "https"} or not url.hostname or url.username or url.password or url.query or url.fragment:
            raise ValueError("OPENROUTER_BASE_URL must be an HTTP(S) base URL without credentials or query")
        if not settings.model:
            raise ValueError("Set OPENROUTER_MODEL to a model slug, e.g. vendor/name:free")
        if not 1 <= settings.timeout_seconds <= 300:
            raise ValueError("Invalid timeout configuration")
        if settings.max_tokens < 512:
            raise ValueError("OPENROUTER_MAX_TOKENS too small for scenario output")
        if settings.reasoning_effort not in {"max", "xhigh", "high", "medium", "low", "minimal", "none"}:
            raise ValueError("OPENROUTER_REASONING_EFFORT must be one of max/xhigh/high/medium/low/minimal/none")
        return settings
