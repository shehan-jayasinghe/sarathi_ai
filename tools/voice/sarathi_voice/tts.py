"""Text-to-speech via local Kokoro. Text in → audio file out.

Keeps a warm KPipeline so repeated sentence TTS stays fast.
"""

from __future__ import annotations

from pathlib import Path
from threading import Lock

import numpy as np
import soundfile as sf

DEFAULT_VOICE = "af_heart"
DEFAULT_SAMPLE_RATE = 24000

_pipeline = None
_pipeline_lock = Lock()
_pipeline_lang: str | None = None


def _get_pipeline(lang_code: str = "a"):
    global _pipeline, _pipeline_lang
    with _pipeline_lock:
        if _pipeline is None or _pipeline_lang != lang_code:
            from kokoro import KPipeline

            _pipeline = KPipeline(lang_code=lang_code, repo_id="hexgrad/Kokoro-82M")
            _pipeline_lang = lang_code
        return _pipeline


def text_to_speech(
    text: str,
    output_path: str | Path,
    *,
    voice: str = DEFAULT_VOICE,
    lang_code: str = "a",
    sample_rate: int = DEFAULT_SAMPLE_RATE,
) -> Path:
    """Synthesize speech from plain text and write a WAV file."""
    cleaned = text.strip()
    if not cleaned:
        raise ValueError("Cannot synthesize empty text")

    out = Path(output_path).expanduser().resolve()
    out.parent.mkdir(parents=True, exist_ok=True)

    pipeline = _get_pipeline(lang_code)
    chunks: list[np.ndarray] = []
    for _gs, _ps, audio in pipeline(cleaned, voice=voice):
        chunks.append(np.asarray(audio, dtype=np.float32))

    if not chunks:
        raise RuntimeError("Kokoro produced no audio")

    waveform = np.concatenate(chunks)
    sf.write(out, waveform, sample_rate)
    return out
