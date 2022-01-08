import datetime
from typing import List, Iterator

from fastapi import Depends, APIRouter
from pydantic import BaseModel
from starlette.requests import Request
from sqlalchemy.ext.asyncio import AsyncSession

from pol import sa, res
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


@router.get(
    "/user/{username}/collections",
    summary="获取用户收藏",
    description="获取对应用户的收藏，查看私有收藏需要access token。",
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
    if username == user.get_username():
        where = []
    else:
        user = await db.scalar(sa.get(ChiiMember, ChiiMember.username == username))
        if not user:
            raise res.not_found(request)

    total = await db.scalar(
        sa.select(sa.count(1)).where(ChiiSubjectInterest.uid == username, *where)
    )

    collections: Iterator[ChiiSubjectInterest] = await db.scalars(
        sa.select(ChiiSubjectInterest)
        .where(ChiiSubjectInterest.uid == username, *where)
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
