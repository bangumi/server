from typing import List, Generic, TypeVar, Optional

from pydantic import Field
from pydantic.generics import GenericModel

from pol.api.v0.models.cursorPage import CursorPage

T = TypeVar("T")
PageKeyT = TypeVar("PageKeyT")


class Paged(GenericModel, Generic[T]):

    total: int = 0
    limit: int = 0
    offset: int = 0
    data: List[T] = []


class ResponseCursorPaged(GenericModel, Generic[T, PageKeyT]):
    pagination: Optional[CursorPage[PageKeyT]] = Field(
        description="None if all replies fit in one page"
    )
    data: List[T]
