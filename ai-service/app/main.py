import asyncio
import logging
import secrets
from contextlib import asynccontextmanager

import httpx
from fastapi import Depends, FastAPI, HTTPException, Request
from fastapi.security import APIKeyHeader

from .config import Settings
from .generator import ScenarioGenerator
from .llm_client import InvalidModelOutput, ModelUnavailable, OllamaClient
from .schemas import GenerateRequest, GenerateResponse

logger = logging.getLogger(__name__)
key_header = APIKeyHeader(name="X-AI-Service-Key", auto_error=False)


def require_service_key(request: Request, key: str | None = Depends(key_header)) -> None:
    if key is None or not secrets.compare_digest(key, request.app.state.settings.service_key):
        raise HTTPException(status_code=401, detail="Invalid service credentials")


def get_generator(request: Request) -> ScenarioGenerator:
    return request.app.state.generator


def create_app(settings: Settings | None = None) -> FastAPI:
    @asynccontextmanager
    async def lifespan(app: FastAPI):
        config = settings or Settings.from_env()
        app.state.settings = config
        app.state.gate = asyncio.Semaphore(1)
        async with httpx.AsyncClient(
            timeout=httpx.Timeout(config.timeout_seconds, connect=5),
            trust_env=False,
        ) as client:
            app.state.generator = ScenarioGenerator(
                OllamaClient(client, config), config.model, config.timeout_seconds
            )
            yield

    app = FastAPI(title="Story2Gherkin AI Service", version="0.1.0", lifespan=lifespan)

    @app.get("/health")
    async def health():
        # Liveness only: does not claim the model is loaded or reachable.
        return {"status": "ok"}

    @app.post("/internal/scenarios/generate", response_model=GenerateResponse,
              dependencies=[Depends(require_service_key)])
    async def generate(request_body: GenerateRequest, request: Request,
                       generator: ScenarioGenerator = Depends(get_generator)):
        gate = request.app.state.gate
        try:
            async with asyncio.timeout(0.1):
                await gate.acquire()
        except TimeoutError:
            raise HTTPException(status_code=503, detail="Generator is busy; retry later")
        try:
            return await generator.generate(request_body)
        except (TimeoutError, httpx.TimeoutException):
            raise HTTPException(status_code=504, detail="Scenario generation timed out")
        except ModelUnavailable:
            raise HTTPException(status_code=503, detail="Language model is unavailable")
        except InvalidModelOutput:
            raise HTTPException(status_code=502, detail="Model returned an invalid scenario result")
        except Exception:
            # Do not log requirements, credentials, or raw model responses.
            logger.error("Unexpected scenario generation failure")
            raise HTTPException(status_code=500, detail="Scenario generation failed")
        finally:
            gate.release()

    return app


app = create_app()
