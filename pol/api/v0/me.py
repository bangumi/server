from fastapi import Depends, APIRouter
from pydantic import Field, BaseModel

from pol import res
from pol.res import ErrorDetail
from pol.router import ErrorCatchRoute
from pol.permission import UserGroup
from pol.api.v0.depends.auth import User, get_current_user

router = APIRouter(tags=["用户"], route_class=ErrorCatchRoute)


class Avatar(BaseModel):
    large: str
    medium: str
    small: str

    @classmethod
    def from_record(cls, s):
        if not s:
            return cls(large="", medium="", small="")
        return cls(
            large="https://lain.bgm.tv/pic/user/l/" + s,
            medium="https://lain.bgm.tv/pic/user/m/" + s,
            small="https://lain.bgm.tv/pic/user/s/" + s,
        )


class Me(BaseModel):
    id: int
    url: str
    username: str = Field(..., description="唯一用户名，初始与uid相同，可修改")
    nickname: str
    user_group: UserGroup
    avatar: Avatar
    sign: str


@router.get(
    "/me",
    response_model=Me,
    description="返回当前 Access Token 对应的用户信息",
    responses={
        403: res.response(model=ErrorDetail, description="unauthorized"),
    },
)
async def get_user(
    user: User = Depends(get_current_user),
):
    d = user.dict(by_alias=False)
    d["avatar"] = Avatar.from_record(user.avatar)
    d["url"] = "https://bgm.tv/user/" + user.username
    d["user_group"] = user.group_id

    return d
