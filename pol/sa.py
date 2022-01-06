from typing import Callable

import sqlalchemy.orm
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
    create_engine,
)
from sqlalchemy.orm import (
    eagerload,
    joinedload,
    selectinload,
    sessionmaker,
    subqueryload,
)
from sqlalchemy.dialects.mysql import insert

from pol import config

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
    "sync_session_maker",
]


def get(T, *where):
    return select(T).where(*where).limit(1)


def sync_session_maker() -> Callable[[], sqlalchemy.orm.Session]:
    engine = create_engine(
        "mysql+pymysql://{}:{}@{}:{}/{}".format(
            config.MYSQL_USER,
            config.MYSQL_PASS,
            config.MYSQL_HOST,
            config.MYSQL_PORT,
            config.MYSQL_DB,
        ),
        pool_recycle=14400,
        pool_size=10,
        max_overflow=20,
    )
    return sessionmaker(
        engine,
        expire_on_commit=False,
    )
