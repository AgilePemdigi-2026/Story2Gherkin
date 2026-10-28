import asyncio
import json

import pytest
from gherkin.parser import Parser
from pytest_bdd import given, parsers, scenarios, then, when

from tests.support import ENDPOINT, PRIVATE_DETAIL, PROVIDER_API_KEY, SERVICE_KEY, request_payload

pytestmark = pytest.mark.bdd
scenarios("features/generate_scenarios.feature", "features/service_reliability.feature")


@given("the AI service has a valid model response", target_fixture="case")
def service_case(harness_factory):
    return {"harness": harness_factory(), "payload": request_payload(), "key": SERVICE_KEY}


@given("the caller has valid service credentials")
def authenticated(case):
    case["key"] = SERVICE_KEY


@given("a valid user story and two acceptance criteria are provided")
def requirements(case):
    case["payload"] = request_payload()


@given(parsers.parse('the model behavior is "{behavior}"'))
def model_behavior(case, behavior):
    case["harness"].upstream.mode = behavior


@given(parsers.parse('the requested language is "{language}"'))
def language(case, language):
    case["payload"]["language"] = language


@given(parsers.parse("the requested scenario limit is {limit:d}"))
def limit(case, limit):
    case["payload"]["max_scenarios"] = limit


@given(parsers.parse('the request has the invalid field "{field}"'))
def invalid_request(case, field):
    changes = {
        "short_story": {"user_story": "short"},
        "empty_criteria": {"acceptance_criteria": []},
        "excessive_limit": {"max_scenarios": 7},
        "unknown_language": {"language": "unknown"},
        "unexpected_field": {"not_a_contract_field": True},
    }
    case["payload"].update(changes[field])


@given(parsers.parse('the caller credentials are "{credentials}"'))
def credentials(case, credentials):
    case["key"] = None if credentials == "missing" else "incorrect-test-key"


@given("the generator slot is occupied")
def busy(case):
    case["harness"].app.state.gate = asyncio.Semaphore(0)


@when("the caller requests scenario generation")
def generate(case):
    case["response"] = case["harness"].post(case["payload"], key=case["key"])


@when("the model recovers")
def recovery(case):
    case["harness"].upstream.mode = "valid"


@when("the caller checks service health")
def health(case):
    case["response"] = case["harness"].client.get("/health")


@when("the caller requests generation using GET")
def unsupported_method(case):
    case["response"] = case["harness"].client.get(ENDPOINT)


@then(parsers.parse("the HTTP status is {status:d}"))
def status(case, status):
    assert case["response"].status_code == status, case["response"].text


@then(parsers.parse("the Gherkin output contains {count:d} parseable scenarios"))
def parseable(case, count):
    data = case["response"].json()
    document = Parser().parse(data["gherkin"])
    assert len(document["feature"]["children"]) == count
    assert len(data["feature"]["scenarios"]) == count


@then("both positive and negative scenarios are returned")
def scenario_types(case):
    assert {s["kind"] for s in case["response"].json()["feature"]["scenarios"]} == {"positive", "negative"}


@then("model metadata and a prompt version are returned")
def metadata(case):
    data = case["response"].json()
    assert data["model"] == "nvidia/fixture-nemotron:free"
    assert data["prompt_version"] == "scenario-v1"
    assert isinstance(data["duration_ms"], int) and data["duration_ms"] >= 0


@then(parsers.parse("the model is called {count:d} times"))
def call_count(case, count):
    assert len(case["harness"].upstream.requests) == count


@then("no model call is made")
def no_call(case):
    assert not case["harness"].upstream.requests


@then("the response does not expose internal secrets")
def safe_error(case):
    assert PRIVATE_DETAIL not in case["response"].text
    assert SERVICE_KEY not in case["response"].text
    assert PROVIDER_API_KEY not in case["response"].text


@then("the response warns that criterion 2 is uncovered")
def warning(case):
    assert "Uncovered acceptance criteria: 2" in case["response"].json()["feature"]["warnings"]


@then(parsers.parse('the model receives the requested language "{language}"'))
def forwarded_language(case, language):
    payload = json.loads(case["harness"].upstream.requests[0].content)
    assert json.loads(payload["messages"][1]["content"])["language"] == language
