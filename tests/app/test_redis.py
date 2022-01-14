import pytest
from redis import Redis

from pol import config
from tests.base import async_test
from pol.redis.json_cache import JSONRedis


@pytest.mark.env("redis")
@async_test
async def test_redis_util(redis_client: Redis):
    key = "test-key"
    redis_client.set(key, b"bb")
    async with JSONRedis.from_url(config.REDIS_URI) as redis:
        v = await redis.get("test-key")
        assert v is None
