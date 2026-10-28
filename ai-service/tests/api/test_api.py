import asyncio

import pytest

from app import main
from tests.support import ENDPOINT, PRIVATE_DETAIL, PROVIDER_API_KEY, SERVICE_KEY, request_payload, settings

pytestmark = pytest.mark.api


def test_complete_http_generation_path(harness_factory):
    harness = harness_factory()
    response = harness.post()
    assert response.status_code == 200
    data = response.json()
    assert data["model"] == "nvidia/fixture-nemotron:free"
    assert data["prompt_version"] == "scenario-v1"
    assert len(data["feature"]["scenarios"]) == 2
    assert data["gherkin"].startswith("Feature: User login\n")
    assert len(harness.upstream.requests) == 1


@pytest.mark.parametrize("key", [None, "wrong-key"])
def test_service_authentication_prevents_model_call(harness_factory, key):
    harness = harness_factory()
    response = harness.post(key=key)
    assert response.status_code == 401
    assert response.json() == {"detail": "Invalid service credentials"}
    assert not harness.upstream.requests


@pytest.mark.parametrize("payload", [
    {}, {**request_payload(), "acceptance_criteria": []},
    {**request_payload(), "max_scenarios": 7},
    {**request_payload(), "user_story": "short"},
    {**request_payload(), "language": "unsupported"},
    {**request_payload(), "extra_field": True},
])
def test_invalid_input_never_reaches_model(harness_factory, payload):
    harness = harness_factory()
    response = harness.post(payload)
    assert response.status_code == 422
    assert not harness.upstream.requests


def test_malformed_json(harness_factory):
    harness = harness_factory()
    response = harness.client.post(ENDPOINT, content="{", headers={"Content-Type": "application/json", "X-AI-Service-Key": SERVICE_KEY})
    assert response.status_code == 422
    assert not harness.upstream.requests


@pytest.mark.parametrize("mode,status,detail", [
    ("unavailable", 503, "Language model is unavailable"),
    ("rate_limited", 503, "Language model rate limit reached; retry later"),
    ("unauthorized", 502, "Language model credentials or configuration are invalid"),
    ("unsupported_schema", 502, "Language model credentials or configuration are invalid"),
    ("connection_error", 503, "Language model is unavailable"),
    ("timeout", 504, "Scenario generation timed out"),
    ("invalid_model_json", 502, "Model returned an invalid scenario result"),
    ("duplicate", 502, "Model returned an invalid scenario result"),
    ("unknown_criterion", 502, "Model returned an invalid scenario result"),
    ("multiline", 502, "Model returned an invalid scenario result"),
    ("internal_error", 500, "Scenario generation failed"),
])
def test_failure_responses_are_safe_and_release_slot(harness_factory, mode, status, detail):
    harness = harness_factory(mode)
    response = harness.post()
    assert response.status_code == status
    assert response.json() == {"detail": detail}
    assert PRIVATE_DETAIL not in response.text and SERVICE_KEY not in response.text and PROVIDER_API_KEY not in response.text
    assert len(harness.upstream.requests) == 1
    # A subsequent request succeeds on the same application after the error.
    harness.upstream.mode = "valid"
    assert harness.post().status_code == 200


def test_busy_generator_does_not_start_another_model_call(harness_factory):
    harness = harness_factory()
    harness.app.state.gate = asyncio.Semaphore(0)
    response = harness.post()
    assert response.status_code == 503
    assert response.json() == {"detail": "Generator is busy; retry later"}
    assert not harness.upstream.requests


def test_health_is_liveness_and_does_not_load_model(harness_factory):
    harness = harness_factory("unavailable")
    response = harness.client.get("/health")
    assert response.status_code == 200 and response.json() == {"status": "ok"}
    assert not harness.upstream.requests


def test_unsupported_method(harness_factory):
    harness = harness_factory()
    assert harness.client.get(ENDPOINT).status_code == 405
    assert not harness.upstream.requests


def test_startup_loads_settings_when_not_explicitly_injected(monkeypatch):
    from fastapi.testclient import TestClient
    calls = []
    def from_env():
        calls.append(True)
        return settings()
    monkeypatch.setattr(main.Settings, "from_env", from_env)
    with TestClient(main.create_app()) as client:
        assert client.get("/health").status_code == 200
    assert calls == [True]


def test_application_uses_its_registered_generator(monkeypatch):
    from fastapi.testclient import TestClient
    from app.schemas import GeneratedFeature
    from tests.support import feature_payload

    calls = []
    async def fake_inference(self, request):
        calls.append(request)
        return GeneratedFeature.model_validate(feature_payload())
    monkeypatch.setattr(main.OpenRouterClient, "generate", fake_inference)
    # Do not override get_generator: verify the production dependency lookup.
    with TestClient(main.create_app(settings())) as client:
        response = client.post(ENDPOINT, json=request_payload(),
                               headers={"X-AI-Service-Key": SERVICE_KEY})
    assert response.status_code == 200
    assert len(calls) == 1
