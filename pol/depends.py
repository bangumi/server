import aioredis
from fastapi import FastAPI
from databases import Database
from starlette.requests import Request


async def fastapi_app(request: Request) -> FastAPI:
    return request.app  # type: ignore


async def get_db(request: Request) -> Database:
    """defined at app.startup"""
    return request.app.state.db  # type: ignore


async def get_redis(request: Request) -> aioredis.Redis:
    """defined at app.startup"""
    return request.app.state.redis  # type: ignore
