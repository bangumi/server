from fastapi import Depends
from pydantic import ValidationError
from starlette.status import HTTP_403_FORBIDDEN
from sqlalchemy.ext.asyncio import AsyncSession

import pol.curd.user
import pol.permission.roles
from pol import res, config
from pol.depends import get_db, get_redis
from pol.curd.user import User
from pol.curd.exceptions import NotFoundError
from pol.redis.json_cache import JSONRedis
from pol.api.v0.depends.auth.schema import HTTPBearer, OptionalHTTPBearer

OPTIONAL_API_KEY_HEADER = OptionalHTTPBearer()

API_KEY_HEADER = HTTPBearer()


async def optional_user(
    token: str = Depends(OPTIONAL_API_KEY_HEADER),
    db: AsyncSession = Depends(get_db),
    redis: JSONRedis = Depends(get_redis),
) -> User:
    """
    if no auth header in request, return a guest object with only basic permission,
    otherwise, return an authorized user.
    """
    if not token:
        return User.default_user()

    return await get_current_user(token, db, redis)


async def get_current_user(
    token: str = Depends(API_KEY_HEADER),
    db: AsyncSession = Depends(get_db),
    redis: JSONRedis = Depends(get_redis),
) -> User:
    cache_key = config.CACHE_KEY_PREFIX + f"access:{token}"
    if value := await redis.get(cache_key):
        try:
            return User.parse_obj(value)
        except ValidationError:
            await redis.delete(cache_key)

    try:
        user = await pol.curd.user.get_by_valid_token(db, token)
    except NotFoundError:
        raise res.HTTPException(
            status_code=HTTP_403_FORBIDDEN,
            title="unauthorized",
            description="Invalid authentication credentials",
        )

    await redis.set_json(cache_key, user.dict(by_alias=True), ex=60)

    return user
