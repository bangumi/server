from fastapi import Path, Depends

from pol import res
from pol.models import PublicUser
from pol.services.user_service import UserService


async def get_public_user(
    username: str = Path(..., description="设置了 username 后无法使用UID"),
    user_service: UserService = Depends(UserService.new),
    not_found: res.HTTPException = Depends(res.not_found_exception),
) -> PublicUser:
    """
    get the user for `username` like `/user/{username}/collections`.
    UID is not working.
    """
    try:
        return await user_service.get_by_name(username)
    except UserService.NotFoundError:
        raise not_found
