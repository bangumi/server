import datetime

from pydantic import BaseModel

from pol.api.v0.models.creator import Creator


class Revision(BaseModel):
    id: int
    creator: Creator
    timestamp: datetime.datetime
    summary: str
