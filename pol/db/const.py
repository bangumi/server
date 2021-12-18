import enum
from typing import TYPE_CHECKING, Dict, Optional, NamedTuple

from pol.db._const import (
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


class Relation(NamedTuple):
    cn: str
    en: str
    jp: str = ""
    desc: str = ""

    def get(self) -> str:
        return self.cn or self.en or self.jp


RELATION_SERIES: Dict[int, Relation] = {
    1: Relation(en="Adaptation", cn="改编", desc="同系列不同平台作品，如柯南漫画与动画版"),
    2: Relation(en="Prequel", cn="前传", desc="发生在故事之前"),
    3: Relation(en="Sequel", cn="续集", desc="发生在故事之后"),
    4: Relation(en="Summary", cn="总集篇", desc="对故事的概括版本"),
    5: Relation(en="Full Story", cn="全集", desc="相对于剧场版/总集篇的完整故事"),
    6: Relation(en="Side Story", cn="番外篇"),
    7: Relation(en="Character", cn="角色出演", desc="相同角色，没有关联的故事"),
    8: Relation(en="Same setting", cn="相同世界观", desc="发生在同一个世界观/时间线下，不同的出演角色"),
    9: Relation(en="Alternative setting", cn="不同世界观", desc="相同的出演角色，不同的世界观/时间线设定"),
    10: Relation(
        en="Alternative version", cn="不同演绎", desc="相同设定、角色，不同的演绎方式（如EVA原作与新剧场版)"
    ),
    11: Relation(en="Spin-off", cn="衍生", desc="如柯南与魔术快斗"),
    12: Relation(en="Parent Story", cn="主线故事"),
    99: Relation(en="Other", cn="其他"),
}

RELATION_BOOK: Dict[int, Relation] = {
    1: Relation(en="Adaptation", cn="改编", desc="同系列不同平台作品，如柯南漫画与动画版"),
    1003: Relation(en="Offprint", cn="单行本"),
    1002: Relation(en="Series", cn="系列"),
    1004: Relation(en="Album", cn="画集"),
    1005: Relation(en="Prequel", cn="前传", desc="发生在故事之前"),
    1006: Relation(en="Sequel", cn="续集", desc="发生在故事之后"),
    1007: Relation(en="Side Story", cn="番外篇"),
    1008: Relation(en="Parent Story", cn="主线故事"),
    1010: Relation(en="Alternative version", cn="不同版本"),
    1011: Relation(en="Character", cn="角色出演", desc="相同角色，没有关联的故事"),
    1012: Relation(en="Same setting", cn="相同世界观", desc="发生在同一个世界观/时间线下，不同的出演角色"),
    1013: Relation(en="Alternative setting", cn="不同世界观", desc="相同的出演角色，不同的世界观/时间线设定"),
    1099: Relation(en="Other", cn="其他"),
}

RELATION_MUSIC: Dict[int, Relation] = {
    3001: Relation(en="OST", cn="原声集"),
    3002: Relation(en="Character Song", cn="角色歌"),
    3003: Relation(en="Opening Song", cn="片头曲"),
    3004: Relation(en="Ending Song", cn="片尾曲"),
    3005: Relation(en="Insert Song", cn="插入歌"),
    3006: Relation(en="Image Song", cn="印象曲"),
    3007: Relation(en="Drama", cn="广播剧"),
    3099: Relation(en="Other", cn="其他"),
}

RELATION_GAME: Dict[int, Relation] = {
    1: Relation(en="Adaptation", cn="改编", desc="同系列不同平台作品，如 CLANNAD 游戏与动画版"),
    4002: Relation(en="Prequel", cn="前传", desc="发生在故事之前"),
    4003: Relation(en="Sequel", cn="续集", desc="发生在故事之后"),
    4006: Relation(en="Side Story", cn="资料片、外传"),
    4012: Relation(en="Parent Story", cn="主线故事"),
    4007: Relation(en="Character", cn="角色出演", desc="相同角色，没有关联的故事"),
    4008: Relation(en="Same setting", cn="相同世界观", desc="发生在同一个世界观/时间线下，不同的出演角色"),
    4009: Relation(en="Alternative setting", cn="不同世界观", desc="相同的出演角色，不同的世界观/时间线设定"),
    4010: Relation(en="Alternative version", cn="不同演绎", desc="相同设定、角色，不同的演绎方式"),
    4099: Relation(en="Other", cn="其他"),
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

RELATION_MAP: Dict[SubjectType, Dict[int, Relation]] = {
    SubjectType.book: RELATION_BOOK,
    SubjectType.anime: RELATION_SERIES,
    SubjectType.music: RELATION_MUSIC,
    SubjectType.game: RELATION_GAME,
    SubjectType.real: RELATION_SERIES,
}

PLATFORM_MAP = {
    SubjectType.book: {
        0: {
            "id": 0,
            "type": "other",
            "type_cn": "其他",
            "alias": "misc",
            "wiki_tpl": "Book",
        },
        1001: {
            "id": 1001,
            "type": "Comic",
            "type_cn": "漫画",
            "alias": "comic",
            "wiki_tpl": "Manga",
            "enable_header": True,
        },
        1002: {
            "id": 1002,
            "type": "Novel",
            "type_cn": "小说",
            "alias": "novel",
            "wiki_tpl": "Novel",
            "enable_header": True,
        },
        1003: {
            "id": 1003,
            "type": "Illustration",
            "type_cn": "画集",
            "alias": "illustration",
            "wiki_tpl": "Book",
            "enable_header": True,
        },
    },
    SubjectType.anime: {
        0: {
            "id": 0,
            "type": "other",
            "type_cn": "其他",
            "alias": "misc",
            "wiki_tpl": "Anime",
        },
        1: {
            "id": 1,
            "type": "TV",
            "type_cn": "TV",
            "alias": "tv",
            "enable_header": True,
            "wiki_tpl": "TVAnime",
        },
        2: {
            "id": 2,
            "type": "OVA",
            "type_cn": "OVA",
            "alias": "ova",
            "enable_header": True,
            "wiki_tpl": "OVA",
        },
        3: {
            "id": 3,
            "type": "movie",
            "type_cn": "剧场版",
            "alias": "movie",
            "enable_header": True,
            "wiki_tpl": "Movie",
        },
        5: {
            "id": 5,
            "type": "web",
            "type_cn": "WEB",
            "alias": "web",
            "enable_header": True,
            "wiki_tpl": "TVAnime",
        },
    },
    SubjectType.real: {
        0: {
            "id": 0,
            "type": "other",
            "type_cn": "其他",
            "alias": "misc",
            "wiki_tpl": "TV",
        },
        1: {
            "id": 1,
            "type": "jp",
            "type_cn": "日剧",
            "alias": "jp",
            "enable_header": True,
            "wiki_tpl": "TV",
        },
        2: {
            "id": 2,
            "type": "en",
            "type_cn": "欧美剧",
            "alias": "en",
            "enable_header": True,
            "wiki_tpl": "TV",
        },
        3: {
            "id": 3,
            "type": "cn",
            "type_cn": "华语剧",
            "alias": "cn",
            "enable_header": True,
            "wiki_tpl": "TV",
        },
    },
    SubjectType.game: {
        4: {
            "id": 4,
            "type": "PC",
            "alias": "pc",
            "search_string": "pc|windows",
            "type_cn": "PC",
        },
        5: {
            "id": 5,
            "type": "NDS",
            "alias": "nds",
            "search_string": "nds",
            "type_cn": "NDS",
        },
        6: {
            "id": 6,
            "type": "PSP",
            "alias": "psp",
            "search_string": "psp",
            "type_cn": "PSP",
        },
        7: {
            "id": 7,
            "type": "PS2",
            "alias": "ps2",
            "search_string": "PS2",
            "type_cn": "PS2",
        },
        8: {
            "id": 8,
            "type": "PS3",
            "alias": "ps3",
            "search_string": "PS3|(PlayStation 3)",
            "type_cn": "PS3",
        },
        9: {
            "id": 9,
            "type": "Xbox360",
            "alias": "xbox360",
            "search_string": "xbox360",
            "type_cn": "Xbox360",
        },
        33: {
            "id": 33,
            "type": "Mac OS",
            "alias": "mac",
            "search_string": "mac",
            "type_cn": "Mac OS",
        },
        38: {
            "id": 38,
            "type": "PS5",
            "alias": "ps5",
            "search_string": "PS5",
            "type_cn": "PS5",
        },
        39: {
            "id": 39,
            "type": "Xbox Series X/S",
            "alias": "xbox_series_xs",
            "search_string": "XSX|XSS|(Xbox Series X)|(Xbox Series S)",
            "type_cn": "Xbox Series X/S",
        },
        34: {
            "id": 34,
            "type": "PS4",
            "alias": "ps4",
            "search_string": "PS4",
            "type_cn": "PS4",
        },
        35: {
            "id": 35,
            "type": "Xbox One",
            "alias": "xbox_one",
            "search_string": "(Xbox One)",
            "type_cn": "Xbox One",
        },
        37: {
            "id": 37,
            "type": "Nintendo Switch",
            "alias": "ns",
            "search_string": "(Nintendo Switch)|NS",
            "type_cn": "Nintendo Switch",
        },
        36: {
            "id": 36,
            "type": "Wii U",
            "alias": "wii_u",
            "search_string": "(Wii U)|WiiU",
            "type_cn": "Wii U",
        },
        10: {
            "id": 10,
            "type": "Wii",
            "alias": "wii",
            "search_string": "Wii",
            "type_cn": "Wii",
        },
        30: {
            "id": 30,
            "type": "PSVita",
            "alias": "psv",
            "search_string": "psv|vita",
            "type_cn": "PS Vita",
        },
        31: {
            "id": 31,
            "type": "3DS",
            "alias": "3ds",
            "search_string": "3ds",
            "type_cn": "3DS",
        },
        11: {
            "id": 11,
            "type": "iOS",
            "alias": "iphone",
            "search_string": "iphone|ipad|ios",
            "type_cn": "iOS",
        },
        32: {
            "id": 32,
            "type": "Android",
            "alias": "android",
            "search_string": "android",
            "type_cn": "Android",
        },
        12: {
            "id": 12,
            "type": "ARC",
            "alias": "arc",
            "search_string": "ARC|街机",
            "type_cn": "街机",
        },
        15: {
            "id": 15,
            "type": "XBOX",
            "alias": "xbox",
            "search_string": "XBOX",
            "type_cn": "XBOX",
        },
        17: {
            "id": 17,
            "type": "GameCube",
            "alias": "gamecube",
            "search_string": "GameCube|ngc",
            "type_cn": "GameCube",
        },
        27: {
            "id": 27,
            "type": "Dreamcast",
            "alias": "dreamcast",
            "search_string": "dc",
            "type_cn": "Dreamcast",
        },
        21: {
            "id": 21,
            "type": "Nintendo 64",
            "alias": "n64",
            "search_string": "n64",
            "type_cn": "Nintendo 64",
        },
        28: {
            "id": 28,
            "type": "PlayStation",
            "alias": "ps",
            "search_string": "ps",
            "type_cn": "PlayStation",
        },
        19: {
            "id": 19,
            "type": "SFC",
            "alias": "sfc",
            "search_string": "SFC",
            "type_cn": "SFC",
        },
        20: {
            "id": 20,
            "type": "FC",
            "alias": "fc",
            "search_string": "FC",
            "type_cn": "FC",
        },
        18: {
            "id": 18,
            "type": "NEOGEO Pocket Color",
            "alias": "ngp",
            "search_string": "ngp",
            "type_cn": "NEOGEO Pocket Color",
        },
        22: {
            "id": 22,
            "type": "GBA",
            "alias": "GBA",
            "search_string": "GBA",
            "type_cn": "GBA",
        },
        23: {
            "id": 23,
            "type": "GB",
            "alias": "GB",
            "search_string": "GB",
            "type_cn": "GB",
        },
        25: {
            "id": 25,
            "type": "Virtual Boy",
            "alias": "vb",
            "search_string": "Virtual Boy",
            "type_cn": "Virtual Boy",
        },
        26: {
            "id": 26,
            "type": "WonderSwan Color",
            "alias": "wsc",
            "search_string": "wsc",
            "type_cn": "WonderSwan Color",
        },
        29: {
            "id": 29,
            "type": "WonderSwan",
            "alias": "ws",
            "search_string": "ws",
            "type_cn": "WonderSwan",
        },
    },
}
