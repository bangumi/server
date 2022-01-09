from fastapi import Depends
from starlette.requests import Request

from pol import res
from pol.models import PublicUser
from pol.services.user_service import UserService


async def get_public_user(
    request: Request,
    username: str,
    user_service: UserService = Depends(UserService.new),
) -> PublicUser:
    """get the user for `username` like `/user/{username}/collections`"""
    try:
        return await user_service.get_by_name(username)
    except user_service.not_found:
        raise res.not_found(request)
