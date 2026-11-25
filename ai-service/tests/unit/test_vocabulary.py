from app.services.vocabulary import ALLOWED_BY_PHASE, parse_step


def test_accepts_supported_steps():
    assert parse_step('I open "/login"') == ("open", ("/login",))
    assert parse_step('I fill "Email" with "<email>"') == ("fill", ("Email", "<email>"))
    assert parse_step('I click "Masuk"') == ("click", ("Masuk",))
    assert parse_step('I see "Dashboard"') == ("assert_text", ("Dashboard",))
    assert parse_step('the URL contains "/dashboard"') == ("assert_url", ("/dashboard",))


def test_rejects_free_form_steps():
    assert parse_step("the user enters a valid email") is None
    assert parse_step("I open /login") is None


def test_phase_rules():
    assert ALLOWED_BY_PHASE["given"] == {"open"}
    assert ALLOWED_BY_PHASE["when"] == {"fill", "click"}
    assert ALLOWED_BY_PHASE["then"] == {"assert_text", "assert_url"}
