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


class BloodType(int, ViewMixin, enum.Enum):
    a = 1
    b = 2
    ab = 3
    o = 4

    def __str__(self):
        try:
            return {1: "A", 2: "B", 3: "AB", 4: "O"}[self.value]
        except KeyError:
            raise ValueError(f"{self.value} is not valid blood type")


class PersonType(int, ViewMixin, enum.Enum):
    person = 1
    company = 2
    band = 3

    def __str__(self):
        try:
            return {1: "个人", 2: "公司", 3: "组合"}[self.value]
        except KeyError:
            raise ValueError(f"{self.value} is not valid person record type")


class Gender(int, ViewMixin, enum.Enum):
    male = 1
    female = 2

    def __str__(self):
        try:
            return {1: "male", 2: "female"}[self.value]
        except KeyError:
            raise ValueError(f"{self.value} is not valid gender")


class SubjectType(int, enum.Enum):
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


def get_staff(o: Staff) -> str:
    v: str = o.cn or o.jp or o.en or o.rdf
    return v
