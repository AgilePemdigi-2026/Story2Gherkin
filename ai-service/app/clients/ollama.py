import httpx

from app.clients.base import InvalidModelOutput, ModelUnavailable
from app.core.config import Settings
from app.schemas.generate import GenerateRequest
from app.services.prompts import build_messages


class OllamaClient:
    """Transport Ollama lokal (/api/chat), mode teks Gherkin.

    Tanpa API key, tanpa kuota, tanpa flaky jaringan: deterministik penuh
    untuk demo dan pengembangan. Kontrak keluarannya (str) identik dengan
    OpenRouterClient sehingga ScenarioGenerator tidak tahu bedanya.
    """

    def __init__(self, client: httpx.AsyncClient, settings: Settings):
        self.client = client
        self.settings = settings
        self.last_model: str = settings.ollama_model
        self.last_usage: dict = {}

    async def generate(
        self, request: GenerateRequest, feedback: str | None = None
    ) -> str:
        payload = {
            "model": self.settings.ollama_model,
            "messages": build_messages(request, feedback),
            "stream": False,
            "options": {
                "temperature": 0,
                "num_predict": self.settings.ollama_num_predict,
                "num_ctx": self.settings.ollama_num_ctx,
            },
        }
        try:
            response = await self.client.post(
                self.settings.ollama_url + "/api/chat", json=payload
            )
        except httpx.TimeoutException:
            raise
        except httpx.HTTPError as exc:
            raise ModelUnavailable() from exc

        if response.status_code != 200:
            raise ModelUnavailable(f"Ollama error {response.status_code}")
        try:
            payload = response.json()
            content = payload["message"]["content"]
            if not isinstance(content, str) or not content.strip():
                raise InvalidModelOutput("Empty model response")
            if payload.get("done") is not True:
                raise InvalidModelOutput("Ollama response was incomplete")
            self.last_model = str(payload.get("model") or self.settings.ollama_model)
            return content
        except (ValueError, KeyError, TypeError) as exc:
            raise InvalidModelOutput("Unreadable model response") from exc
