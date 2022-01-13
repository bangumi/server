import datetime
from typing import Any, Optional

from pydantic import Field, BaseModel

from .creator import Creator


class Revision(BaseModel):
    id: int
    type: int
    creator: Optional[Creator]
    summary: str
    created_at: datetime.datetime


class DetailedRevision(Revision):
    data: Optional[Any] = Field(None, description="编辑修改内容")
