import pytest

from app.services.validation import validate_gherkin

VALID = """Feature: Delete user
  Scenario: Delete successfully
    Given the user list table is visible
    When I click the Delete button
    And I accept the Confirm dialog
    Then the account is removed
"""
ELEMENTS = ["Delete button", "Confirm dialog", "User list table", "Success message"]


def test_actual_coverage_is_75_not_mockup_80():
    report = validate_gherkin(VALID, ELEMENTS)
    assert report.valid
    assert report.coverage_percent == 75
    assert report.used_elements == ELEMENTS[:3]
    assert report.not_covered_elements == ["Success message"]
    assert all(report.checks.model_dump().values())


def test_titles_comments_and_descriptions_do_not_count():
    text = """Feature: Secret button
  Scenario: Secret button
    # Secret button
    Given the user is ready
    When the user acts
    Then the operation succeeds
"""
    report = validate_gherkin(text, ["Secret button"])
    assert report.valid and report.coverage_percent == 0


def test_normalization_ignores_quotes_punctuation_case_and_whitespace():
    text = VALID.replace("Delete button", 'DELETE   "BUTTON"!')
    report = validate_gherkin(text, ["Delete button"])
    assert report.valid and report.coverage_percent == 100


@pytest.mark.parametrize("text,fragment", [
    ("", "kosong"),
    ("# comment only", "Feature"),
    (VALID.replace("Feature: Delete user", "Feature:"), "Feature"),
    ("Feature: No scenarios\n", "Scenario"),
    (VALID.replace("Given the user list table is visible", "And the user list table is visible"), "langkah pertama"),
    (VALID.replace("Given the user list table is visible", "When the user list table is visible"), "Given"),
    (VALID.replace("When I click the Delete button", "Then I click the Delete button"), "When"),
    (VALID.replace("Then the account is removed", "Given the account is removed"), "Given harus"),
    (VALID.replace("When I click the Delete button", "When "), "deskripsi"),
    (VALID.replace("Scenario: Delete successfully", "Scenario:"), "Nama Scenario"),
    (VALID + VALID, "satu Feature"),
    (VALID.replace("And I accept", "But I accept"), "Hanya Given"),
    (VALID.replace("Scenario:", "Scenario Outline:"), "Scenario Outline"),
    (VALID.replace("  Scenario:", "  Background:\n    Given preparation\n  Scenario:"), "Background"),
])
def test_invalid_structure_has_line_and_reason(text, fragment):
    report = validate_gherkin(text, ELEMENTS)
    assert not report.valid
    assert any(fragment.lower() in item.message.lower() for item in report.issues)
    assert all(item.line >= 1 and item.column >= 1 for item in report.issues)


def test_duplicate_names_are_rejected():
    scenario = VALID[VALID.index("  Scenario:"):]
    report = validate_gherkin(VALID + scenario, ELEMENTS)
    assert not report.valid and any("unik" in item.message for item in report.issues)


def test_each_scenario_needs_all_three_steps():
    report = validate_gherkin(VALID + "\n  Scenario: Incomplete\n    Given a user\n", ELEMENTS)
    assert not report.valid and not report.checks.when and not report.checks.then


def test_written_srs_policy_allows_another_when_then_cycle():
    report = validate_gherkin(VALID + "    When I refresh\n    Then the account remains absent\n", ELEMENTS)
    assert report.valid


def test_original_invalid_trailing_text_is_not_silently_trimmed():
    report = validate_gherkin(VALID + "Feature: Another feature\n", ELEMENTS)
    assert not report.valid


def test_coverage_is_rounded_to_two_decimals():
    report = validate_gherkin(VALID, ["Delete button", "Missing one", "Missing two"])
    assert report.valid and report.coverage_percent == 33.33


def test_bare_keywords_keep_present_headings_and_explain_each_empty_step():
    text = 'Feature: User login\n\n  Scenario: Successful login\n    Given\n    When\n    Then\n'
    report = validate_gherkin(text, ELEMENTS)
    assert not report.valid
    assert report.feature_name == 'User login'
    assert report.checks.feature and report.checks.scenario
    assert not report.checks.given and not report.checks.when and not report.checks.then
    assert [(item.line, item.message.split()[0]) for item in report.issues] == [(4, 'Given'), (5, 'When'), (6, 'Then')]
    assert all('belum memiliki deskripsi langkah' in item.message for item in report.issues)
    assert not report.coverage_available
    assert report.used_elements == report.not_covered_elements == []


@pytest.mark.parametrize('keyword', ['Given', 'When', 'Then', 'And'])
def test_empty_step_has_a_friendly_reason_even_when_parser_accepts_description(keyword):
    text = VALID.replace(f'{keyword} ', f'{keyword}\n    # ', 1)
    report = validate_gherkin(text, ELEMENTS)
    assert not report.valid
    assert any(item.message.startswith(f'{keyword} belum memiliki deskripsi langkah') for item in report.issues)


def test_parser_failure_does_not_expose_internal_tokens_or_invent_zero_coverage():
    report = validate_gherkin(VALID + 'Feature: Another feature\n', ELEMENTS)
    assert not report.valid and not report.coverage_available
    assert all('expected:' not in item.message and '#EOF' not in item.message for item in report.issues)
    assert report.used_elements == report.not_covered_elements == []


def test_parseable_invalid_order_still_uses_ast_coverage():
    text = VALID.replace('When I click the Delete button', 'Then I click the Delete button')
    report = validate_gherkin(text, ELEMENTS)
    assert not report.valid and report.coverage_available
    assert report.checks.feature and report.checks.scenario and report.checks.given and report.checks.then
    assert not report.checks.when
    assert report.coverage_percent == 75


def test_real_zero_coverage_remains_available():
    report = validate_gherkin(VALID, ['Missing field'])
    assert report.valid and report.coverage_available and report.coverage_percent == 0
    assert report.not_covered_elements == ['Missing field']


def test_doc_string_bare_keywords_are_not_empty_steps():
    text = VALID + '      """\n      Given\n      When\n      Then\n      """\n'
    report = validate_gherkin(text, ELEMENTS)
    assert report.valid


def test_parser_failure_diagnostics_ignore_fake_headings_inside_doc_strings():
    text = VALID + '      """\nFeature: Fake\nScenario: Fake\nGiven fake\n      """\n    When\n'
    report = validate_gherkin(text, ELEMENTS)
    assert not report.valid
    assert report.checks.feature and report.checks.scenario
    assert report.feature_name == 'Delete user'


def test_each_scenario_needs_complete_descriptions():
    text = VALID + '\n  Scenario: Incomplete\n    Given\n    When\n    Then\n'
    report = validate_gherkin(text, ELEMENTS)
    assert report.checks.feature and report.checks.scenario
    assert not report.checks.given and not report.checks.when and not report.checks.then
