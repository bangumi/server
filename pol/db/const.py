import enum
from typing import TYPE_CHECKING, Optional

from pol.db._const import (
    Staff,
    staff_job_book,
    staff_job_game,
    staff_job_real,
    staff_job_anime,
    staff_job_music,
)


class ViewMixin:
    if TYPE_CHECKING:

        def __init__(self, v):
            pass

    @classmethod
    def to_view(cls, v: Optional[int]) -> Optional[str]:
        if not v:
            return None
        return str(cls(v))


class BloodType(ViewMixin, enum.IntEnum):
    a = 1
    b = 2
    ab = 3
    o = 4

    def __str__(self):
        try:
            return {1: "A", 2: "B", 3: "AB", 4: "O"}[self.value]
        except KeyError:
            raise ValueError(f"{self.value} is not valid blood type")


class PersonType(ViewMixin, enum.IntEnum):
    person = 1
    company = 2
    band = 3

    def __str__(self):
        try:
            return {1: "个人", 2: "公司", 3: "组合"}[self.value]
        except KeyError:
            raise ValueError(f"{self.value} is not valid person record type")


class Gender(ViewMixin, enum.IntEnum):
    male = 1
    female = 2

    def __str__(self):
        try:
            return {1: "male", 2: "female"}[self.value]
        except KeyError:
            raise ValueError(f"{self.value} is not valid gender")


class EpType(enum.IntEnum):
    normal = 0
    sp = 1
    op = 2
    ed = 3


class SubjectType(enum.IntEnum):
    book = 1
    anime = 2
    music = 3
    game = 4
    real = 6


StaffMap = {
    SubjectType.book: staff_job_book,
    SubjectType.anime: staff_job_anime,
    SubjectType.music: staff_job_music,
    SubjectType.game: staff_job_game,
    SubjectType.real: staff_job_real,
}


def get_character_rel(o: int) -> str:
    return {
        1: "主角",
        2: "配角",
        3: "客串",
    }[o]


def get_staff(o: Staff) -> str:
    v: str = o.cn or o.jp or o.en or o.rdf
    return v
