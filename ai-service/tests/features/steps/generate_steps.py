import asyncio

from behave import given, then, when

from app.clients.base import InvalidModelOutput, ModelUnavailable
from app.schemas.generate import GenerateRequest, GeneratedFeature
from app.services.generator import MAX_ATTEMPTS, ScenarioGenerator


def make_request():
    return GenerateRequest(
        user_story="As an admin I want to delete user data so that stale accounts are gone.",
        gui_elements=["Delete button", "Confirm dialog"],
        free_scenario="Admin opens the page and clicks delete.",
    )


def valid_feature():
    return GeneratedFeature(
        feature_name="Hapus User",
        scenarios=[
            {
                "name": "Hapus user terdaftar",
                "kind": "positive",
                "gui_element_ids": [1],
                "given": ["admin ada di halaman user"],
                "when": ["admin menekan Delete button"],
                "then": ["user hilang"],
            }
        ],
        warnings=[],
    )


def dup_feature():
    return GeneratedFeature(
        feature_name="Hapus User",
        scenarios=[
            {
                "name": "Sama", "kind": "positive", "gui_element_ids": [1],
                "given": ["a"], "when": ["b"], "then": ["c"],
            },
            {
                "name": "sama", "kind": "negative", "gui_element_ids": [2],
                "given": ["a"], "when": ["b"], "then": ["c"],
            },
        ],
        warnings=[],
    )


def ghost_feature():
    return GeneratedFeature(
        feature_name="Hapus User",
        scenarios=[
            {
                "name": "Hantu", "kind": "edge", "gui_element_ids": [3],
                "given": ["a"], "when": ["b"], "then": ["c"],
            }
        ],
        warnings= [],
    )


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
    context.client = FakeClient([valid_feature()])


@given('model mengembalikan nama skenario duplikat lalu fitur valid')
def step_dup_lalu_valid(context):
    context.client = FakeClient([dup_feature(), valid_feature()])


@given('model selalu mengembalikan nama skenario duplikat')
def step_selalu_dup(context):
    context.client = FakeClient([dup_feature()] * MAX_ATTEMPTS)


@given('model sedang down')
def step_model_down(context):
    context.client = FakeClient([ModelUnavailable("down")])


@given('model mengembalikan gui_element_ids tak dikenal lalu fitur valid')
def step_ghost_lalu_valid(context):
    context.client = FakeClient([ghost_feature(), valid_feature()])


@given('model mengembalikan respons terpotong lalu fitur valid')
def step_truncated_lalu_valid(context):
    context.client = FakeClient([
        InvalidModelOutput("Response truncated by token limit; return a shorter complete JSON"),
        valid_feature(),
    ])


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
