from typing import List, Generic, TypeVar

from pydantic.main import BaseModel
from pydantic.generics import GenericModel

T = TypeVar("T")
PageKeyT = TypeVar("PageKeyT")


class Paged(GenericModel, Generic[T]):

    total: int = 0
    limit: int = 0
    offset: int = 0
    data: List[T] = []


class OffsetPage(BaseModel):
    total: int = 0
    page: int = 0
    size: int = 0


class OffsetPagedResponse(GenericModel, Generic[T]):
    page: OffsetPage
    data: List[T]
