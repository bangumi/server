from pydantic import BaseModel


class Creator(BaseModel):
    id: int
    nickname: str
