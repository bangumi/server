import datetime
from typing import Any, Optional

from pydantic import Field, BaseModel

from .creator import Creator


class Revision(BaseModel):
    id: int
    type: int
    creator: Creator
    summary: str
    timestamp: datetime.datetime


class DetailedRevision(Revision):
    data: Optional[Any] = Field(None, description="编辑修改内容")
