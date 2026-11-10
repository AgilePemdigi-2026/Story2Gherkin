from typing import Protocol

from app.schemas.generate import GenerateRequest


class ModelUnavailable(Exception):
    """LLM tidak bisa dihubungi: timeout, rate-limit, downtime provider."""


class InvalidModelOutput(Exception):
    """Respons model ada tapi tidak bisa dipakai: terpotong, bukan Gherkin valid."""


class ModelClient(Protocol):
    """Port ke LLM: mengembalikan teks Gherkin mentah.

    Parsing + validasi struktur milik ScenarioGenerator (parser resmi),
    bukan model — sehingga klien ini jalan di model APAPUN, termasuk
    yang tak mendukung structured outputs.
    """

    async def generate(
        self, request: GenerateRequest, feedback: str | None = None
    ) -> str: ...
