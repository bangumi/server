from typing import List, Generic, TypeVar

from pydantic.generics import GenericModel

S = TypeVar("S")
T = TypeVar("T")


class Paged(GenericModel, Generic[T]):
    total: int = 0
    limit: int = 0
    offset: int = 0
    data: List[T] = []


class PageInfo(GenericModel, Generic[T]):
    next: T


class Pagination(GenericModel, Generic[S, T]):
    pagination: PageInfo[S]
    data: List[T] = []
