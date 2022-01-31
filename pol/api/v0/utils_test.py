import difflib

from pol.api.v0.utils import short_description


def test_short_description():
    for input, expected in [
        ("=" * 100, "=" * 77 + "..."),
        ("==", "=="),
        ("=" * 79, "=" * 79),
    ]:
        actual = short_description(input)
        assert actual == expected, difflib.context_diff(actual, expected)
