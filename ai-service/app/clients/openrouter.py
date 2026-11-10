import asyncio
import json
import logging

import httpx

from app.clients.base import InvalidModelOutput, ModelUnavailable
from app.core.config import Settings
from app.services.prompts import build_messages
from app.schemas.generate import GenerateRequest

logger = logging.getLogger(__name__)
TRANSIENT_STATUS = (401, 408, 429)
MAX_TRANSPORT_ATTEMPTS = 3


class OpenRouterClient:
    """Transport Chat Completions OpenRouter, mode teks Gherkin.

    Tanpa response_format: model menulis Gherkin mentah, struktur
    diekstrak + divalidasi parser resmi di ScenarioGenerator. Jalan di
    model apapun termasuk yang tak mendukung structured outputs.
    """

    def __init__(self, client: httpx.AsyncClient, settings: Settings):
        self.client = client
        self.settings = settings
        self.last_model: str = ""
        self.last_usage: dict = {}

    async def generate(
        self, request: GenerateRequest, feedback: str | None = None
    ) -> str:
        payload: dict = {
            "model": self.settings.model,
            "messages": build_messages(request, feedback),
            "temperature": 0,
            "max_tokens": self.settings.max_tokens,
        }
        if "openrouter.ai" in self.settings.base_url:
            payload["models"] = [self.settings.model, *self.settings.fallbacks]
            payload["reasoning"] = {"effort": self.settings.reasoning_effort}
            if self.settings.provider_order:
                payload["provider"] = {"order": list(self.settings.provider_order)}
        last_exc: Exception | None = None
        for attempt in range(1, MAX_TRANSPORT_ATTEMPTS + 1):
            try:
                status, raw, model, usage = await self._post(payload)
            except (httpx.TimeoutException, httpx.HTTPError) as exc:
                last_exc = exc
                logger.error("[openrouter-error] attempt=%d transport=%r", attempt, exc)
                await asyncio.sleep(attempt)
                continue
            if status == 429 or status in TRANSIENT_STATUS or status >= 500:
                last_exc = ModelUnavailable(f"Provider error {status}")
                logger.error(
                    "[openrouter-error] attempt=%d status=%s body=%.300s",
                    attempt, status, raw[:300],
                )
                await asyncio.sleep(attempt)
                continue
            if status != 200:
                logger.error(
                    "[openrouter-error] status=%s body=%.300s",
                    status, raw[:300],
                )
                raise InvalidModelOutput(f"Unexpected status {status}")
            self.last_model = model
            self.last_usage = usage
            return raw
        raise ModelUnavailable(f"Giving up after retries: {last_exc}") from last_exc

    async def _post(self, payload: dict) -> tuple[int, str, str, dict]:
        async with self.client.stream(
            "POST", self.settings.base_url + "/chat/completions", json=payload
        ) as response:
            status = response.status_code
            body = bytearray()
            async for chunk in response.aiter_bytes():
                body.extend(chunk)
                if len(body) > 1_048_576:
                    raise InvalidModelOutput("Model response exceeds limit")
        text = bytes(body).decode("utf-8", errors="replace")
        if status != 200:
            return status, text, "", {}
        try:
            envelope = json.loads(text)
            choice = envelope["choices"][0]
            content = choice["message"]["content"]
            if not isinstance(content, str) or not content.strip():
                raise InvalidModelOutput("Empty model response")
            if choice.get("finish_reason") == "length":
                raise InvalidModelOutput(
                    "Response truncated by token limit; return a shorter complete feature"
                )
            model = str(envelope.get("model", self.settings.model))
            usage = envelope.get("usage")
            return status, content, model, dict(usage) if isinstance(usage, dict) else {}
        except InvalidModelOutput:
            raise
        except (ValueError, KeyError, IndexError, TypeError) as exc:
            raise InvalidModelOutput("Unreadable model response") from exc
