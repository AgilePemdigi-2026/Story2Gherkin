import asyncio
from dataclasses import replace
import json

import httpx
import pytest

from app.llm_client import (
    CHAT_URL, InvalidModelOutput, ModelConfigurationError,
    ModelRateLimited, ModelUnavailable, OpenRouterClient, response_schema,
)
from app.prompts import SYSTEM_PROMPT
from app.schemas import GenerateRequest
from tests.support import OpenRouterResponder, PROVIDER_API_KEY, feature_payload, request_payload, settings

pytestmark = pytest.mark.unit


def invoke(mode="valid", transport=None, config=None):
    responder = OpenRouterResponder(mode)

    async def run():
        async with httpx.AsyncClient(transport=transport or httpx.MockTransport(responder), trust_env=False) as client:
            return await OpenRouterClient(client, config or settings()).generate(GenerateRequest.model_validate(request_payload()))
    return run, responder


def test_upstream_request_and_result():
    run, responder = invoke()
    feature = asyncio.run(run())
    assert feature.feature_name == "User login"
    assert len(responder.requests) == 1
    request = responder.requests[0]
    assert request.method == "POST" and str(request.url) == CHAT_URL
    assert request.headers["Authorization"] == "Bearer " + PROVIDER_API_KEY
    assert "X-AI-Service-Key" not in request.headers
    payload = json.loads(request.content)
    assert payload["model"].endswith(":free")
    assert payload["stream"] is False
    assert payload["temperature"] == 0
    assert payload["max_tokens"] == 4096
    assert "response_format" not in payload
    assert "format" not in payload and "options" not in payload
    assert settings().service_key not in request.content.decode()
    assert PROVIDER_API_KEY not in request.content.decode()
    assert "Required output JSON schema:" in payload["messages"][0]["content"]
    assert SYSTEM_PROMPT in payload["messages"][0]["content"]


def test_native_json_schema_is_opt_in():
    run, responder = invoke(config=replace(settings(), response_format="json_schema"))
    asyncio.run(run())
    payload = json.loads(responder.requests[0].content)
    assert payload["response_format"]["type"] == "json_schema"
    assert payload["response_format"]["json_schema"]["strict"] is True
    assert payload["provider"] == {"require_parameters": True}
    schema = payload["response_format"]["json_schema"]["schema"]
    assert schema["additionalProperties"] is False
    assert "warnings" in schema["required"]
    assert "default" not in json.dumps(schema)


def test_schema_preparation_does_not_mutate_model_defaults():
    first = response_schema()
    first["required"].append("changed")
    assert "changed" not in response_schema()["required"]


@pytest.mark.parametrize("mode", [
    "invalid_envelope_json", "missing_message", "non_text_content", "list_envelope",
    "invalid_model_json", "invalid_schema", "multiline", "oversized",
    "truncated", "refusal", "empty_content", "no_choices",
])
def test_invalid_upstream_output(mode):
    run, responder = invoke(mode)
    with pytest.raises(InvalidModelOutput):
        asyncio.run(run())
    assert len(responder.requests) == 1


@pytest.mark.parametrize("mode", ["unavailable", "connection_error"])
def test_model_unavailable(mode):
    run, responder = invoke(mode)
    with pytest.raises(ModelUnavailable):
        asyncio.run(run())
    assert len(responder.requests) == 1


@pytest.mark.parametrize("mode", ["unauthorized", "forbidden", "no_credits", "not_found", "invalid_parameters", "unsupported_schema", "redirect"])
def test_configuration_or_credentials_error(mode):
    run, responder = invoke(mode)
    with pytest.raises(ModelConfigurationError):
        asyncio.run(run())
    assert len(responder.requests) == 1


def test_rate_limit_is_distinguished():
    run, responder = invoke("rate_limited")
    with pytest.raises(ModelRateLimited):
        asyncio.run(run())
    assert len(responder.requests) == 1


def test_timeout_is_preserved():
    run, _ = invoke("timeout")
    with pytest.raises(httpx.ReadTimeout):
        asyncio.run(run())


def test_stream_is_closed_after_success():
    class Stream(httpx.AsyncByteStream):
        closed = False

        async def __aiter__(self):
            encoded = json.dumps({"choices": [{
                "finish_reason": "stop", "message": {"content": json.dumps(feature_payload())}
            }]}).encode()
            yield encoded[:20]
            yield encoded[20:]

        async def aclose(self):
            self.closed = True
    stream = Stream()
    run, _ = invoke(transport=httpx.MockTransport(lambda request: httpx.Response(200, stream=stream)))
    asyncio.run(run())
    assert stream.closed


def test_stream_is_closed_after_size_limit():
    class OversizedStream(httpx.AsyncByteStream):
        closed = False

        async def __aiter__(self):
            yield b"x" * 600_000
            yield b"x" * 600_000

        async def aclose(self):
            self.closed = True
    stream = OversizedStream()
    run, _ = invoke(transport=httpx.MockTransport(lambda request: httpx.Response(200, stream=stream)))
    with pytest.raises(InvalidModelOutput):
        asyncio.run(run())
    assert stream.closed
