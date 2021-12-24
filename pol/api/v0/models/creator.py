from typing import Optional

from pydantic import BaseModel


class Creator(BaseModel):
    id: int
    username: str
    nickname: str
    avatar: Optional[str]
