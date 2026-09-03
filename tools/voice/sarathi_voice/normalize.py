"""Speech normalization: Whisper text → clean structured text for Qwen."""

from __future__ import annotations

import re

_ONES = {
    "zero": 0,
    "one": 1,
    "two": 2,
    "three": 3,
    "four": 4,
    "five": 5,
    "six": 6,
    "seven": 7,
    "eight": 8,
    "nine": 9,
    "ten": 10,
    "eleven": 11,
    "twelve": 12,
    "thirteen": 13,
    "fourteen": 14,
    "fifteen": 15,
    "sixteen": 16,
    "seventeen": 17,
    "eighteen": 18,
    "nineteen": 19,
}
_TENS = {
    "twenty": 20,
    "thirty": 30,
    "forty": 40,
    "fifty": 50,
    "sixty": 60,
    "seventy": 70,
    "eighty": 80,
    "ninety": 90,
}

_NUMBER_WORDS = "|".join(
    sorted({*_ONES.keys(), *_TENS.keys()}, key=len, reverse=True)
)
_SPOKEN_NUMBER = re.compile(
    rf"\b(?:{_NUMBER_WORDS})(?:[\s-](?:and\s+)?(?:{_NUMBER_WORDS}))?\b",
    re.IGNORECASE,
)


def _words_to_int(phrase: str) -> int | None:
    parts = [
        p
        for p in re.split(r"[\s-]+", phrase.strip().lower())
        if p and p != "and"
    ]
    if not parts:
        return None
    if len(parts) == 1:
        if parts[0] in _ONES:
            return _ONES[parts[0]]
        if parts[0] in _TENS:
            return _TENS[parts[0]]
        return None
    if len(parts) == 2 and parts[0] in _TENS and parts[1] in _ONES:
        return _TENS[parts[0]] + _ONES[parts[1]]
    return None


def _to_digits(token: str) -> str:
    if token.isdigit():
        return token
    value = _words_to_int(token)
    return str(value) if value is not None else token


def normalize_speech(text: str) -> str:
    """Clean Whisper output into structured text for the LLM.

    Plain string in → plain string out. No embeddings or model tensors.
    """
    cleaned = re.sub(r"\s+", " ", text.strip())
    if not cleaned:
        return ""

    # "file number twenty five" / "file 25" → "file number 25"
    cleaned = re.sub(
        rf"\bfile(?:\s+number)?\s+(\d+|{_SPOKEN_NUMBER.pattern[2:-2]})\b",
        lambda m: f"file number {_to_digits(m.group(1))}",
        cleaned,
        flags=re.IGNORECASE,
    )

    # "line ten" / "line 10" → "line 10"
    cleaned = re.sub(
        rf"\bline\s+(\d+|{_SPOKEN_NUMBER.pattern[2:-2]})\b",
        lambda m: f"line {_to_digits(m.group(1))}",
        cleaned,
        flags=re.IGNORECASE,
    )

    # Remaining clear spoken numbers
    cleaned = _SPOKEN_NUMBER.sub(lambda m: _to_digits(m.group(0)), cleaned)

    cleaned = cleaned.replace(" ,", ",").replace(" .", ".")
    if cleaned and cleaned[-1] not in ".!?":
        cleaned += "."
    return cleaned[0].upper() + cleaned[1:]
