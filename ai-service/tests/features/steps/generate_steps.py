import asyncio

from behave import given, then, when

from app.clients.base import InvalidModelOutput, ModelUnavailable
from app.schemas.generate import GenerateRequest
from app.services.generator import MAX_ATTEMPTS, ScenarioGenerator


def make_request():
    return GenerateRequest(
        user_story="As an admin I want to delete user data so that stale accounts are gone.",
        gui_elements=["Delete button", "Confirm dialog"],
        free_scenario="Admin opens the page and clicks delete.",
    )


VALID_TEXT = """Feature: Hapus User

  @positive @gui-1
  Scenario: Hapus user terdaftar
    Given admin ada di halaman user
    When admin menekan Delete button
    Then user hilang
"""

DUP_TEXT = """Feature: Hapus User

  @positive @gui-1
  Scenario: Sama
    Given a
    When b
    Then c

  @negative @gui-2
  Scenario: sama
    Given a
    When b
    Then c
"""

GHOST_TEXT = """Feature: Hapus User

  @edge @gui-3
  Scenario: Hantu
    Given a
    When b
    Then c
"""

PROSE_TEXT = """Berikut skenarionya:

Feature: Hapus User

  @positive @gui-1
  Scenario: Hapus user terdaftar
    Given admin ada di halaman user
    When admin menekan Delete button
    Then user hilang

Semoga membantu!"""


class FakeClient:
    def __init__(self, script):
        self.script = list(script)
        self.calls = 0
        self.feedbacks = []
        self.last_model = "test-model"

    async def generate(self, request, feedback=None):
        self.calls += 1
        self.feedbacks.append(feedback)
        item = self.script[min(self.calls - 1, len(self.script) - 1)]
        if isinstance(item, Exception):
            raise item
        return item


@given('model mengembalikan fitur valid')
def step_model_valid(context):
    context.client = FakeClient([VALID_TEXT])


@given('model mengembalikan nama skenario duplikat lalu fitur valid')
def step_dup_lalu_valid(context):
    context.client = FakeClient([DUP_TEXT, VALID_TEXT])


@given('model selalu mengembalikan nama skenario duplikat')
def step_selalu_dup(context):
    context.client = FakeClient([DUP_TEXT] * MAX_ATTEMPTS)


@given('model sedang down')
def step_model_down(context):
    context.client = FakeClient([ModelUnavailable("down")])


@given('model mengembalikan gui_element_ids tak dikenal lalu fitur valid')
def step_ghost_lalu_valid(context):
    context.client = FakeClient([GHOST_TEXT, VALID_TEXT])


@given('model mengembalikan respons terpotong lalu fitur valid')
def step_truncated_lalu_valid(context):
    context.client = FakeClient([
        InvalidModelOutput("Response truncated by token limit; return a shorter complete feature"),
        VALID_TEXT,
    ])


@given('model mengembalikan prosa plus blok fitur valid')
def step_prose_lalu_valid(context):
    context.client = FakeClient([PROSE_TEXT])


@when('saya generate skenario')
def step_generate(context):
    generator = ScenarioGenerator(context.client, "m", 30)
    try:
        context.result = asyncio.run(generator.generate(make_request()))
        context.error = None
    except Exception as exc:  # noqa: BLE001 - dicatat untuk diassert
        context.result = None
        context.error = exc


@then('hasilnya valid dengan nama fitur "{nama}"')
def step_hasil_valid_nama(context, nama):
    assert context.error is None, f"error tak terduga: {context.error}"
    assert context.result.feature.feature_name == nama


@then('hasilnya valid')
def step_hasil_valid(context):
    assert context.error is None, f"error tak terduga: {context.error}"
    assert context.result.feature.feature_name == "Hapus User"


@then('model tercatat dipakai')
def step_model_tercatat(context):
    assert context.result.model == "test-model"


@then('model dipanggil {n:d} kali')
def step_model_dipanggil(context, n):
    assert context.client.calls == n, f"dipanggil {context.client.calls} kali"


@then('feedback menyebut nama duplikat')
def step_feedback_duplikat(context):
    assert any("Duplicate scenario name" in (fb or "") for fb in context.client.feedbacks)


@then('feedback menyebut referensi tak dikenal')
def step_feedback_ghost(context):
    assert any("Unknown GUI element reference" in (fb or "") for fb in context.client.feedbacks)


@then('feedback menyebut respons terpotong')
def step_feedback_truncated(context):
    assert any("truncated" in (fb or "") for fb in context.client.feedbacks)


@then('generate gagal')
def step_generate_gagal(context):
    assert isinstance(context.error, InvalidModelOutput), f"error = {context.error!r}"


@then('generate gagal karena model down')
def step_generate_gagal_down(context):
    assert isinstance(context.error, ModelUnavailable), f"error = {context.error!r}"
