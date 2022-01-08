from typing import Optional

import pytest

from pol import permission


@pytest.mark.parametrize(
    ["input", "expected"],
    [
        ("1", 1),
        ("-1", None),
        ("021", None),
        ("username", None),
        ("0x021", None),
    ],
)
def test_is_user_id(input: str, expected: Optional[int]):
    assert permission.is_user_id(input) == expected
