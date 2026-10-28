from typing import Protocol

from app.schemas.generate import GenerateRequest, GeneratedFeature


class ModelUnavailable(Exception):
    """LLM tidak bisa dihubungi: timeout, rate-limit, downtime provider."""


class InvalidModelOutput(Exception):
    """Respons model ada tapi tidak bisa dipakai: skema salah, melanggar batas."""


class ModelClient(Protocol):
    async def generate(
        self, request: GenerateRequest, feedback: str | None = None
    ) -> GeneratedFeature: ...
