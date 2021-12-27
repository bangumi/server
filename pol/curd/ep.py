import datetime

from pydantic import Field, BaseModel


class Ep(BaseModel):
    id: int = Field(alias="ep_id")
    subject_id: int = Field(alias="ep_subject_id")
    sort: float = Field(alias="ep_sort")
    type: int = Field(alias="ep_type")
    disc: int = Field(0, alias="ep_disc")
    name: str = Field(alias="ep_name")
    name_cn: str = Field(alias="ep_name_cn")
    duration: str = Field(alias="ep_duration")
    airdate: str = Field(alias="ep_airdate")
    online: str = Field(alias="ep_online")
    comment: int = Field(alias="ep_comment")
    # resources: int = Field(alias="ep_resources")
    desc: str = Field(alias="ep_desc")
    dateline: datetime.datetime = Field(alias="ep_dateline")
    lastpost: datetime.datetime = Field(alias="ep_lastpost")
    lock: bool = Field(alias="ep_lock")
    ban: bool = Field(alias="ep_ban")

    class Config:
        orm_mode = True
