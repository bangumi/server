import aioredis
from starlette.requests import Request


async def get_redis(request: Request) -> aioredis.Redis:
    """defined at app.startup"""
    return request.app.state.redis  # type: ignore


async def get_db(request: Request):
    async with request.app.state.Session() as session:
        yield session
