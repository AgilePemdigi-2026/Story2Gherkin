from pydantic import BaseModel, ConfigDict, Field


class ValidateScenarioRequest(BaseModel):
    model_config = ConfigDict(extra="forbid")

    gherkin: str = Field(max_length=65536)
    gui_elements: list[str] = Field(min_length=1, max_length=30)


class ValidationIssue(BaseModel):
    message: str
    line: int = 1
    column: int = 1


class StructureChecks(BaseModel):
    feature: bool = False
    scenario: bool = False
    given: bool = False
    when: bool = False
    then: bool = False


class ScenarioValidation(BaseModel):
    valid: bool
    feature_name: str = ""
    checks: StructureChecks
    issues: list[ValidationIssue]
    used_elements: list[str]
    not_covered_elements: list[str]
    coverage_percent: float
