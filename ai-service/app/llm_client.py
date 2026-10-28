import httpx
from pydantic import ValidationError

from .config import Settings
from .prompts import build_messages
from .schemas import GeneratedFeature, GenerateRequest


class ModelUnavailable(Exception):
    pass


class InvalidModelOutput(Exception):
    pass


class OllamaClient:
    def __init__(self, client: httpx.AsyncClient, settings: Settings):
        self.client = client
        self.settings = settings

    async def generate(self, request: GenerateRequest) -> GeneratedFeature:
        payload = {
            "model": self.settings.model,
            "messages": build_messages(request),
            "format": GeneratedFeature.model_json_schema(),
            "stream": False,
            "options": {"temperature": 0, "num_predict": 4096, "num_ctx": 8192},
        }
        try:
            # Limit response buffering even if the upstream behaves unexpectedly.
            async with self.client.stream(
                "POST", self.settings.ollama_url + "/api/chat", json=payload
            ) as response:
                response.raise_for_status()
                body = bytearray()
                async for chunk in response.aiter_bytes():
                    body.extend(chunk)
                    if len(body) > 1_048_576:
                        raise InvalidModelOutput("Model response exceeds limit")
        except httpx.TimeoutException:
            raise
        except httpx.HTTPError as exc:
            raise ModelUnavailable() from exc

        try:
            envelope = httpx.Response(200, content=bytes(body)).json()
            content = envelope["message"]["content"]
            if not isinstance(content, str):
                raise TypeError("Expected text output")
            return GeneratedFeature.model_validate_json(content)
        except (ValueError, KeyError, TypeError, ValidationError) as exc:
            raise InvalidModelOutput("Invalid structured model response") from exc
