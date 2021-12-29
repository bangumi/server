from typing import Generic, TypeVar

from pydantic.generics import GenericModel

T = TypeVar("T")


class Pagination(GenericModel, Generic[T]):
    page: int
    size: int
    isAsc: bool
    # pagination key
    key: T
