import copy
import json
from dataclasses import dataclass

import httpx
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.config import Settings

PROVIDER_API_KEY = "test-only-openrouter-api-key-123456789"
SERVICE_KEY = "unit-test-only-service-key-123456789"
ENDPOINT = "/internal/scenarios/generate"
PRIVATE_DETAIL = "private-upstream-token-and-stack-details"


def request_payload():
    return {
        "user_story": "As a registered user, I want to log in so I can access my account.",
        "acceptance_criteria": [
            "Valid credentials allow login.",
            "An incorrect password is rejected.",
        ],
        "max_scenarios": 4,
        "language": "en",
    }


def feature_payload():
    return {
        "feature_name": "User login",
        "scenarios": [
            {
                "name": "Valid credentials", "kind": "positive", "criterion_ids": [1],
                "given": ["a registered account exists"],
                "when": ["the user submits valid credentials"],
                "then": ["login succeeds"],
            },
            {
                "name": "Incorrect password", "kind": "negative", "criterion_ids": [2],
                "given": ["a registered account exists"],
                "when": ["the user submits an incorrect password"],
                "then": ["login is rejected"],
            },
        ],
        "warnings": [],
    }


def settings():
    return Settings(model="nvidia/fixture-nemotron:free", api_key=PROVIDER_API_KEY, service_key=SERVICE_KEY, timeout_seconds=2)


class OpenRouterResponder:
    """Only the external OpenRouter HTTP boundary is replaced in API/BDD tests."""

    def __init__(self, mode="valid", feature=None):
        self.mode = mode
        self.feature = copy.deepcopy(feature if feature is not None else feature_payload())
        self.requests = []

    def __call__(self, request):
        self.requests.append(request)
        if self.mode == "timeout":
            raise httpx.ReadTimeout(PRIVATE_DETAIL, request=request)
        if self.mode == "connection_error":
            raise httpx.ConnectError(PRIVATE_DETAIL, request=request)
        if self.mode == "internal_error":
            raise RuntimeError(PRIVATE_DETAIL)
        statuses = {
            "unavailable": 503, "not_found": 404, "unauthorized": 401,
            "forbidden": 403, "invalid_parameters": 400, "no_credits": 402,
            "unsupported_schema": 422, "rate_limited": 429, "redirect": 302,
        }
        if self.mode in statuses:
            return httpx.Response(statuses[self.mode], text=PRIVATE_DETAIL,
                                  headers={"Location": "https://untrusted.example.test"})
        if self.mode == "invalid_envelope_json":
            return httpx.Response(200, content=b"{")
        if self.mode == "oversized":
            return httpx.Response(200, content=b"x" * (1_048_576 + 1))
        if self.mode == "missing_message":
            return httpx.Response(200, json={"choices": [{"finish_reason": "stop"}]})
        if self.mode == "non_text_content":
            return httpx.Response(200, json={"choices": [{"finish_reason": "stop", "message": {"content": 42}}]})
        if self.mode == "list_envelope":
            return httpx.Response(200, json=[])
        data = copy.deepcopy(self.feature)
        if self.mode == "duplicate":
            data["scenarios"][1]["name"] = data["scenarios"][0]["name"].upper()
        elif self.mode == "unknown_criterion":
            data["scenarios"][0]["criterion_ids"] = [3]
        elif self.mode == "uncovered":
            data["scenarios"] = data["scenarios"][:1]
        elif self.mode == "invalid_schema":
            data["scenarios"][0]["given"] = []
        elif self.mode == "multiline":
            data["scenarios"][0]["then"] = ["accepted\nScenario: injected"]
        content = "not JSON: " + PRIVATE_DETAIL if self.mode == "invalid_model_json" else json.dumps(data)
        choice = {"finish_reason": "stop", "message": {"role": "assistant", "content": content}}
        if self.mode == "truncated":
            choice["finish_reason"] = "length"
        elif self.mode == "refusal":
            choice["message"]["refusal"] = "Request refused"
        elif self.mode == "empty_content":
            choice["message"]["content"] = ""
        elif self.mode == "no_choices":
            return httpx.Response(200, json={"choices": []})
        return httpx.Response(200, json={"choices": [choice]})


@dataclass
class Harness:
    app: FastAPI
    client: TestClient
    upstream: OpenRouterResponder

    def post(self, payload=None, key=SERVICE_KEY):
        headers = {} if key is None else {"X-AI-Service-Key": key}
        return self.client.post(ENDPOINT, json=request_payload() if payload is None else payload,
                                headers=headers)
