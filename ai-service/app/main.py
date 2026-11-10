import asyncio
from contextlib import asynccontextmanager

import httpx
from fastapi import Depends, FastAPI

from app.api.deps import require_service_key
from app.api.routes import generate as generate_routes
from app.api.routes import health as health_routes
from app.api.routes import review as review_routes
from app.clients.ollama import OllamaClient
from app.clients.openrouter import OpenRouterClient
from app.core.config import Settings
from app.services.generator import ScenarioGenerator


def build_client(client: httpx.AsyncClient, config: Settings):
    if config.provider == "ollama":
        return OllamaClient(client, config), config.ollama_model
    return OpenRouterClient(client, config), config.model


def create_app(settings: Settings | None = None) -> FastAPI:
    @asynccontextmanager
    async def lifespan(app: FastAPI):
        config = settings or Settings.from_env()
        app.state.settings = config
        app.state.gate = asyncio.Semaphore(1)

        async with httpx.AsyncClient(
            timeout=httpx.Timeout(config.timeout_seconds, connect=5),
            trust_env=False,
            headers={"Connection": "close"},
        ) as client:
            model_client, model_name = build_client(client, config)
            print(
                f"ai-service ready | provider={config.provider} "
                f"model={model_name}",
                flush=True,
            )
            app.state.generator = ScenarioGenerator(
                model_client, model_name, config.timeout_seconds
            )
            yield

    app = FastAPI(
        title="Story2Gherkin AI Service",
        version="0.2.0",
        lifespan=lifespan,
    )

    app.include_router(health_routes.router)
    app.include_router(
        generate_routes.router,
        dependencies=[Depends(require_service_key)],
    )
    app.include_router(
        review_routes.router,
        dependencies=[Depends(require_service_key)],
    )

    return app


app = create_app()
