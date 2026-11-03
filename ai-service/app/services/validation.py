"""Validate the original text, never a repaired or truncated model response.

SRS 6.2 currently allows repeated When/Then cycles and lists the six English
keywords below. Approval, review history, and other TBD policies are not added.
"""

import re
import unicodedata

from gherkin.parser import Parser

from app.schemas.review import ScenarioValidation, StructureChecks, ValidationIssue


def normalize_gui_text(text: str) -> str:
    text = "".join(
        char for char in text.lower()
        if unicodedata.category(char)[0] not in {"P", "S"}
    )
    return " ".join(text.split())


def validate_gherkin(text: str, elements: list[str]) -> ScenarioValidation:
    issues: list[ValidationIssue] = []
    checks = StructureChecks()
    feature_name = ""
    step_texts: list[str] = []

    def issue(message: str, node: dict | None = None) -> None:
        location = (node or {}).get("location", {})
        issues.append(ValidationIssue(
            message=message, line=location.get("line", 1),
            column=location.get("column", 1),
        ))

    document = None
    if not text.strip():
        issue("Teks Gherkin tidak boleh kosong.")
    else:
        try:
            document = Parser().parse(text)
        except Exception as error:
            # Cucumber reports each syntax error as '(line:column): ...'.
            for detail in getattr(error, "errors", [error]):
                message = str(detail)
                match = re.search(r"\((\d+):(\d+)\):\s*(.*)", message, re.S)
                if match:
                    issues.append(ValidationIssue(
                        line=int(match[1]), column=int(match[2]), message=match[3],
                    ))
                else:
                    issue("Sintaks Gherkin tidak valid: " + message)

    if document is not None:
        feature = document.get("feature")
        if not feature:
            issue("Harus terdapat tepat satu Feature dengan nama tidak kosong.")
        else:
            feature_name = feature.get("name", "").strip()
            checks.feature = bool(feature_name) and feature.get("keyword") == "Feature"
            if not checks.feature:
                issue("Gunakan satu Feature berbahasa Inggris dengan nama tidak kosong.", feature)

            scenarios: list[dict] = []
            for child in feature.get("children", []):
                if "scenario" not in child:
                    node = next(iter(child.values()), feature)
                    issue("Hanya Scenario yang didukung; Background dan Rule belum didukung.", node)
                else:
                    scenarios.append(child["scenario"])
            checks.scenario = bool(scenarios)
            if not scenarios:
                issue("Feature harus memiliki minimal satu Scenario.", feature)

            names: set[str] = set()
            presence: list[set[str]] = []
            for scenario in scenarios:
                name = scenario.get("name", "").strip()
                if scenario.get("keyword") != "Scenario" or scenario.get("examples"):
                    issue("Gunakan Scenario; Scenario Outline dan Examples belum didukung.", scenario)
                if not name:
                    issue("Nama Scenario tidak boleh kosong.", scenario)
                    checks.scenario = False
                key = name.casefold()
                if key in names:
                    issue("Nama Scenario harus unik: " + name, scenario)
                    checks.scenario = False
                names.add(key)

                phase = ""
                found: set[str] = set()
                for step in scenario.get("steps", []):
                    keyword = step.get("keyword", "").strip()
                    description = step.get("text", "").strip()
                    # Coverage only inspects real AST step text, never titles,
                    # comments, descriptions, tags, doc strings, or tables.
                    step_texts.append(description)
                    if not description:
                        issue("Setiap langkah wajib memiliki deskripsi.", step)
                    if keyword not in {"Given", "When", "Then", "And"}:
                        issue("Hanya Given, When, Then, dan And yang didukung.", step)
                        continue
                    if keyword == "And":
                        if not phase:
                            issue("And tidak boleh menjadi langkah pertama; mulai dengan Given.", step)
                        continue
                    if keyword == "Given" and phase not in {"", "Given"}:
                        issue("Given harus muncul sebelum When dan Then.", step)
                    elif keyword == "When" and not phase:
                        issue("When harus didahului Given.", step)
                    elif keyword == "Then" and phase not in {"When", "Then"}:
                        issue("Then harus didahului When.", step)
                    phase = keyword
                    found.add(keyword)
                presence.append(found)
                for required in ("Given", "When", "Then"):
                    if required not in found:
                        issue(f'Scenario "{name}" wajib memiliki {required}.', scenario)
            checks.given = bool(presence) and all("Given" in item for item in presence)
            checks.when = bool(presence) and all("When" in item for item in presence)
            checks.then = bool(presence) and all("Then" in item for item in presence)

    normalized_steps = [normalize_gui_text(step) for step in step_texts]
    used, uncovered = [], []
    for element in elements:
        normalized = normalize_gui_text(element)
        target = used if normalized and any(normalized in step for step in normalized_steps) else uncovered
        target.append(element)
    return ScenarioValidation(
        valid=not issues, feature_name=feature_name, checks=checks, issues=issues,
        used_elements=used, not_covered_elements=uncovered,
        coverage_percent=round(len(used) / len(elements) * 100, 2) if elements else 0,
    )
