from typing import Generic, TypeVar, Optional

from pydantic import Field
from pydantic.generics import GenericModel

from pol.api.v0.models.misc import Order

T = TypeVar("T")


class CursorPage(GenericModel, Generic[T]):
    """pagination with cursor"""

    page: int = Field(le=10000)
    size: int = Field(le=100)  # number of items in this page
    total: int  # total number of items
    order: Order
    key: str = Field(description="key to sort by, e.g. id, createdAt")
    cursor: Optional[T] = Field(
        description="offset value, e.g. 1, 2020-01-01T00:00:00Z"
    )
