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


RELATION_SERIES = {
    1: {"en": "Adaptation", "cn": "改编", "jp": "", "desc": "同系列不同平台作品，如柯南漫画与动画版"},
    2: {"en": "Prequel", "cn": "前传", "jp": "", "desc": "发生在故事之前"},
    3: {"en": "Sequel", "cn": "续集", "jp": "", "desc": "发生在故事之后"},
    4: {"en": "Summary", "cn": "总集篇", "jp": "", "desc": "对故事的概括版本"},
    5: {"en": "Full Story", "cn": "全集", "jp": "", "desc": "相对于剧场版/总集篇的完整故事"},
    6: {"en": "Side Story", "cn": "番外篇", "jp": "", "desc": ""},
    7: {"en": "Character", "cn": "角色出演", "jp": "", "desc": "相同角色，没有关联的故事"},
    8: {
        "en": "Same setting",
        "cn": "相同世界观",
        "jp": "",
        "desc": "发生在同一个世界观/时间线下，不同的出演角色",
    },
    9: {
        "en": "Alternative setting",
        "cn": "不同世界观",
        "jp": "",
        "desc": "相同的出演角色，不同的世界观/时间线设定",
    },
    10: {
        "en": "Alternative version",
        "cn": "不同演绎",
        "jp": "",
        "desc": "相同设定、角色，不同的演绎方式（如EVA原作与新剧场版)",
    },
    11: {"en": "Spin-off", "cn": "衍生", "jp": "", "desc": "如柯南与魔术快斗"},
    12: {"en": "Parent Story", "cn": "主线故事", "jp": "", "desc": ""},
    99: {"en": "Other", "cn": "其他", "jp": "", "desc": ""},
}

RELATION_BOOK = {
    1: {"en": "Adaptation", "cn": "改编", "jp": "", "desc": "同系列不同平台作品，如柯南漫画与动画版"},
    1003: {"en": "Offprint", "cn": "单行本", "jp": "", "desc": ""},
    1002: {"en": "Series", "cn": "系列", "jp": "", "desc": ""},
    1004: {"en": "Album", "cn": "画集", "jp": "", "desc": ""},
    1005: {"en": "Prequel", "cn": "前传", "jp": "", "desc": "发生在故事之前"},
    1006: {"en": "Sequel", "cn": "续集", "jp": "", "desc": "发生在故事之后"},
    1007: {"en": "Side Story", "cn": "番外篇", "jp": "", "desc": ""},
    1008: {"en": "Parent Story", "cn": "主线故事", "jp": "", "desc": ""},
    1010: {"en": "Alternative version", "cn": "不同版本", "jp": "", "desc": ""},
    1011: {"en": "Character", "cn": "角色出演", "jp": "", "desc": "相同角色，没有关联的故事"},
    1012: {
        "en": "Same setting",
        "cn": "相同世界观",
        "jp": "",
        "desc": "发生在同一个世界观/时间线下，不同的出演角色",
    },
    1013: {
        "en": "Alternative setting",
        "cn": "不同世界观",
        "jp": "",
        "desc": "相同的出演角色，不同的世界观/时间线设定",
    },
    1099: {"en": "Other", "cn": "其他", "jp": "", "desc": ""},
}

RELATION_MUSIC = {
    3001: {"en": "OST", "cn": "原声集", "jp": "", "desc": ""},
    3002: {"en": "Character Song", "cn": "角色歌", "jp": "", "desc": ""},
    3003: {"en": "Opening Song", "cn": "片头曲", "jp": "", "desc": ""},
    3004: {"en": "Ending Song", "cn": "片尾曲", "jp": "", "desc": ""},
    3005: {"en": "Insert Song", "cn": "插入歌", "jp": "", "desc": ""},
    3006: {"en": "Image Song", "cn": "印象曲", "jp": "", "desc": ""},
    3007: {"en": "Drama", "cn": "广播剧", "jp": "", "desc": ""},
    3099: {"en": "Other", "cn": "其他", "jp": "", "desc": ""},
}

RELATION_GAME = {
    1: {
        "en": "Adaptation",
        "cn": "改编",
        "jp": "",
        "desc": "同系列不同平台作品，如 CLANNAD 游戏与动画版",
    },
    4002: {"en": "Prequel", "cn": "前传", "jp": "", "desc": "发生在故事之前"},
    4003: {"en": "Sequel", "cn": "续集", "jp": "", "desc": "发生在故事之后"},
    4006: {"en": "Side Story", "cn": "资料片、外传", "jp": "", "desc": ""},
    4012: {"en": "Parent Story", "cn": "主线故事", "jp": "", "desc": ""},
    4007: {"en": "Character", "cn": "角色出演", "jp": "", "desc": "相同角色，没有关联的故事"},
    4008: {
        "en": "Same setting",
        "cn": "相同世界观",
        "jp": "",
        "desc": "发生在同一个世界观/时间线下，不同的出演角色",
    },
    4009: {
        "en": "Alternative setting",
        "cn": "不同世界观",
        "jp": "",
        "desc": "相同的出演角色，不同的世界观/时间线设定",
    },
    4010: {
        "en": "Alternative version",
        "cn": "不同演绎",
        "jp": "",
        "desc": "相同设定、角色，不同的演绎方式",
    },
    4099: {"en": "Other", "cn": "其他", "jp": "", "desc": ""},
}

RELATION_REVERSE_MAP = {
    SubjectType.book: {
        "vice_versa": True,
        "exchange_set": {
            1002: 1003,
            1003: 1002,
            1005: 1006,
            1006: 1005,
            1007: 1008,
            1008: 1007,
        },
    },
    SubjectType.anime: {
        "vice_versa": True,
        "exchange_set": {2: 3, 3: 2, 4: 5, 5: 4, 6: 12, 12: 11, 11: 12},
    },
    SubjectType.music: {"vice_versa": True},
    SubjectType.game: {
        "vice_versa": True,
        "exchange_set": {
            4002: 4003,
            4003: 4002,
            4004: 4005,
            4005: 4004,
            4006: 4012,
            4012: 4006,
        },
    },
    SubjectType.real: {
        "vice_versa": True,
        "exchange_set": {2: 3, 3: 2, 4: 5, 5: 4, 6: 12, 12: 6},
    },
}

RELATION_MAP = {
    SubjectType.anime: RELATION_SERIES,
    SubjectType.real: RELATION_SERIES,
    SubjectType.book: RELATION_BOOK,
    SubjectType.game: RELATION_GAME,
    SubjectType.music: RELATION_MUSIC,
}
