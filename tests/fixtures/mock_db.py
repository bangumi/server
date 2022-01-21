"""
mock `sqlalchemy.ext.asyncio.AsyncSession` returned by `pol.depends.get_db`
"""
from typing import Protocol
from unittest import mock
from contextlib import asynccontextmanager

import pytest
from sqlalchemy.orm import sessionmaker
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine

from pol import config
from tests.base import async_lambda
from pol.depends import get_db


class MockAsyncSession(Protocol):
    get: mock.AsyncMock
    scalar: mock.AsyncMock
    scalars: mock.AsyncMock


@pytest.fixture()
def mock_db(app) -> MockAsyncSession:
    """mock mock AsyncSession, also override dependency `get_db` for all router"""
    db = mock.Mock()
    db.get = mock.AsyncMock(return_value=None)
    db.scalar = mock.AsyncMock(return_value=None)
    db.scalars = mock.AsyncMock(return_value=None)
    app.dependency_overrides[get_db] = async_lambda(db)
    yield db
    app.dependency_overrides.pop(get_db, None)


@pytest.fixture()
def AsyncSessionMaker():
    @asynccontextmanager
    async def get():
        engine = create_async_engine(
            "mysql+aiomysql://{}:{}@{}:{}/{}".format(
                config.MYSQL_USER,
                config.MYSQL_PASS,
                config.MYSQL_HOST,
                config.MYSQL_PORT,
                config.MYSQL_DB,
            )
        )

        SS = sessionmaker(engine, expire_on_commit=False, class_=AsyncSession)

        async with SS() as s:
            yield s

        # ensure to dispose the engine after usage.
        # otherwise, asyncio will raise a RuntimeError
        await engine.dispose()

    return get
