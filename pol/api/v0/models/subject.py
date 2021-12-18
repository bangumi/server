import datetime
from typing import Dict, Optional

from pydantic import Field, BaseModel

from pol.api.v0.models.wiki import Wiki


class SubjectEp(BaseModel):
    id: int
    type: int = Field(description="`0` 本篇，`1` SP，`2` OP，`3` ED")
    sort: float
    name: str
    name_cn: str
    duration: str
    airdate: str
    comment: int
    desc: str
    disc: int = Field(description="用于音乐条目")


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
    locked: bool
    date: Optional[datetime.date]
    platform: str = Field(description="TV, Web, 欧美剧, PS4...")
    images: Optional[Images]
    infobox: Optional[Wiki]

    volumes: int = Field(description="书籍条目的册数，由旧服务端从wiki中解析")
    eps: int = Field(description="由旧服务端从wiki中解析，对于书籍条目为`话数`")

    total_episodes: int = Field(description="数据库中的章节数量")

    rating: Rating

    collection: Collection


class RelSubject(BaseModel):
    id: int
    type: int
    name: str
    name_cn: str
    images: Optional[Images]
    relation: str


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
