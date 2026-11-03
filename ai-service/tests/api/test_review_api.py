from fastapi.testclient import TestClient

from app.core.config import Settings
from app.main import create_app

KEY = "review-test-service-key-12345678901234567890"
ENDPOINT = "/internal/scenarios/validate"
VALID = "Feature: Login\n  Scenario: Success\n    Given a page\n    When I click Login button\n    Then I see a dashboard\n"


def test_review_validation_is_private_and_does_not_call_llm():
    with TestClient(create_app(Settings(service_key=KEY, ollama_model="not-used"))) as client:
        payload = {"gherkin": VALID, "gui_elements": ["Login button", "Missing field"]}
        assert client.post(ENDPOINT, json=payload).status_code == 401
        assert client.post(ENDPOINT, json=payload, headers={"X-AI-Service-Key": "wrong"}).status_code == 401
        result = client.post(ENDPOINT, json=payload, headers={"X-AI-Service-Key": KEY})
        assert result.status_code == 200
        assert result.json()["valid"] and result.json()["coverage_percent"] == 50
        invalid = client.post(ENDPOINT, json={**payload, "gherkin": ""}, headers={"X-AI-Service-Key": KEY})
        assert invalid.status_code == 200 and not invalid.json()["valid"]
        assert client.post(ENDPOINT, json={**payload, "gui_elements": []}, headers={"X-AI-Service-Key": KEY}).status_code == 422
