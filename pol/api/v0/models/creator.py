from typing import Optional

from pydantic import BaseModel


class Creator(BaseModel):
    id: int
    nickname: Optional[str]
    avatar: Optional[str]
