from fastapi import APIRouter

from app.schemas.review import ScenarioValidation, ValidateScenarioRequest
from app.services.validation import validate_gherkin

router = APIRouter(tags=["scenarios"])


@router.post("/internal/scenarios/validate", response_model=ScenarioValidation)
def validate(request: ValidateScenarioRequest) -> ScenarioValidation:
    # Parsing is deterministic and does not require an LLM request or its gate.
    return validate_gherkin(request.gherkin, request.gui_elements)
