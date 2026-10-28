import asyncio
import time

from gherkin.parser import Parser

from app.clients.base import InvalidModelOutput, ModelClient
from app.schemas.generate import GeneratedFeature, GenerateRequest, GenerateResponse
from app.services.prompts import PROMPT_VERSION

MAX_ATTEMPTS = 2


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
        try:
            feature, feedback = await asyncio.wait_for(
                self._generate_with_feedback(request), timeout=self.timeout_seconds
            )
        except asyncio.TimeoutError as exc:
            raise TimeoutError("Scenario generation timed out") from exc
        gherkin = render_gherkin(feature)
        return GenerateResponse(
            feature=feature,
            gherkin=gherkin,
            model=getattr(self.client, "last_model", "") or self.model,
            prompt_version=PROMPT_VERSION,
            duration_ms=int((time.monotonic() - started) * 1000),
        )

    async def _generate_with_feedback(
        self, request: GenerateRequest
    ) -> tuple[GeneratedFeature, str | None]:
        feedback: str | None = None
        for _ in range(MAX_ATTEMPTS):
            try:
                feature = await self.client.generate(request, feedback)
            except InvalidModelOutput as exc:
                # Respons tak bisa dipakai (terpotong/skema rusak): jadikan
                # feedback agar percobaan berikut memperbaikinya.
                feedback = str(exc) or "Invalid output"
                continue
            try:
                self._check(feature, request)
                gherkin = render_gherkin(feature)
                document = Parser().parse(gherkin)
                if len(document["feature"]["children"]) != len(feature.scenarios):
                    raise InvalidModelOutput("Scenario count mismatch")
            except InvalidModelOutput as exc:
                feedback = str(exc) or "Invalid output"
                continue
            return feature, feedback
        raise InvalidModelOutput(feedback or "Invalid model output")

    @staticmethod
    def _check(feature: GeneratedFeature, request: GenerateRequest) -> None:
        if len(feature.scenarios) > 6:
            raise InvalidModelOutput("Too many scenarios")
        seen = set()
        for scenario in feature.scenarios:
            if scenario.name.casefold() in seen:
                raise InvalidModelOutput(f"Duplicate scenario name: {scenario.name}")
            seen.add(scenario.name.casefold())
            if any(gid > len(request.gui_elements) for gid in scenario.gui_element_ids):
                raise InvalidModelOutput("Unknown GUI element reference")
