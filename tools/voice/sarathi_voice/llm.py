"""Qwen chat via local Ollama. Text in → text out."""

from __future__ import annotations

import json
import urllib.error
import urllib.request

DEFAULT_OLLAMA_URL = "http://localhost:11434"
DEFAULT_MODEL = "qwen2.5:3b"

SYSTEM_PROMPT = (
    "You are Sarathi, a local desktop voice assistant. "
    "You can listen (Whisper STT), clean speech transcripts, chat, "
    "speak replies (Kokoro TTS), and run the full voice pipeline end-to-end. "
    "You run offline via Ollama; you cannot browse the web or use cloud APIs "
    "unless the user connects them. "
    "When asked what you can do, briefly list these capabilities. "
    "Reply in clear spoken-friendly English, 1-3 short sentences unless code is required. "
    "Do not mention being an AI model unless asked."
)


def chat_qwen(
    user_text: str,
    *,
    model: str = DEFAULT_MODEL,
    base_url: str = DEFAULT_OLLAMA_URL,
    system: str = SYSTEM_PROMPT,
) -> str:
    """Send plain text to local Qwen and return plain text."""
    payload = {
        "model": model,
        "stream": False,
        "messages": [
            {"role": "system", "content": system},
            {"role": "user", "content": user_text},
        ],
    }
    req = urllib.request.Request(
        f"{base_url.rstrip('/')}/api/chat",
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=120) as resp:
            body = json.loads(resp.read().decode("utf-8"))
    except urllib.error.URLError as exc:
        raise RuntimeError(
            f"Cannot reach Ollama at {base_url}. Is it running?"
        ) from exc

    message = body.get("message") or {}
    text = (message.get("content") or "").strip()
    if not text:
        raise RuntimeError(f"Empty reply from Ollama model {model}")
    return text
