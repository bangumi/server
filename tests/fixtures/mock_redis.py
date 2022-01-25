"""mock redis"""
from typing import Protocol
from unittest import mock

import redis
import pytest

from pol import config
from pol.depends import get_redis

__all__ = ["MockRedis", "mock_redis", "redis_client"]


class MockRedis(Protocol):
    get: mock.AsyncMock
    set_json: mock.AsyncMock
    get_with_model: mock.AsyncMock


@pytest.fixture()
def mock_redis(app) -> MockRedis:
    """mock redis client, also override dependency `get_redis` for all router"""
    r = mock.Mock()
    r.set_json = mock.AsyncMock()
    r.get = mock.AsyncMock(return_value=None)
    r.get_with_model = mock.AsyncMock(return_value=None)

    async def mocker():
        return r

    app.dependency_overrides[get_redis] = mocker
    yield r
    app.dependency_overrides.pop(get_redis, None)


@pytest.fixture()
def redis_client():
    """fixture to access redis server"""
    with redis.Redis.from_url(config.REDIS_URI) as redis_client:
        redis_client.flushdb()
        try:
            yield redis_client
        finally:
            redis_client.flushdb()
