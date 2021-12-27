import aioredis
from databases import Database
from starlette.requests import Request
from sqlalchemy.ext.asyncio import AsyncSession


async def get_db(request: Request) -> Database:
    """defined at app.startup"""
    return request.app.state.db  # type: ignore


async def get_redis(request: Request) -> aioredis.Redis:
    """defined at app.startup"""
    return request.app.state.redis  # type: ignore


async def get_session(request: Request) -> AsyncSession:
    async with request.app.state.Session() as session:
        yield session
