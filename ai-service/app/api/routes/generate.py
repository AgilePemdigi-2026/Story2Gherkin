import asyncio
import logging

import httpx
from fastapi import APIRouter, Depends, HTTPException, Request

from app.api.deps import get_generator
from app.clients.base import InvalidModelOutput, ModelUnavailable
from app.schemas.generate import GenerateRequest, GenerateResponse
from app.services.generator import ScenarioGenerator
from app.services.rule_based import generate_rule_based

logger = logging.getLogger(__name__)

router = APIRouter(tags=["scenarios"])


@router.post("/internal/scenarios/generate", response_model=GenerateResponse)
async def generate(request_body: GenerateRequest, request: Request,
                   generator: ScenarioGenerator = Depends(get_generator)):
    gate = request.app.state.gate
    try:
        await asyncio.wait_for(gate.acquire(), timeout=0.1)
    except asyncio.TimeoutError:
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


@router.post("/internal/scenarios/generate-rule-based", response_model=GenerateResponse)
async def generate_rule_based_route(request_body: GenerateRequest):
    return generate_rule_based(request_body)
