import asyncio
from contextlib import ExitStack

import httpx
import pytest
from fastapi.testclient import TestClient

from app.generator import ScenarioGenerator
from app.llm_client import OpenRouterClient
from app.main import create_app, get_generator
from tests.support import Harness, OpenRouterResponder, settings


@pytest.fixture(autouse=True)
def forbid_real_http(monkeypatch):
    """No real model/API access is permitted, even if a mock is forgotten."""
    def blocked_sync(*args, **kwargs):
        raise AssertionError("Real HTTP is disabled in this test suite")

    async def blocked_async(*args, **kwargs):
        raise AssertionError("Real HTTP is disabled in this test suite")

    monkeypatch.setattr(httpx.HTTPTransport, "handle_request", blocked_sync)
    monkeypatch.setattr(httpx.AsyncHTTPTransport, "handle_async_request", blocked_async)


@pytest.fixture
def harness_factory():
    clients = []
    with ExitStack() as stack:
        def build(mode="valid", feature=None):
            config = settings()
            responder = OpenRouterResponder(mode, feature)
            upstream = httpx.AsyncClient(transport=httpx.MockTransport(responder), trust_env=False)
            clients.append(upstream)
            generator = ScenarioGenerator(OpenRouterClient(upstream, config), config.model, config.timeout_seconds)
            app = create_app(config)
            app.dependency_overrides[get_generator] = lambda: generator
            client = stack.enter_context(TestClient(app))
            return Harness(app, client, responder)
        yield build
    for client in clients:
        asyncio.run(client.aclose())
