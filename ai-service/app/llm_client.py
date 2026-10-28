import json

import httpx
from pydantic import ValidationError

from .config import Settings
from .prompts import build_messages
from .schemas import GeneratedFeature, GenerateRequest

CHAT_URL = "https://openrouter.ai/api/v1/chat/completions"
MAX_RESPONSE_BYTES = 1_048_576


class ModelUnavailable(Exception):
    pass


class ModelRateLimited(ModelUnavailable):
    pass


class ModelConfigurationError(Exception):
    pass


class InvalidModelOutput(Exception):
    pass


def response_schema() -> dict:
    """Prepare object requirements for strict JSON-schema providers."""
    schema = GeneratedFeature.model_json_schema()

    def visit(node):
        if isinstance(node, dict):
            node.pop("default", None)
            if node.get("type") == "object":
                node["additionalProperties"] = False
                node["required"] = list(node.get("properties", {}))
            for child in node.values():
                visit(child)
        elif isinstance(node, list):
            for child in node:
                visit(child)
    visit(schema)
    return schema


class OpenRouterClient:
    def __init__(self, client: httpx.AsyncClient, settings: Settings):
        self.client = client
        self.settings = settings

    async def generate(self, request: GenerateRequest) -> GeneratedFeature:
        schema = response_schema()
        messages = build_messages(request)
        messages[0]["content"] += "\nRequired output JSON schema:\n" + json.dumps(schema)
        payload = {
            "model": self.settings.model,
            "messages": messages,
            "stream": False,
            "temperature": 0,
            "max_tokens": self.settings.max_tokens,
        }
        # Prompt mode works without provider-native structured outputs.
        # Returned JSON is still validated locally in both modes.
        if self.settings.response_format == "json_schema":
            payload["response_format"] = {
                "type": "json_schema",
                "json_schema": {"name": "gherkin_feature", "strict": True, "schema": schema},
            }
            payload["provider"] = {"require_parameters": True}
        headers = {
            "Authorization": "Bearer " + self.settings.api_key,
            "Content-Type": "application/json",
        }
        try:
            async with self.client.stream(
                "POST", CHAT_URL, json=payload, headers=headers, follow_redirects=False
            ) as response:
                if response.status_code == 429:
                    raise ModelRateLimited()
                if response.status_code in {400, 401, 402, 403, 404, 422} or 300 <= response.status_code < 400:
                    raise ModelConfigurationError()
                response.raise_for_status()
                body = bytearray()
                async for chunk in response.aiter_bytes():
                    body.extend(chunk)
                    if len(body) > MAX_RESPONSE_BYTES:
                        raise InvalidModelOutput("Model response exceeds limit")
        except httpx.TimeoutException:
            raise
        except httpx.HTTPError as exc:
            raise ModelUnavailable() from exc

        try:
            envelope = json.loads(body)
            if not isinstance(envelope, dict):
                raise ValueError("Expected a completion object")
            choices = envelope.get("choices")
            if not isinstance(choices, list) or not choices or not isinstance(choices[0], dict):
                raise ValueError("Missing completion choices")
            choice = choices[0]
            if choice.get("finish_reason") != "stop":
                raise ValueError("Incomplete, refused, or unexpected completion")
            message = choice.get("message")
            if not isinstance(message, dict) or message.get("refusal"):
                raise ValueError("Missing or refused completion")
            content = message.get("content")
            if not isinstance(content, str) or not content.strip():
                raise ValueError("Expected nonempty text")
            return GeneratedFeature.model_validate_json(content)
        except (ValueError, TypeError, ValidationError) as exc:
            raise InvalidModelOutput("Invalid structured model response") from exc
