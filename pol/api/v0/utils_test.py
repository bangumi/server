import pytest

from pol.api.v0.utils import short_description


@pytest.mark.parametrize(
    ("input", "expected"),
    [
        ("=" * 100, "=" * 77 + "..."),
        ("==", "=="),
        ("=" * 79, "=" * 79),
    ],
)
def test_short_description(input: str, expected: str):
    assert short_description(input) == expected
