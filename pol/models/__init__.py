from pydantic import BaseModel


class ErrorDetail(BaseModel):
    detail: str
