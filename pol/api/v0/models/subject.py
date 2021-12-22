import datetime
from typing import Dict, Optional

from pydantic import Field, BaseModel

from pol.api.v0.models.wiki import Wiki


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


class Episode(BaseModel):
    id: int
    type: int = Field(description="`0` 本篇，`1` SP，`2` OP，`3` ED")
    name: str
    name_cn: str
    sort: float = Field(description="同类条目的排序和集数")
    ep: float = Field(None, description="条目内的集数, 从`1`开始。非本篇剧集的此字段无意义")
    airdate: str
    comment: int
    duration: str
    desc: str = Field(description="简介")
    disc: int = Field(description="音乐曲目的碟片数")


class EpisodeDetail(Episode):
    subject_id: int
