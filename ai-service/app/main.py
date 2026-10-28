import asyncio
from contextlib import asynccontextmanager

import httpx
from fastapi import Depends, FastAPI

from app.api.deps import require_service_key
from app.api.routes import generate as generate_routes
from app.api.routes import health as health_routes
from app.clients.openrouter import OpenRouterClient
from app.core.config import Settings
from app.services.generator import ScenarioGenerator


def create_app(settings: Settings | None = None) -> FastAPI:
    @asynccontextmanager
    async def lifespan(app: FastAPI):
        config = settings or Settings.from_env()
        app.state.settings = config
        # Serializes LLM calls: free-tier quota is tiny and shared.
        # TODO: ganti dengan rate-limit per-user di backend (SRS 5.3), lalu lepas gate ini.
        app.state.gate = asyncio.Semaphore(1)
        async with httpx.AsyncClient(
            timeout=httpx.Timeout(config.timeout_seconds, connect=5),
            trust_env=False,
        ) as client:
            app.state.generator = ScenarioGenerator(
                OpenRouterClient(client, config), config.model, config.timeout_seconds
            )
            yield

    app = FastAPI(title="Story2Gherkin AI Service", version="0.2.0", lifespan=lifespan)
    app.include_router(health_routes.router)
    app.include_router(generate_routes.router, dependencies=[Depends(require_service_key)])
    return app


app = create_app()
