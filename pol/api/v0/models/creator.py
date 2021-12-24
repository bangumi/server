from typing import Union

from pydantic import BaseModel


class Creator(BaseModel):
    id: int
    username: str
    nickname: str
    avatar: Union[str, None]
