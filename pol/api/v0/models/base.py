from typing import List, Generic, TypeVar

from pydantic.generics import GenericModel

__all__ = ["Paged"]

T = TypeVar("T")


class Paged(GenericModel, Generic[T]):
    total: int = 0
    limit: int = 0
    offset: int = 0
    data: List[T] = []
