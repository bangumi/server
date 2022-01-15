import asyncio
import inspect
from functools import wraps


def async_test(f):
    if not inspect.iscoroutinefunction(f):  # pragma: no cover
        raise ValueError("not async test function")

    @wraps(f)
    def test_a(*args, **kwargs):
        return asyncio.get_event_loop().run_until_complete(f(*args, **kwargs))

    return test_a


def async_lambda(value):
    async def f():
        return value

    return f
