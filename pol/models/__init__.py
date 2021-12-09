from typing import Any

from pydantic import Field, BaseModel


class ErrorDetail(BaseModel):
    title: str
    description: str
    detail: Any = Field(..., description="can be anything")
