from fastapi import Depends
from pydantic import ValidationError
from starlette.status import HTTP_403_FORBIDDEN

from pol import res, config
from pol.models import User
from pol.depends import get_redis
from pol.permission import Role
from pol.redis.json_cache import JSONRedis
from pol.services.user_service import UserService
from pol.api.v0.depends.auth.schema import HTTPBearer, OptionalHTTPBearer

OPTIONAL_API_KEY_HEADER = OptionalHTTPBearer()

API_KEY_HEADER = HTTPBearer()


class Guest(Role):
    """this is a guest with only basic permission"""

    def allow_nsfw(self) -> bool:
        return False

    def get_user_id(self) -> int:
        return 0


guest = Guest()


async def optional_user(
    token: str = Depends(OPTIONAL_API_KEY_HEADER),
    service: UserService = Depends(UserService.new),
    redis: JSONRedis = Depends(get_redis),
) -> Role:
    """
    if no auth header in request, return a guest object with only basic permission,
    otherwise, return a authorized user.
    """
    if not token:
        return guest

    return await get_current_user(token=token, redis=redis, service=service)


async def get_current_user(
    token: str = Depends(API_KEY_HEADER),
    redis: JSONRedis = Depends(get_redis),
    service: UserService = Depends(UserService.new),
) -> User:
    cache_key = config.CACHE_KEY_PREFIX + f"access:{token}"
    if value := await redis.get(cache_key):
        try:
            return User.parse_obj(value)
        except ValidationError:
            await redis.delete(cache_key)

    try:
        user = await service.get_by_access_token(token)
    except UserService.not_found:
        raise res.HTTPException(
            status_code=HTTP_403_FORBIDDEN,
            title="unauthorized",
            description="Invalid authentication credentials",
        )

    await redis.set_json(cache_key, user.dict(by_alias=True), ex=60)

    return user
