import datetime
from typing import List, Iterator, Optional

from fastapi import Depends, APIRouter
from pydantic import BaseModel
from starlette.requests import Request
from sqlalchemy.ext.asyncio import AsyncSession

from pol import sa, res, permission
from pol.models import ErrorDetail
from pol.router import ErrorCatchRoute
from pol.depends import get_db
from pol.db.const import SubjectType, CollectionType
from pol.db.tables import ChiiMember, ChiiSubjectInterest
from pol.permission import Role
from pol.api.v0.models import Paged, Pager
from pol.api.v0.depends.auth import optional_user

router = APIRouter(tags=["用户"], route_class=ErrorCatchRoute)


class Model(BaseModel):
    subject_id: int
    subject_type: SubjectType
    rate: int
    type: CollectionType
    has_comment: bool
    comment: str
    tag: List[str]
    ep_status: int
    vol_status: int
    updated_at: datetime.datetime
    private: int


async def get_user(db: AsyncSession, username: str) -> Optional[ChiiMember]:
    if user_id := permission.is_user_id(username):  # raw user id like `1`
        return await db.scalar(sa.get(ChiiMember, ChiiMember.uid == user_id))
    else:
        return await db.scalar(sa.get(ChiiMember, ChiiMember.username == username))


@router.get(
    "/user/{username}/collections",
    summary="获取用户收藏",
    description="获取对应用户的收藏，查看私有收藏需要access token。\n设置了用户名的用户无法使用数字ID进行查询",
    response_model=Paged[Model],
    responses={
        404: res.response(model=ErrorDetail, description="用户不存在"),
    },
)
async def get_subject(
    username: str,
    request: Request,
    user: Role = Depends(optional_user),
    page: Pager = Depends(),
    db: AsyncSession = Depends(get_db),
):
    where = [ChiiSubjectInterest.private == 0]

    if user_id := permission.is_user_id(username):  # raw user id like `1`
        if user_id == user.get_user_id():
            where = []
    else:
        if username == user.get_username():
            where = []

    u = await get_user(db, username)
    if not u:
        raise res.not_found(request)

    where.append(ChiiSubjectInterest.uid == u.uid)
    total = await db.scalar(sa.select(sa.count(1)).where(*where))

    collections: Iterator[ChiiSubjectInterest] = await db.scalars(
        sa.select(ChiiSubjectInterest)
        .where(*where)
        .limit(page.limit)
        .offset(page.offset)
        .order_by(ChiiSubjectInterest.last_touch)
    )

    return {
        "total": total,
        **page.dict(),
        "data": [
            {
                "subject_id": x.subject_id,
                "subject_type": x.subject_type,
                "rate": x.rate,
                "type": x.type,
                "has_comment": x.has_comment,
                "comment": x.comment,
                "tag": tags(x.tag),
                "ep_status": x.ep_status,
                "vol_status": x.vol_status,
                "updated_at": x.last_touch,
                "private": x.private,
            }
            for x in collections
        ],
    }


def tags(s: str) -> List[str]:
    return [x.strip() for x in s.split()]
