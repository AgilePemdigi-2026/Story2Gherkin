from app.schemas.generate import GenerateRequest
from app.services.rule_based import generate_rule_based
from app.services.validation import validate_gherkin


def make_request(elements: list[str]) -> GenerateRequest:
    return GenerateRequest(
        user_story="Sebagai admin, saya ingin mengelola pengguna agar data rapi.",
        gui_elements=elements,
        free_scenario="Pengguna dapat melihat daftar pengguna.",
    )


def test_covers_every_element_and_passes_validation():
    elements = [f"Tombol Aksi {i}" for i in range(1, 8)]
    response = generate_rule_based(make_request(elements))
    validation = validate_gherkin(response.gherkin, elements)

    assert response.model == "rule-based"
    assert response.prompt_version == "rule-based-v1"
    assert validation.valid is True
    assert validation.not_covered_elements == []
    assert validation.coverage_percent == 100.0


def test_is_deterministic():
    request = make_request(["Kolom Nama", "Tombol Simpan"])
    assert generate_rule_based(request).gherkin == generate_rule_based(request).gherkin


def test_handles_maximum_element_length_and_count():
    elements = [("E" * 199) + str(i % 10) for i in range(30)]
    response = generate_rule_based(make_request(elements))
    validation = validate_gherkin(response.gherkin, elements)
    assert validation.valid is True
    assert validation.coverage_percent == 100.0
    assert all(len(line) <= 500 for line in response.gherkin.splitlines())
