from fastapi import Depends, APIRouter
from pydantic import BaseModel

from pol import res
from pol.models import ErrorDetail
from pol.permission import UserGroup
from .depends.auth import User, get_current_user

router = APIRouter(tags=["用户"])


class Me(BaseModel):
    id: int
    username: str
    nickname: str
    group_id: UserGroup


@router.get(
    "/me",
    response_model=Me,
    responses={
        403: res.response(model=ErrorDetail, description="unauthorized"),
    },
)
async def get_subject(
    user: User = Depends(get_current_user),
):
    return user.dict(by_alias=False)
