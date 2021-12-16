from typing import List, Union, Optional

from pydantic import Field, BaseModel


class Ep(BaseModel):
    id: int
    url: str
    type: int
    sort: int
    name: str
    name_cn: str
    duration: str
    airdate: str
    comment: int
    desc: str
    status: str


class Count(BaseModel):
    field_1: int = Field(..., alias="1")
    field_2: int = Field(..., alias="2")
    field_3: int = Field(..., alias="3")
    field_4: int = Field(..., alias="4")
    field_5: int = Field(..., alias="5")
    field_6: int = Field(..., alias="6")
    field_7: int = Field(..., alias="7")
    field_8: int = Field(..., alias="8")
    field_9: int = Field(..., alias="9")
    field_10: int = Field(..., alias="10")


class Rating(BaseModel):
    total: int
    count: Count
    score: float


class Images(BaseModel):
    large: str
    common: str
    medium: str
    small: str
    grid: str


class Collection(BaseModel):
    wish: int
    collect: int
    doing: int
    on_hold: int
    dropped: int


class Images1(BaseModel):
    large: str
    medium: str
    small: str
    grid: str


class Alias(BaseModel):
    field_0: Optional[str] = Field(None, alias="0")
    field_1: Optional[str] = Field(None, alias="1")
    field_2: Optional[str] = Field(None, alias="2")
    field_3: Optional[str] = Field(None, alias="3")
    field_4: Optional[str] = Field(None, alias="4")
    en: Optional[str] = None
    zh: Optional[str] = None
    en二: Optional[str] = None
    jp: Optional[str] = None
    kana: Optional[str] = None
    romaji: Optional[str] = None
    nick: Optional[str] = None


class Info(BaseModel):
    name_cn: str
    alias: Alias
    gender: str
    birth: Optional[str] = None
    bloodtype: Optional[str] = None
    height: Optional[str] = None
    weight: Optional[str] = None
    source: Optional[Union[str, List[str]]] = None


class Images2(BaseModel):
    large: str
    medium: str
    small: str
    grid: str


class Actor(BaseModel):
    id: int
    url: str
    name: str
    images: Images2


class CrtItem(BaseModel):
    id: int
    url: str
    name: str
    name_cn: str
    role_name: str
    images: Images1
    comment: int
    collects: int
    info: Info
    actors: List[Actor]


class Image(BaseModel):
    large: str
    medium: str
    small: str
    grid: str


class Alias1(BaseModel):
    kana: str
    romaji: str
    jp: Optional[str] = None
    field_0: Optional[str] = Field(None, alias="0")
    field_1: Optional[str] = Field(None, alias="1")


class Info1(BaseModel):
    name_cn: str
    alias: Optional[Alias1] = None
    gender: Optional[str] = None
    birth: Optional[str] = None
    source: Optional[List[str]] = None
    出生地: Optional[str] = None
    毕业院校: Optional[str] = None
    Twitter: Optional[str] = None


class StaffItem(BaseModel):
    id: int
    url: str
    name: str
    name_cn: str
    role_name: str
    images: Optional[Image]
    comment: int
    collects: int
    info: Info1
    jobs: List[str]


class Avatar(BaseModel):
    large: str
    medium: str
    small: str


class User(BaseModel):
    id: int
    url: str
    username: str
    nickname: str
    avatar: Avatar
    sign: Optional[str]


class TopicItem(BaseModel):
    id: int
    url: str
    title: str
    main_id: int
    timestamp: int
    lastpost: int
    replies: int
    user: User


class Avatar1(BaseModel):
    large: str
    medium: str
    small: str


class BlogItem(BaseModel):
    id: int
    url: str
    title: str
    summary: str
    image: str
    replies: int
    timestamp: int
    dateline: str
    user: User


class Model(BaseModel):
    id: int
    url: str
    type: int
    name: str
    name_cn: str
    summary: str
    eps: List[Ep]
    eps_count: int
    air_date: str
    air_weekday: int
    rating: Rating
    rank: int
    images: Images
    collection: Collection
    crt: List[CrtItem]
    staff: List[StaffItem]
    topic: List[TopicItem]
    blog: List[BlogItem]
