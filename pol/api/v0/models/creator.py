from pydantic import BaseModel

__all__ = ["Creator"]


class Creator(BaseModel):
    __doc__ = '意义同<a href="#model-Me">Me</a>'

    username: str
    nickname: str
