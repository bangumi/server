import datetime
from typing import List, Iterator, Optional

from fastapi import Depends, APIRouter
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession

from pol import sa, res
from pol.models import PublicUser, ErrorDetail
from pol.router import ErrorCatchRoute
from pol.depends import get_db
from pol.db.const import SubjectType, CollectionType
from pol.db.tables import ChiiSubjectInterest
from pol.permission import Role
from pol.api.v0.models import Paged, Pager
from pol.api.v0.depends import get_public_user
from pol.api.v0.depends.auth import optional_user

router = APIRouter(
    tags=["用户", "社区"],
    route_class=ErrorCatchRoute,
    redirect_slashes=False,
)


class UserCollection(BaseModel):
    subject_id: int
    subject_type: SubjectType
    rate: int
    type: CollectionType
    comment: Optional[str]
    tag: List[str]
    ep_status: int
    vol_status: int
    updated_at: datetime.datetime
    private: int


@router.get(
    "/user/{username}/collections",
    summary="获取用户收藏",
    description="获取对应用户的收藏，查看私有收藏需要access token。\n设置了 username 的用户无法通过数字ID查询",
    response_model=Paged[UserCollection],
    responses={
        404: res.response(model=ErrorDetail, description="用户不存在"),
    },
)
async def get_subject(
    user: Role = Depends(optional_user),
    page: Pager = Depends(),
    u: PublicUser = Depends(get_public_user),
    db: AsyncSession = Depends(get_db),
):
    where = [ChiiSubjectInterest.private == 0]

    if u.id == user.get_user_id():
        where = []

    where.append(ChiiSubjectInterest.uid == u.id)
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
                "comment": x.comment if x.has_comment else None,
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
