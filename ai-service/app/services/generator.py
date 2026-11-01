import asyncio
import re
import time

from gherkin.parser import Parser
from pydantic import ValidationError

from app.clients.base import InvalidModelOutput, ModelClient
from app.schemas.generate import GeneratedFeature, GenerateRequest, GenerateResponse, Scenario
from app.services.prompts import PROMPT_VERSION
MAX_ATTEMPTS = 2

KINDS = ("positive", "negative", "edge")
GUI_TAG = re.compile(r"^@gui-(\d+)$")


def extract_feature_block(text: str) -> str:
    """Ambil blok feature dari teks model: mulai dari baris Feature:
    pertama, buang prosa trailing dengan memundurkan akhir sampai lolos parser.
    """
    lines = text.splitlines()
    start = next(
        (i for i, line in enumerate(lines) if line.strip().lower().startswith("feature:")),
        None,
    )
    if start is None:
        raise InvalidModelOutput("No feature block found")
    for end in range(len(lines), start, -1):
        candidate = "\n".join(lines[start:end]).strip() + "\n"
        try:
            Parser().parse(candidate)
            return candidate
        except Exception:
            continue
    raise InvalidModelOutput("Invalid Gherkin syntax")


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
            feature = await asyncio.wait_for(
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

    async def _generate_with_feedback(self, request: GenerateRequest) -> GeneratedFeature:
        feedback: str | None = None
        for _ in range(MAX_ATTEMPTS):
            try:
                text = await self.client.generate(request, feedback)
                return self._build(text, request)
            except InvalidModelOutput as exc:
                feedback = str(exc) or "Invalid output"
                continue
            except ValidationError as exc:
                feedback = f"Invalid generated fields: {exc.errors()[0]['loc']}"
                continue
        raise InvalidModelOutput(feedback or "Invalid model output")

    def _build(self, text: str, request: GenerateRequest) -> GeneratedFeature:
        gherkin = extract_feature_block(text)
        try:
            document = Parser().parse(gherkin)
        except Exception as exc:
            raise InvalidModelOutput(f"Invalid Gherkin syntax: {exc}") from exc
        children = (document.get("feature") or {}).get("children", [])
        if not children:
            raise InvalidModelOutput("No scenarios found")
        if len(children) > 6:
            raise InvalidModelOutput("Too many scenarios")
        scenarios = [self._build_scenario(ch.get("scenario", {}), request) for ch in children]
        seen = set()
        for sc in scenarios:
            key = sc.name.casefold()
            if key in seen:
                raise InvalidModelOutput(f"Duplicate scenario name: {sc.name}")
            seen.add(key)
        return GeneratedFeature(
            feature_name=(document["feature"].get("name") or "").strip(),
            scenarios=scenarios,
            warnings=[],
        )

    def _build_scenario(self, node: dict, request: GenerateRequest) -> Scenario:
        name = (node.get("name") or "").strip()
        if not name:
            raise InvalidModelOutput("Empty scenario name")
        tags = [t.get("name", "") for t in node.get("tags", [])]
        kind = next((t[1:] for t in tags if t[1:] in KINDS), "positive")
        ids = sorted({int(m.group(1)) for t in tags if (m := GUI_TAG.match(t))})
        if any(i < 1 or i > len(request.gui_elements) for i in ids):
            raise InvalidModelOutput(f"Unknown GUI element reference in scenario: {name}")
        if not ids:
            ids = self._detect_ids(node.get("steps", []), request.gui_elements)
        if not ids:
            raise InvalidModelOutput(f"Scenario has no GUI element reference: {name}")
        given, when, then = [], [], []
        current: list | None = None
        for step in node.get("steps", []):
            text = (step.get("text") or "").strip()
            if not text:
                raise InvalidModelOutput(f"Empty step in scenario: {name}")
            ktype = step.get("keywordType", "")
            if ktype == "Context":
                current = given
            elif ktype == "Action":
                current = when
            elif ktype == "Outcome":
                current = then
            elif current is None:
                raise InvalidModelOutput(f"First step must be Given in scenario: {name}")
            current.append(text)
        if not given or not when or not then:
            raise InvalidModelOutput(f"Scenario missing Given/When/Then: {name}")
        return Scenario(
            name=name, kind=kind, gui_element_ids=ids,
            given=given[:5], when=when[:5], then=then[:5],
        )

    @staticmethod
    def _detect_ids(steps: list, elements: list[str]) -> list[int]:
        """Fallback bila model lupa tag @gui-N: cocokkan substring
        case-insensitive. Coverage otoritatif tetap dihitung backend."""
        joined = " ".join(s.get("text", "") for s in steps).casefold()
        return [i for i, e in enumerate(elements, start=1) if e.casefold() in joined]
