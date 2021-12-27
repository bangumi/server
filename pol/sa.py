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
from sqlalchemy.orm import eagerload, joinedload, selectinload, subqueryload
from sqlalchemy.dialects.mysql import insert

count = func.count

__all__ = [
    "CHAR",
    "selectinload",
    "joinedload",
    "Text",
    "Column",
    "subqueryload",
    "String",
    "DateTime",
    "func",
    "join",
    "eagerload",
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
