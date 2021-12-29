from typing import Generic, TypeVar

from pydantic import Field
from pydantic.generics import GenericModel

from pol.api.v0.models import Order

T = TypeVar("T")


class CursorPage(GenericModel, Generic[T]):
    """pagination with cursor"""
    page: int
    size: int
    order: Order
    key: str = Field(description="key to sort by, e.g. id, createdAt")
    cursor: T = Field(description="offset value, e.g. 1, 2020-01-01T00:00:00Z")
