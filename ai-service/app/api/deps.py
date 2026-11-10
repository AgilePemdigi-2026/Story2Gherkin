import secrets

from fastapi import Depends, HTTPException, Request
from fastapi.security import APIKeyHeader

from app.services.generator import ScenarioGenerator

key_header = APIKeyHeader(name="X-AI-Service-Key", auto_error=False)


def require_service_key(request: Request, key: str | None = Depends(key_header)) -> None:
    if key is None or not secrets.compare_digest(key, request.app.state.settings.service_key):
        raise HTTPException(status_code=401, detail="Invalid service credentials")


def get_generator(request: Request) -> ScenarioGenerator:
    return request.app.state.generator
