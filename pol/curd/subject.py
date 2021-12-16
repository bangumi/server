import datetime

from pydantic import Field, BaseModel
from databases import Database

from pol.utils import subject_images
from pol.db.tables import ChiiSubject, ChiiSubjectField
from pol.db_models import sa
from pol.curd.exceptions import NotFoundError


class Subject(BaseModel):
    id: int = Field(alias="subject_id")

    name: str = Field(alias="subject_name")
    name_cn: str = Field(alias="subject_name_cn")

    type: int = Field(alias="subject_type_id")
    ban: bool = Field(alias="subject_ban")
    redirect: int = Field(alias="field_redirect")
    date: datetime.date = Field(alias="field_date")
    dateline: int = Field(alias="subject_dateline")
    nsfw: bool = Field(alias="subject_nsfw")
    platform: int = Field(alias="subject_platform")
    image: str = Field(alias="subject_image")
    summary: str = Field(alias="field_summary")
    infobox: str = Field(alias="field_infobox")

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
        score = round(total / total_count, 1)

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

    @property
    def collection(self):
        return {
            "wish": self.wish,
            "collect": self.collect,
            "doing": self.doing,
            "on_hold": self.on_hold,
            "dropped": self.dropped,
        }


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
