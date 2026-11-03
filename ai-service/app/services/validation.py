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


def inspect_incomplete_source(text: str) -> tuple[StructureChecks, str, list[ValidationIssue]]:
    """Locate recognizable headings/steps for diagnostics when parsing fails.

    This never makes a document valid and never supplies coverage. Only the
    original document's successful Cucumber AST can do that.
    """
    feature_names: list[str] = []
    scenarios: list[dict[str, list[bool]]] = []
    scenario_names: list[str] = []
    issues: list[ValidationIssue] = []
    doc_string = ""
    for number, line in enumerate(text.splitlines(), 1):
        stripped = line.strip()
        if doc_string:
            if stripped == doc_string:
                doc_string = ""
            continue
        if stripped.startswith(('"""', '```')):
            doc_string = stripped[:3]
            continue
        if not stripped or stripped.startswith(('#', '@', '|')):
            continue
        if stripped.startswith('Feature:'):
            feature_names.append(stripped[len('Feature:'):].strip())
            continue
        if stripped.startswith('Scenario:') and feature_names:
            scenario_names.append(stripped[len('Scenario:'):].strip())
            scenarios.append({})
            continue
        match = re.fullmatch(r'(Given|When|Then|And)(?:\s+(.*))?', stripped)
        if not match or not scenarios:
            continue
        keyword, description = match.groups()
        complete = bool(description and description.strip())
        scenarios[-1].setdefault(keyword, []).append(complete)
        if not complete:
            issues.append(ValidationIssue(
                line=number, column=len(line) - len(line.lstrip()) + 1,
                message=f'{keyword} belum memiliki deskripsi langkah. Tambahkan penjelasan setelah {keyword}.',
            ))
    checks = StructureChecks(
        feature=len(feature_names) == 1 and bool(feature_names[0]),
        scenario=bool(scenario_names) and all(scenario_names)
        and len(set(name.casefold() for name in scenario_names)) == len(scenario_names),
        **{keyword.lower(): bool(scenarios) and all(
            bool(steps.get(keyword)) and all(steps[keyword]) for steps in scenarios
        ) for keyword in ('Given', 'When', 'Then')},
    )
    return checks, feature_names[0] if len(feature_names) == 1 else '', issues


def validate_gherkin(text: str, elements: list[str]) -> ScenarioValidation:
    source_checks, source_feature_name, source_issues = inspect_incomplete_source(text)
    issues: list[ValidationIssue] = list(source_issues)
    checks = StructureChecks()
    feature_name = ""
    step_texts: list[str] = []

    def issue(message: str, node: dict | None = None) -> None:
        location = (node or {}).get("location", {})
        item = ValidationIssue(
            message=message, line=location.get("line", 1),
            column=location.get("column", 1),
        )
        if item not in issues:
            issues.append(item)

    document = None
    if not text.strip():
        issue("Teks Gherkin tidak boleh kosong.")
    else:
        try:
            document = Parser().parse(text)
        except Exception as error:
            checks, feature_name = source_checks, source_feature_name
            # Keep Cucumber locations, but never expose its internal token names.
            for detail in getattr(error, "errors", [error]):
                message = str(detail)
                match = re.search(r"\((\d+):(\d+)\):\s*(.*)", message, re.S)
                if match:
                    line, column = int(match[1]), int(match[2])
                    if any(item.line == line for item in source_issues):
                        continue
                    source_lines = text.splitlines()
                    current = source_lines[line - 1].strip() if line <= len(source_lines) else ''
                    if current.startswith('Feature:'):
                        friendly = 'Hanya satu Feature yang diperbolehkan dalam satu skenario review.'
                    else:
                        friendly = 'Sintaks pada baris ini belum sesuai. Periksa penulisan Feature:, Scenario:, dan langkah Given, When, Then, atau And.'
                    issues.append(ValidationIssue(
                        line=line, column=column, message=friendly,
                    ))
                else:
                    issue('Sintaks Gherkin belum sesuai. Periksa judul dan penulisan setiap langkah.')

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
            for index, scenario in enumerate(scenarios):
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
                complete_steps: dict[str, list[bool]] = {}
                for step in scenario.get("steps", []):
                    keyword = step.get("keyword", "").strip()
                    description = step.get("text", "").strip()
                    # Coverage only inspects real AST step text, never titles,
                    # comments, descriptions, tags, doc strings, or tables.
                    step_texts.append(description)
                    if not description:
                        issue(f'{keyword} belum memiliki deskripsi langkah. Tambahkan penjelasan setelah {keyword}.', step)
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
                    complete_steps.setdefault(keyword, []).append(bool(description))
                presence.append({keyword for keyword, complete in complete_steps.items() if all(complete)})
                start_line = scenario.get('location', {}).get('line', 1)
                next_scenario = scenarios[index + 1] if index + 1 < len(scenarios) else {}
                end_line = next_scenario.get('location', {}).get('line', len(text.splitlines()) + 1)
                for required in ("Given", "When", "Then"):
                    empty_step = any(
                        start_line < item.line < end_line
                        and item.message.startswith(f'{required} belum memiliki')
                        for item in source_issues
                    )
                    if required not in found and not empty_step:
                        issue(f'Scenario "{name}" wajib memiliki {required}.', scenario)
            checks.given = bool(presence) and all("Given" in item for item in presence)
            checks.when = bool(presence) and all("When" in item for item in presence)
            checks.then = bool(presence) and all("Then" in item for item in presence)

    for item in source_issues:
        keyword = item.message.split()[0].lower()
        if keyword in {'given', 'when', 'then'}:
            setattr(checks, keyword, False)
    normalized_steps = [normalize_gui_text(step) for step in step_texts]
    used, uncovered = [], []
    coverage_available = document is not None and bool(document.get('feature')) and not source_issues
    for element in elements if coverage_available else []:
        normalized = normalize_gui_text(element)
        target = used if normalized and any(normalized in step for step in normalized_steps) else uncovered
        target.append(element)
    return ScenarioValidation(
        valid=not issues, feature_name=feature_name, checks=checks, issues=issues,
        used_elements=used, not_covered_elements=uncovered,
        coverage_percent=round(len(used) / len(elements) * 100, 2) if elements else 0,
        coverage_available=coverage_available,
    )
