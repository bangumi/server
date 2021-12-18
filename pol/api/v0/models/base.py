from typing import Any

from pydantic import BaseModel


class Paged(BaseModel):
    total: int
    limit: int
    offset: int
    data: Any
