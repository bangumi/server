from sqlalchemy import (
    CHAR,
    Text,
    Column,
    String,
    DateTime,
    or_,
    and_,
    func,
    join,
    text,
    delete,
    select,
    update,
)
from sqlalchemy.dialects.mysql import insert

count = func.count

__all__ = [
    "CHAR",
    "Text",
    "Column",
    "String",
    "DateTime",
    "func",
    "join",
    "text",
    "select",
    "update",
    "insert",
    "and_",
    "func",
    "count",
    "or_",
    "get",
    "delete",
]


def get(T, *where):
    return select(T).where(*where).limit(1)
