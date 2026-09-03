"""Speech-to-text via local mlx-whisper. Audio in → text out."""

from __future__ import annotations

from pathlib import Path

DEFAULT_WHISPER_MODEL = "mlx-community/whisper-base-mlx"


def speech_to_text(
    audio_path: str | Path,
    *,
    model: str = DEFAULT_WHISPER_MODEL,
) -> str:
    """Transcribe a local audio file to plain text."""
    import mlx_whisper

    path = Path(audio_path).expanduser().resolve()
    if not path.is_file():
        raise FileNotFoundError(f"Audio file not found: {path}")

    result = mlx_whisper.transcribe(str(path), path_or_hf_repo=model)
    text = (result.get("text") or "").strip()
    return text
