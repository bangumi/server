from fastapi import Path, Depends

from pol import res
from pol.models import PublicUser
from pol.permission import Role
from pol.models.subject import Subject
from pol.http_cache.depends import CacheControl
from pol.api.v0.depends.auth import optional_user
from pol.services.user_service import UserService
from pol.services.index_service import IndexService
from pol.services.subject_service import SubjectService


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


async def get_subject(
    subject_id: int = Path(..., gt=0),
    subject_service: SubjectService = Depends(SubjectService.new),
    user: Role = Depends(optional_user),
    not_found: res.HTTPException = Depends(res.not_found_exception),
) -> Subject:
    """
    make sure current subject is visible for current user
    also omit merged subject
    """
    try:
        return await subject_service.get_by_id(
            subject_id,
            user.allow_nsfw(),
            include_redirect=False,
            include_banned=False,
        )
    except SubjectService.NotFoundError:
        raise not_found


async def get_index(
    index_id: int = Path(0, gt=0),
    index_service: IndexService = Depends(IndexService.new),
    user: Role = Depends(optional_user),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(res.not_found_exception),
):
    """
    make sure current subject is visible for current user
    also omit merged subject
    """
    if not user.allow_nsfw() and await index_service.get_index_nsfw_by_id(index_id):
        raise not_found
    cache_control(300)
    try:
        return await index_service.get_index_by_id(index_id)
    except IndexService.NotFoundError:
        raise not_found
