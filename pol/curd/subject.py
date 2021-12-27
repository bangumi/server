import datetime
from typing import Optional

from pydantic import Field, BaseModel, validator
from databases import Database

from pol import sa
from pol.utils import subject_images
from pol.db.tables import ChiiSubject, ChiiSubjectField
from pol.curd.exceptions import NotFoundError
from pol.compat.phpseralize import loads, dict_to_list


class Subject(BaseModel):
    id: int = Field(alias="subject_id")

    name: str = Field(alias="subject_name")
    name_cn: str = Field(alias="subject_name_cn")

    type: int = Field(alias="subject_type_id")
    ban: int = Field(alias="subject_ban", description="1 为重定向 2 为锁定")
    redirect: int = Field(alias="field_redirect")
    date: Optional[datetime.date] = Field(alias="field_date")
    dateline: int = Field(alias="subject_dateline")
    nsfw: bool = Field(alias="subject_nsfw")
    platform: int = Field(alias="subject_platform")
    image: str = Field(alias="subject_image")
    summary: str = Field(alias="field_summary")
    infobox: str = Field(alias="field_infobox")

    volumes: int = Field(alias="field_volumes")
    eps: int = Field(alias="field_eps")

    rank: int = Field(alias="field_rank")

    dropped: int = Field(alias="subject_dropped")
    on_hold: int = Field(alias="subject_on_hold")
    doing: int = Field(alias="subject_doing")
    collect: int = Field(alias="subject_collect")
    wish: int = Field(alias="subject_wish")

    # 评分
    rate_1: int = Field(alias="field_rate_1")
    rate_2: int = Field(alias="field_rate_2")
    rate_3: int = Field(alias="field_rate_3")
    rate_4: int = Field(alias="field_rate_4")
    rate_5: int = Field(alias="field_rate_5")
    rate_6: int = Field(alias="field_rate_6")
    rate_7: int = Field(alias="field_rate_7")
    rate_8: int = Field(alias="field_rate_8")
    rate_9: int = Field(alias="field_rate_9")
    rate_10: int = Field(alias="field_rate_10")

    # 序列化标签
    tags_serialized: str = Field(alias="field_tags")

    class Config:
        orm_mode = True

    @validator("date", pre=True)
    def handle_mysql_zero_value(cls, v):
        if v == "0000-00-00":
            return None
        return v

    @property
    def locked(self):
        return self.ban == 2

    @property
    def images(self):
        return subject_images(self.image)

    def rating(self):
        scores = self.scores
        total = 0
        total_count = 0
        for key, value in scores.items():
            total += int(key) * value
            total_count += value
        if total_count != 0:
            score = round(total / total_count, 1)
        else:
            score = 0

        return {
            "rank": self.rank,
            "score": score,
            "count": scores,
            "total": total,
        }

    @property
    def scores(self):
        return {
            "1": self.rate_1,
            "2": self.rate_2,
            "3": self.rate_3,
            "4": self.rate_4,
            "5": self.rate_5,
            "6": self.rate_6,
            "7": self.rate_7,
            "8": self.rate_8,
            "9": self.rate_9,
            "10": self.rate_10,
        }

    def tags(self):
        if not self.tags_serialized:
            return []

        # defaults to utf-8
        tags_deserialized = dict_to_list(loads(self.tags_serialized.encode()))

        return [
            {"name": tag["tag_name"], "count": tag["result"]}
            for tag in tags_deserialized
            if tag["tag_name"] is not None  # remove tags like { "tag_name": None }
        ]


async def get_one(db: Database, *where) -> Subject:
    query = (
        sa.select(ChiiSubject, ChiiSubjectField)
        .join(ChiiSubjectField, ChiiSubjectField.field_sid == ChiiSubject.subject_id)
        .where(*where)
        .limit(1)
    )
    r = await db.fetch_one(query)

    if r:
        return Subject.parse_obj(r)

    raise NotFoundError()
