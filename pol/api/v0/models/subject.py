import datetime
from typing import Dict

from pydantic import BaseModel

from pol.api.v0.models.wiki import Wiki


class SubjectEp(BaseModel):
    id: int
    # url: str
    type: int
    sort: int
    name: str
    name_cn: str
    duration: str
    airdate: str
    comment: int
    desc: str
    # status: str


class Rating(BaseModel):
    rank: int
    total: int
    count: Dict[str, int]
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


class Subject(BaseModel):
    id: int
    type: int
    name: str
    name_cn: str
    summary: str
    nsfw: bool
    date: datetime.date
    platform: int
    images: Images
    infobox: Wiki

    rating: Rating

    collection: Collection

    # air_date: str
    # air_weekday: int
    # rating: Rating
    # rank: int
    # collection: Collection
    # crt: List[CrtItem]
    # staff: List[StaffItem]
    # topic: List[TopicItem]
    # blog: List[BlogItem]


class Ep(BaseModel):
    id: int
    # url: str
    type: int
    sort: int
    name: str
    name_cn: str
    duration: str
    airdate: str
    comment: int
    desc: str
    comment: int

    # status: str
