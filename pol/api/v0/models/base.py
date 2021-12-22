from typing import List, Generic, TypeVar

from pydantic.generics import GenericModel

T = TypeVar("T")


class Paged(GenericModel, Generic[T]):
    total: int
    limit: int
    offset: int
    data: List[T]
