#!/usr/bin/env python3
"""Sarathi Voice MCP server (stdio).

Exposes the local speech pipeline as MCP tools:

  Audio → Whisper → Normalizer → Qwen → Kokoro → Audio

Plain text/audio paths only between stages — no cross-model vectors.

Run (from repo):
  source tools/voice/.venv/bin/activate
  python tools/voice/mcp_server.py

Cursor MCP config example:
  {
    "mcpServers": {
      "sarathi-voice": {
        "command": "/ABS/PATH/tools/voice/.venv/bin/python",
        "args": ["/ABS/PATH/tools/voice/mcp_server.py"]
      }
    }
  }
"""

from __future__ import annotations

import sys
from pathlib import Path

# Ensure local package imports work when launched via absolute path.
VOICE_ROOT = Path(__file__).resolve().parent
REPO_ROOT = VOICE_ROOT.parents[1]
if str(VOICE_ROOT) not in sys.path:
    sys.path.insert(0, str(VOICE_ROOT))

from mcp.server.fastmcp import FastMCP

from sarathi_voice.llm import chat_qwen
from sarathi_voice.normalize import normalize_speech
from sarathi_voice.pipeline import run_pipeline
from sarathi_voice.stt import speech_to_text
from sarathi_voice.tts import text_to_speech

mcp = FastMCP(
    "sarathi-voice",
    instructions=(
        "Local Sarathi voice tools. Use plain text between stages. "
        "Pipeline: Whisper STT → speech normalizer → Qwen (Ollama) → Kokoro TTS."
    ),
)

DEFAULT_REPLY_WAV = REPO_ROOT / "data" / "voice" / "reply.wav"


@mcp.tool()
def speech_to_text_tool(audio_path: str) -> str:
    """Transcribe a local audio file to plain text with Whisper.

    Args:
        audio_path: Absolute or relative path to a WAV/audio file.
    """
    return speech_to_text(audio_path)


@mcp.tool()
def normalize_speech_tool(text: str) -> str:
    """Normalize Whisper transcript text for Qwen (numbers, cleanup).

    Args:
        text: Raw speech transcript.
    """
    return normalize_speech(text)


@mcp.tool()
def chat_qwen_tool(text: str, model: str = "qwen2.5:3b") -> str:
    """Send plain text to local Qwen via Ollama and return the reply text.

    Args:
        text: User message (preferably already normalized).
        model: Ollama model name (default qwen2.5:3b).
    """
    return chat_qwen(text, model=model)


@mcp.tool()
def text_to_speech_tool(
    text: str,
    output_path: str = "",
    voice: str = "af_heart",
) -> str:
    """Synthesize speech with Kokoro and write a WAV file. Returns the path.

    Args:
        text: Reply text to speak.
        output_path: Optional WAV output path (default data/voice/reply.wav).
        voice: Kokoro voice id.
    """
    target = Path(output_path) if output_path else DEFAULT_REPLY_WAV
    return str(text_to_speech(text, target, voice=voice))


@mcp.tool()
def voice_pipeline(
    text: str = "",
    audio_path: str = "",
    output_path: str = "",
    skip_tts: bool = False,
    ollama_model: str = "qwen2.5:3b",
    voice: str = "af_heart",
) -> dict:
    """Run full pipeline: STT (optional) → normalize → Qwen → Kokoro (optional).

    Provide exactly one of text or audio_path.

    Args:
        text: Skip STT; feed this text into normalizer + Qwen.
        audio_path: Audio file for Whisper STT.
        output_path: Optional Kokoro WAV path.
        skip_tts: If true, return text only (no Kokoro).
        ollama_model: Ollama model name.
        voice: Kokoro voice id.
    """
    if bool(text.strip()) == bool(audio_path.strip()):
        raise ValueError("Provide exactly one of text or audio_path")

    result = run_pipeline(
        audio_path=audio_path or None,
        text=text or None,
        output_audio=output_path or DEFAULT_REPLY_WAV,
        skip_tts=skip_tts,
        ollama_model=ollama_model,
        voice=voice,
    )
    return {
        "raw_transcript": result.raw_transcript,
        "normalized_text": result.normalized_text,
        "reply_text": result.reply_text,
        "output_audio": str(result.output_audio) if result.output_audio else None,
    }


if __name__ == "__main__":
    mcp.run(transport="stdio")
