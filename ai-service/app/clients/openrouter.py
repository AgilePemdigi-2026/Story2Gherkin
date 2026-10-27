import httpx
from pydantic import ValidationError

from app.clients.base import InvalidModelOutput, ModelUnavailable
from app.core.config import Settings
from app.schemas.generate import GeneratedFeature, GenerateRequest
from app.services.prompts import build_messages


class OpenRouterClient:
    """Transport Chat Completions OpenRouter.

    Basis riset: response_format json_schema (skema ditegakkan provider),
    models[] untuk fallback antar-model, retry hanya untuk 429/5xx/timeout.
    """

    def __init__(self, client: httpx.AsyncClient, settings: Settings):
        self.client = client
        self.settings = settings
        self.last_model: str = ""
        self.last_usage: dict = {}

    async def generate(
        self, request: GenerateRequest, feedback: str | None = None
    ) -> GeneratedFeature:
        payload = {
            "model": self.settings.model,
            "models": [self.settings.model, *self.settings.fallbacks],
            "messages": build_messages(request, feedback),
            "response_format": {
                "type": "json_schema",
                "json_schema": {
                    "name": "gherkin_feature",
                    "strict": True,
                    "schema": GeneratedFeature.model_json_schema(),
                },
            },
            "temperature": 0,
            "max_tokens": self.settings.max_tokens,
            "reasoning": {"effort": self.settings.reasoning_effort},
        }
        try:
            async with self.client.stream(
                "POST", self.settings.base_url + "/chat/completions", json=payload
            ) as response:
                status = response.status_code
                body = bytearray()
                async for chunk in response.aiter_bytes():
                    body.extend(chunk)
                    if len(body) > 1_048_576:
                        raise InvalidModelOutput("Model response exceeds limit")
        except httpx.TimeoutException:
            raise
        except httpx.HTTPError as exc:
            raise ModelUnavailable() from exc

        if status == 429:
            raise ModelUnavailable("Rate limited")
        if status in (401, 408) or status >= 500:
            raise ModelUnavailable(f"Provider error {status}")
        if status != 200:
            raise InvalidModelOutput(f"Unexpected status {status}")

        try:
            envelope = httpx.Response(200, content=bytes(body)).json()
            choice = envelope["choices"][0]
            if choice.get("finish_reason") == "length":
                raise InvalidModelOutput(
                    "Response truncated by token limit; return a shorter complete JSON"
                )
            content = choice["message"]["content"]
            if not isinstance(content, str):
                raise TypeError("Expected text output")
            feature = GeneratedFeature.model_validate_json(content)
        except ValidationError as exc:
            errs = exc.errors()
            if errs:
                loc = ".".join(str(p) for p in errs[0].get("loc", ()))
                detail = f"{loc}: {errs[0].get('msg', 'invalid')}"
            else:
                detail = "schema violation"
            raise InvalidModelOutput(f"Schema violation: {detail}") from exc
        except (ValueError, KeyError, IndexError, TypeError) as exc:
            raise InvalidModelOutput("Invalid structured model response") from exc

        self.last_model = str(envelope.get("model", self.settings.model))
        usage = envelope.get("usage")
        self.last_usage = dict(usage) if isinstance(usage, dict) else {}
        return feature
