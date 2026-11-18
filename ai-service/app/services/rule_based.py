import time

from app.schemas.generate import (
    GenerateRequest,
    GenerateResponse,
    GeneratedFeature,
    Scenario,
)
from app.services.generator import render_gherkin

PROMPT_VERSION = "rule-based-v1"
ELEMENTS_PER_SCENARIO = 6
ELEMENTS_PER_STEP = 2


def _feature_name(user_story: str) -> str:
    first_line = user_story.strip().splitlines()[0].strip()
    if first_line.endswith("."):
        first_line = first_line[:-1]
    return first_line[:200] or "User story"


def _chunks(items: list[str], size: int) -> list[list[str]]:
    return [items[i:i + size] for i in range(0, len(items), size)]


def _steps_for(chunk: list[str]) -> list[str]:
    return [
        "the user uses " + " and ".join(f'"{element}"' for element in group)
        for group in _chunks(chunk, ELEMENTS_PER_STEP)
    ]


def generate_rule_based(request: GenerateRequest) -> GenerateResponse:
    started = time.monotonic()
    scenarios = []
    offset = 0
    for index, chunk in enumerate(_chunks(request.gui_elements, ELEMENTS_PER_SCENARIO), start=1):
        scenarios.append(Scenario(
            name=f"Rule-based check {index}",
            kind="positive",
            gui_element_ids=list(range(offset + 1, offset + len(chunk) + 1)),
            given=["the user is on the page described by the user story"],
            when=_steps_for(chunk),
            then=[f"the system shows the expected result: {request.free_scenario[:200]}"],
        ))
        offset += len(chunk)

    feature = GeneratedFeature(
        feature_name=_feature_name(request.user_story),
        scenarios=scenarios,
    )
    return GenerateResponse(
        feature=feature,
        gherkin=render_gherkin(feature),
        model="rule-based",
        prompt_version=PROMPT_VERSION,
        duration_ms=int((time.monotonic() - started) * 1000),
    )
