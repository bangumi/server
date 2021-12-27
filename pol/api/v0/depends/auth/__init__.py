from fastapi import Depends
from pydantic import ValidationError
from starlette.status import HTTP_403_FORBIDDEN
from sqlalchemy.ext.asyncio import AsyncSession

from pol import res, curd, config
from pol.curd import NotFoundError
from pol.depends import get_redis, get_session
from pol.curd.user import User
from pol.permission import Role
from pol.redis.json_cache import JSONRedis
from pol.api.v0.depends.auth.schema import HTTPBearer, OptionalHTTPBearer

OPTIONAL_API_KEY_HEADER = OptionalHTTPBearer()

API_KEY_HEADER = HTTPBearer()


class Guest(Role):
    """this is a guest with only basic permission"""

    def allow_nsfw(self) -> bool:
        return False


guest = Guest()


async def optional_user(
    token: str = Depends(OPTIONAL_API_KEY_HEADER),
    db_session: AsyncSession = Depends(get_session),
    redis: JSONRedis = Depends(get_redis),
) -> Role:
    """
    if no auth header in request, return a guest object with only basic permission,
    otherwise, return a authorized user.
    """
    if not token:
        return guest

    return await get_current_user(token, db_session, redis)


async def get_current_user(
    token: str = Depends(API_KEY_HEADER),
    db_session: AsyncSession = Depends(get_session),
    redis: JSONRedis = Depends(get_redis),
) -> User:
    cache_key = config.CACHE_KEY_PREFIX + f"access:{token}"
    if value := await redis.get(cache_key):
        try:
            return User.parse_obj(value)
        except ValidationError:
            await redis.delete(cache_key)

    try:
        user = await curd.user.get_by_valid_token(db_session, token)
    except NotFoundError:
        raise res.HTTPException(
            status_code=HTTP_403_FORBIDDEN,
            title="unauthorized",
            description="Invalid authentication credentials",
        )

    await redis.set_json(cache_key, user.dict(by_alias=True))

    return user
