"""End-to-end Sarathi voice pipeline.

Audio → Whisper → Speech Normalizer → Qwen → Kokoro → Audio

Each stage exchanges plain text (or audio files) only — no cross-model vectors.
"""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from sarathi_voice.llm import chat_qwen
from sarathi_voice.normalize import normalize_speech
from sarathi_voice.stt import speech_to_text
from sarathi_voice.tts import text_to_speech


@dataclass
class PipelineResult:
    raw_transcript: str
    normalized_text: str
    reply_text: str
    output_audio: Path | None


def run_pipeline(
    *,
    audio_path: str | Path | None = None,
    text: str | None = None,
    output_audio: str | Path | None = None,
    skip_tts: bool = False,
    whisper_model: str = "mlx-community/whisper-base-mlx",
    ollama_model: str = "qwen2.5:3b",
    voice: str = "af_heart",
) -> PipelineResult:
    """Run STT → normalize → Qwen → TTS.

    Provide either ``audio_path`` (Whisper) or ``text`` (skip STT).
    """
    if bool(audio_path) == bool(text):
        raise ValueError("Provide exactly one of audio_path or text")

    if audio_path:
        raw = speech_to_text(audio_path, model=whisper_model)
    else:
        raw = (text or "").strip()

    if not raw:
        raise RuntimeError("No speech/text to process")

    normalized = normalize_speech(raw)
    reply = chat_qwen(normalized, model=ollama_model)

    audio_out: Path | None = None
    if not skip_tts:
        target = Path(output_audio) if output_audio else Path("data/voice/reply.wav")
        audio_out = text_to_speech(reply, target, voice=voice)

    return PipelineResult(
        raw_transcript=raw,
        normalized_text=normalized,
        reply_text=reply,
        output_audio=audio_out,
    )
