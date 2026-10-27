import asyncio
import time
from typing import Protocol

from gherkin.parser import Parser

from .llm_client import InvalidModelOutput
from .prompts import PROMPT_VERSION
from .schemas import GeneratedFeature, GenerateRequest, GenerateResponse


class ModelClient(Protocol):
    async def generate(self, request: GenerateRequest) -> GeneratedFeature: ...


def render_gherkin(feature: GeneratedFeature) -> str:
    lines = [f"Feature: {feature.feature_name}"]
    for scenario in feature.scenarios:
        lines.extend(["", f"  Scenario: {scenario.name}"])
        for keyword, steps in (("Given", scenario.given), ("When", scenario.when), ("Then", scenario.then)):
            lines.extend(
                f"    {keyword if index == 0 else 'And'} {step}"
                for index, step in enumerate(steps)
            )
    return "\n".join(lines) + "\n"


class ScenarioGenerator:
    def __init__(self, client: ModelClient, model: str, timeout_seconds: float = 120):
        self.client = client
        self.model = model
        self.timeout_seconds = timeout_seconds

    async def generate(self, request: GenerateRequest) -> GenerateResponse:
        started = time.monotonic()
        # Overall deadline includes loading and inference, not only individual socket reads.
        async with asyncio.timeout(self.timeout_seconds):
            feature = await self.client.generate(request)
        if len(feature.scenarios) > request.max_scenarios:
            raise InvalidModelOutput("Too many scenarios")
        seen = set()
        covered = set()
        for scenario in feature.scenarios:
            if scenario.name.casefold() in seen:
                raise InvalidModelOutput("Duplicate scenario names")
            seen.add(scenario.name.casefold())
            if any(cid > len(request.acceptance_criteria) for cid in scenario.criterion_ids):
                raise InvalidModelOutput("Unknown acceptance criterion reference")
            covered.update(scenario.criterion_ids)
        missing = sorted(set(range(1, len(request.acceptance_criteria) + 1)) - covered)
        if missing:
            feature.warnings = feature.warnings[:19]
            feature.warnings.append("Uncovered acceptance criteria: " + ", ".join(map(str, missing)))
        gherkin = render_gherkin(feature)
        try:
            document = Parser().parse(gherkin)
            if len(document["feature"]["children"]) != len(feature.scenarios):
                raise ValueError("Scenario count mismatch")
        except Exception as exc:
            raise InvalidModelOutput("Invalid Gherkin syntax") from exc
        return GenerateResponse(
            feature=feature,
            gherkin=gherkin,
            model=self.model,
            prompt_version=PROMPT_VERSION,
            duration_ms=int((time.monotonic() - started) * 1000),
        )
