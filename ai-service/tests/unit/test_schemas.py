import copy

import pytest
from pydantic import ValidationError

from app.schemas import GeneratedFeature, GenerateRequest
from tests.support import feature_payload, request_payload

pytestmark = pytest.mark.unit


def test_request_defaults_and_whitespace():
    request = GenerateRequest(user_story="  As a user, I want to log in.  ", acceptance_criteria=["  Login succeeds.  "])
    assert request.user_story == "As a user, I want to log in."
    assert request.acceptance_criteria == ["Login succeeds."]
    assert request.language == "en"
    assert request.max_scenarios == 4


@pytest.mark.parametrize("field,value", [
    ("user_story", "x" * 10), ("user_story", "x" * 8000),
    ("acceptance_criteria", ["x" * 500]), ("acceptance_criteria", ["criterion"] * 20),
    ("max_scenarios", 1), ("max_scenarios", 6), ("language", "id"),
])
def test_request_valid_boundaries(field, value):
    GenerateRequest.model_validate({**request_payload(), field: value})


@pytest.mark.parametrize("field,value", [
    ("user_story", "short"), ("user_story", " " * 20), ("user_story", "x" * 8001),
    ("acceptance_criteria", []), ("acceptance_criteria", ["x"] * 21),
    ("acceptance_criteria", [""]), ("acceptance_criteria", [" " * 3]),
    ("acceptance_criteria", ["a\nb"]), ("acceptance_criteria", ["a\rb"]),
    ("acceptance_criteria", ["x" * 501]),
    ("max_scenarios", 0), ("max_scenarios", 7), ("max_scenarios", True),
    ("max_scenarios", "4"), ("language", "fr"), ("unexpected", "value"),
])
def test_request_invalid_values(field, value):
    with pytest.raises(ValidationError):
        GenerateRequest.model_validate({**request_payload(), field: value})


@pytest.mark.parametrize("field,value", [
    ("name", ""), ("name", "name\ninjection"), ("name", "x" * 201),
    ("kind", "unknown"), ("criterion_ids", []), ("criterion_ids", [0]),
    ("criterion_ids", [21]), ("criterion_ids", [True]), ("criterion_ids", ["1"]),
    ("given", []), ("when", []), ("then", []),
    ("given", ["step"] * 6), ("when", ["x" * 501]), ("then", ["a\nScenario: injected"]),
    ("extra", "unexpected field"),
])
def test_invalid_scenario_structure(field, value):
    data = feature_payload()
    data["scenarios"][0][field] = value
    with pytest.raises(ValidationError):
        GeneratedFeature.model_validate(data)


@pytest.mark.parametrize("field,value", [
    ("feature_name", ""), ("feature_name", "x\ny"), ("feature_name", "x" * 201),
    ("scenarios", []), ("scenarios", [feature_payload()["scenarios"][0]] * 7),
    ("warnings", ["warning"] * 21), ("extra", "field"),
])
def test_invalid_feature_structure(field, value):
    with pytest.raises(ValidationError):
        GeneratedFeature.model_validate({**feature_payload(), field: copy.deepcopy(value)})
