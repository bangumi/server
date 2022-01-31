from typing import Optional

from fastapi import Depends, APIRouter

from pol import res
from pol.res import ErrorDetail
from pol.router import ErrorCatchRoute
from pol.db.const import SubjectType
from pol.api.v0.models import Paged, Pager
from pol.api.v0.depends import get_index
from pol.services.user_service import UserService
from pol.services.index_service import Index as _Index
from pol.services.index_service import IndexService
from pol.services.subject_service import SubjectService
from .models.index import Index, IndexSubject

router = APIRouter(prefix="/indices", tags=["目录"], route_class=ErrorCatchRoute)


@router.get(
    "/{index_id}",
    response_model=Index,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_index_by_id(
    r: _Index = Depends(get_index),
    user_service: UserService = Depends(UserService.new),
):
    creator = await user_service.get_by_uid(r.creator_id)
    return {
        "id": r.id,
        "title": r.title,
        "desc": r.desc,
        "stat": r.stat,
        "total": r.total,
        "created_at": r.created_at,
        "updated_at": r.updated_at,
        "creator": creator,
        "ban": r.ban,
    }


@router.get(
    "/{index_id}/subjects",
    response_model=Paged[IndexSubject],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_index_subjects(
    r: _Index = Depends(get_index),
    type: Optional[SubjectType] = None,
    index_service: IndexService = Depends(IndexService.new),
    subject_service: SubjectService = Depends(SubjectService.new),
    page: Pager = Depends(),
):
    total = await index_service.count_index_subjects(r.id, type)
    page.check(total)
    results = await index_service.list_index_subjects(
        r.id, page.limit, page.offset, type
    )
    subjects = await subject_service.get_by_ids(
        *(x.id for x in results), include_nsfw=True
    )

    return {
        "limit": page.limit,
        "offset": page.offset,
        "data": [
            IndexSubject(
                id=r.id,
                type=s.type,
                infobox=s.infobox,
                name=s.name,
                images=s.image,
                date=s.date,
                comment=r.comment,
                added_at=r.added_at,
            )
            for r, s in ((r, subjects[r.id]) for r in results)
        ],
        "total": total,
    }
