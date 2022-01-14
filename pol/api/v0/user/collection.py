import datetime
from typing import List, Iterator, Optional

from fastapi import Query, Depends, APIRouter
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession

from pol import res
from pol.db import sa
from pol.res import ErrorDetail
from pol.models import PublicUser
from pol.router import ErrorCatchRoute
from pol.depends import get_db
from pol.db.const import SubjectType, CollectionType
from pol.db.tables import ChiiSubjectInterest
from pol.permission import Role
from pol.api.v0.models import Paged, Pager
from pol.api.v0.depends import get_public_user
from pol.api.v0.depends.auth import optional_user

router = APIRouter(
    tags=["用户"],
    route_class=ErrorCatchRoute,
    redirect_slashes=False,
)


class UserCollection(BaseModel):
    subject_id: int
    subject_type: SubjectType
    rate: int
    type: CollectionType
    comment: Optional[str]
    tags: List[str]
    ep_status: int
    vol_status: int
    updated_at: datetime.datetime
    private: bool


@router.get(
    "/users/{username}/collections",
    summary="获取用户收藏",
    description="获取对应用户的收藏，查看私有收藏需要access token。",
    response_model=Paged[UserCollection],
    responses={
        404: res.response(model=ErrorDetail, description="用户不存在"),
    },
)
@router.get(
    "/user/{username}/collections",
    include_in_schema=False,
    response_model=Paged[UserCollection],
)
async def get_user_collection(
    user: Role = Depends(optional_user),
    page: Pager = Depends(),
    u: PublicUser = Depends(get_public_user),
    db: AsyncSession = Depends(get_db),
    subject_type: SubjectType = Query(
        None, description="条目类型，默认为全部\n\n具体含义见 [SubjectType](#model-SubjectType)"
    ),
    type: CollectionType = Query(
        None, description="收藏类型，默认为全部\n\n具体含义见 [CollectionType](#model-CollectionType)"
    ),
):
    where = [ChiiSubjectInterest.private == 0]

    if u.id == user.get_user_id():
        where = []

    if subject_type is not None:
        where.append(ChiiSubjectInterest.subject_type == subject_type)

    if type is not None:
        where.append(ChiiSubjectInterest.type == type)

    where.append(ChiiSubjectInterest.user_id == u.id)
    total = await db.scalar(sa.select(sa.count(1)).where(*where))

    page.check(total)

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
                "tags": tags(x.tag),
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
