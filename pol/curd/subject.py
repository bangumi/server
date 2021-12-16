import datetime

from pydantic import Field, BaseModel
from databases import Database

from pol.db.tables import ChiiSubject, ChiiSubjectField
from pol.db_models import sa
from pol.curd.exceptions import NotFoundError


class Subject(BaseModel):
    id: int = Field(alias="subject_id")

    name: str = Field(alias="subject_name")
    name_cm: str = Field(alias="subject_name_cn")

    type_id: int = Field(alias="subject_type_id")
    ban: bool = Field(alias="subject_ban")
    redirect: int = Field(alias="field_redirect")
    date: datetime.date = Field(alias="field_date")
    dateline: int = Field(alias="subject_dateline")
    nsfw: bool = Field(alias="subject_nsfw")
    platform: int = Field(alias="subject_platform")
    image: str = Field(alias="subject_image")
    summary: str = Field(alias="field_summary")
    infobox: str = Field(alias="field_infobox")

    # series_entry: int
    # volumes: int

    dropped: int = Field(alias="subject_dropped")
    on_hold: int = Field(alias="subject_on_hold")
    doing: int = Field(alias="subject_doing")
    collect: int = Field(alias="subject_collect")
    wish: int = Field(alias="subject_wish")


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
