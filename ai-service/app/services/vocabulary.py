"""Fixed step vocabulary shared by the generator and the end-to-end runner.

Gherkin steps must match one of these forms so the backend can execute them
without a manual mapping step.
"""

import re

STEP_PATTERNS = (
    ("open", re.compile(r'^I open "([^"]+)"$')),
    ("fill", re.compile(r'^I fill "([^"]+)" with "([^"]*)"$')),
    ("click", re.compile(r'^I click "([^"]+)"$')),
    ("assert_text", re.compile(r'^I see "([^"]+)"$')),
    ("assert_url", re.compile(r'^the URL contains "([^"]+)"$')),
)

ALLOWED_BY_PHASE = {
    "given": {"open"},
    "when": {"fill", "click"},
    "then": {"assert_text", "assert_url"},
}


def parse_step(text: str) -> tuple[str, tuple[str, ...]] | None:
    """Return (action, captured groups) when the step is in the vocabulary."""
    for action, pattern in STEP_PATTERNS:
        match = pattern.match(text.strip())
        if match:
            return action, match.groups()
    return None
