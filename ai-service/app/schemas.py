from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field, StringConstraints

Line = Annotated[str, StringConstraints(strip_whitespace=True, min_length=1, max_length=500, pattern=r"^[^\r\n]+$")]
Title = Annotated[str, StringConstraints(strip_whitespace=True, min_length=1, max_length=200, pattern=r"^[^\r\n]+$")]
Story = Annotated[str, StringConstraints(strip_whitespace=True, min_length=10, max_length=8000)]
CriterionID = Annotated[int, Field(strict=True, ge=1, le=20)]


class StrictModel(BaseModel):
    model_config = ConfigDict(extra="forbid")


class GenerateRequest(StrictModel):
    user_story: Story
    acceptance_criteria: list[Line] = Field(min_length=1, max_length=20)
    max_scenarios: int = Field(default=4, strict=True, ge=1, le=6)
    language: Literal["en", "id"] = "en"


class Scenario(StrictModel):
    name: Title
    kind: Literal["positive", "negative", "edge"]
    criterion_ids: list[CriterionID] = Field(min_length=1, max_length=20)
    given: list[Line] = Field(min_length=1, max_length=5)
    when: list[Line] = Field(min_length=1, max_length=5)
    then: list[Line] = Field(min_length=1, max_length=5)


class GeneratedFeature(StrictModel):
    feature_name: Title
    scenarios: list[Scenario] = Field(min_length=1, max_length=6)
    warnings: list[Line] = Field(default_factory=list, max_length=20)


class GenerateResponse(StrictModel):
    feature: GeneratedFeature
    gherkin: str
    model: str
    prompt_version: str
    duration_ms: int
