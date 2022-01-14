from starlette.requests import Request

from pol.redis.json_cache import JSONRedis


async def get_redis(request: Request) -> JSONRedis:
    """defined at app.startup"""
    return request.app.state.redis  # type: ignore


async def get_db(request: Request):
    async with request.app.state.Session() as session:
        yield session
