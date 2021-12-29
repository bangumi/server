from typing import List, Generic, TypeVar, Optional

from pydantic import Field
from pydantic.generics import GenericModel

from pol.api.v0.models.cursorPage import CursorPage

T = TypeVar("T")


class Paged(GenericModel, Generic[T]):
    total: int
    limit: int
    offset: int
    data: List[T]


class CursorPaged(GenericModel, Generic[T]):
    pagination: Optional[CursorPage] = Field(
        description="None if all replies fit in one page")
    data: List[T]


class ListResponse(Generic[T]):
    data: T
