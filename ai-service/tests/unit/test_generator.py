import asyncio
from types import SimpleNamespace

import pytest
from gherkin.parser import Parser

from app import generator as module
from app.generator import ScenarioGenerator, render_gherkin
from app.llm_client import InvalidModelOutput, ModelUnavailable
from app.schemas import GeneratedFeature, GenerateRequest
from tests.support import feature_payload, request_payload

pytestmark = pytest.mark.unit


class FakeModel:
    def __init__(self, data=None, error=None):
        self.data = data if data is not None else feature_payload()
        self.error = error
        self.calls = 0
        self.request = None

    async def generate(self, request):
        self.calls += 1
        self.request = request
        if self.error:
            raise self.error
        return GeneratedFeature.model_validate(self.data)


def generate(model=None, request=None):
    model = model or FakeModel()
    request = request or GenerateRequest.model_validate(request_payload())
    return asyncio.run(ScenarioGenerator(model, "test-model").generate(request))


def test_generate_returns_parseable_gherkin_and_metadata(monkeypatch):
    ticks = iter([10.0, 10.125])
    monkeypatch.setattr(module, "time", SimpleNamespace(monotonic=lambda: next(ticks)))
    model = FakeModel()
    request = GenerateRequest.model_validate(request_payload())
    response = generate(model, request)
    parsed = Parser().parse(response.gherkin)
    assert len(parsed["feature"]["children"]) == 2
    assert response.model == "test-model" and response.prompt_version == "scenario-v1"
    assert response.duration_ms == 125
    assert response.feature.warnings == []
    assert model.request is request and model.calls == 1


def test_renderer_uses_and_for_additional_steps():
    data = feature_payload()
    scenario = data["scenarios"][0]
    scenario["given"].append("the login page is open")
    scenario["when"].append("the login button is pressed")
    scenario["then"].append("the dashboard is visible")
    text = render_gherkin(GeneratedFeature.model_validate(data))
    assert "    And the login page is open\n" in text
    assert "    And the login button is pressed\n" in text
    assert "    And the dashboard is visible\n" in text
    assert text.endswith("\n")
    assert len(Parser().parse(text)["feature"]["children"]) == 2


@pytest.mark.parametrize("fault", ["too_many", "duplicate_name", "unknown_reference"])
def test_semantically_invalid_structure_rejected(fault):
    data = feature_payload()
    request = GenerateRequest.model_validate(request_payload())
    if fault == "too_many":
        request.max_scenarios = 1
    elif fault == "duplicate_name":
        data["scenarios"][1]["name"] = data["scenarios"][0]["name"].upper()
    else:
        data["scenarios"][0]["criterion_ids"] = [3]
    with pytest.raises(InvalidModelOutput):
        generate(FakeModel(data), request)


def test_uncovered_criteria_warning_preserves_existing_warnings():
    data = feature_payload()
    data["scenarios"] = data["scenarios"][:1]
    data["warnings"] = ["A product decision still needs clarification."]
    response = generate(FakeModel(data))
    assert response.feature.warnings == [data["warnings"][0], "Uncovered acceptance criteria: 2"]


def test_uncovered_warning_stays_within_schema_limit():
    data = feature_payload()
    data["scenarios"] = data["scenarios"][:1]
    data["warnings"] = [f"Existing warning {i}" for i in range(20)]
    response = generate(FakeModel(data))
    assert len(response.feature.warnings) == 20
    assert response.feature.warnings[-1] == "Uncovered acceptance criteria: 2"


def test_parser_failure_is_translated(monkeypatch):
    def fail(*args, **kwargs):
        raise ValueError("private parser details")
    monkeypatch.setattr(module.Parser, "parse", fail)
    with pytest.raises(InvalidModelOutput, match="Invalid Gherkin syntax"):
        generate()


def test_parser_count_mismatch_is_rejected(monkeypatch):
    monkeypatch.setattr(module.Parser, "parse", lambda *args, **kwargs: {"feature": {"children": []}})
    with pytest.raises(InvalidModelOutput):
        generate()


def test_overall_deadline_includes_inference():
    class BlockedModel:
        async def generate(self, request):
            await asyncio.Event().wait()
    request = GenerateRequest.model_validate(request_payload())
    with pytest.raises(TimeoutError):
        asyncio.run(ScenarioGenerator(BlockedModel(), "fake", 0.01).generate(request))


def test_cancellation_propagates():
    with pytest.raises(asyncio.CancelledError):
        generate(FakeModel(error=asyncio.CancelledError()))


def test_upstream_error_is_not_retried_in_python():
    model = FakeModel(error=ModelUnavailable())
    with pytest.raises(ModelUnavailable):
        generate(model)
    assert model.calls == 1
