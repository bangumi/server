from typing import TypeVar, Optional

from pydantic.generics import GenericModel

from pol.api.v0.models.misc import Order
from pol.api.v0.models.cursorPage import CursorPage

T = TypeVar("T")


class SortParams(GenericModel, T):
    sortBy: str
    pageKey: T


def request_to_sort_params(
    q: Optional[str] = None,
    key: Optional[str] = None,
    cursor: Optional[T] = None,
    page: Optional[int] = 0,
    size: Optional[int] = 0,
    order: Optional[Order] = Order.asc,
) -> CursorPage[T]:

    return CursorPage[T](page=page, size=size, order=order, key=key, cursor=cursor)
