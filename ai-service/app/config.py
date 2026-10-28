import os
from dataclasses import dataclass, field
import math

from dotenv import load_dotenv


@dataclass(frozen=True)
class Settings:
    model: str
    api_key: str = field(repr=False)
    service_key: str = field(repr=False)
    timeout_seconds: float = 120
    response_format: str = "prompt"
    max_tokens: int = 4096

    def __post_init__(self):
        # Never automatically select a paid model or paid-model fallback.
        if not self.model.startswith("nvidia/") or "nemotron" not in self.model.lower() or not self.model.endswith(":free"):
            raise ValueError("OPENROUTER_MODEL must be your exact nvidia/...nemotron...:free model ID")
        if len(self.api_key.strip()) < 20 or self.api_key.startswith("REPLACE_"):
            raise ValueError("Set OPENROUTER_API_KEY to your OpenRouter API key")
        if len(self.service_key.strip()) < 32 or self.service_key.startswith("REPLACE_"):
            raise ValueError("Set AI_SERVICE_KEY to a random secret of at least 32 characters")
        if self.api_key == self.service_key:
            raise ValueError("OPENROUTER_API_KEY and AI_SERVICE_KEY must be different")
        if not math.isfinite(self.timeout_seconds) or not 1 <= self.timeout_seconds <= 300:
            raise ValueError("LLM_TIMEOUT_SECONDS must be between 1 and 300")
        if self.response_format not in {"prompt", "json_schema"}:
            raise ValueError("OPENROUTER_RESPONSE_FORMAT must be prompt or json_schema")
        if not 256 <= self.max_tokens <= 16384:
            raise ValueError("OPENROUTER_MAX_TOKENS must be between 256 and 16384")

    @classmethod
    def from_env(cls) -> "Settings":
        load_dotenv()
        return cls(
            model=os.getenv("OPENROUTER_MODEL", "").strip(),
            api_key=os.getenv("OPENROUTER_API_KEY", "").strip(),
            service_key=os.getenv("AI_SERVICE_KEY", "").strip(),
            timeout_seconds=float(os.getenv("LLM_TIMEOUT_SECONDS", "120")),
            response_format=os.getenv("OPENROUTER_RESPONSE_FORMAT", "prompt").strip().lower(),
            max_tokens=int(os.getenv("OPENROUTER_MAX_TOKENS", "4096")),
        )
