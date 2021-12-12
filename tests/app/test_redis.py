import asyncio
import inspect
from functools import wraps

from redis import Redis

from pol import config
from pol.redis.json_cache import JSONRedis


def async_test(f):
    if not inspect.iscoroutinefunction(f):
        raise ValueError("not async test function")

    @wraps(f)
    def test_a(*args, **kwargs):
        return asyncio.get_event_loop().run_until_complete(f(*args, **kwargs))

    return test_a


@async_test
async def test_redis_util(redis_client: Redis):
    key = "test-key"
    redis_client.set(key, b"bb")
    async with JSONRedis.from_url(config.REDIS_URI) as redis:
        v = await redis.get("test-key")
        assert v is None
