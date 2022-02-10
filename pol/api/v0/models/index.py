import datetime
from typing import Any, Optional

from pydantic import Field, BaseModel

from pol import wiki
from pol.api.v0.utils import subject_images
from . import Stat
from .wiki import Wiki
from .creator import Creator
from .subject import Images


class Index(BaseModel):
    id: int
    title: str
    desc: str
    total: int = Field(0, description="收录条目总数")
    stat: Stat = Field(..., description="目录评论及收藏数")
    created_at: datetime.datetime
    creator: Creator
    ban: bool


class IndexSubject(BaseModel):
    __doc__ = '同名字段意义同<a href="#model-Subject">Subject</a>'

    id: int
    type: int
    name: str
    images: Optional[Images]
    infobox: Optional[Wiki]
    date: Optional[str]
    comment: str
    added_at: datetime.datetime

    def __init__(
        self,
        infobox: Any = None,
        images: Any = None,
        **data: Any,
    ) -> None:
        if isinstance(infobox, str):
            try:
                infobox = wiki.parse(infobox).info
            except wiki.WikiSyntaxError:
                infobox = None
        if isinstance(images, str):
            images = subject_images(images)

        super().__init__(**data, infobox=infobox, images=images)
