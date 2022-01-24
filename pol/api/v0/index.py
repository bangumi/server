from typing import Optional

from fastapi import Path, Depends, APIRouter

from pol import res
from pol.res import ErrorDetail, not_found_exception
from pol.router import ErrorCatchRoute
from pol.db.const import SubjectType
from pol.permission import Role
from pol.api.v0.models import Paged, Pager
from pol.http_cache.depends import CacheControl
from pol.api.v0.depends.auth import optional_user
from pol.services.user_service import UserService
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
async def get_index(
    index_id: int = Path(0, gt=0),
    index_service: IndexService = Depends(IndexService.new),
    user_service: UserService = Depends(UserService.new),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(not_found_exception),
    user: Role = Depends(optional_user),
):
    cache_control(300)
    if not user.allow_nsfw() and await index_service.get_index_nsfw_by_id(index_id):
        raise not_found
    try:
        r = await index_service.get_index_by_id(index_id)
    except IndexService.NotFoundError:
        raise not_found
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
    index_id: int = Path(0, gt=0),
    type: Optional[SubjectType] = None,
    index_service: IndexService = Depends(IndexService.new),
    subject_service: SubjectService = Depends(SubjectService.new),
    page: Pager = Depends(),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(not_found_exception),
    user: Role = Depends(optional_user),
):
    cache_control(300)
    if not user.allow_nsfw() and await index_service.get_index_nsfw_by_id(index_id):
        raise not_found
    try:
        await index_service.get_index_by_id(index_id)
    except IndexService.NotFoundError:
        raise not_found
    total = await index_service.count_index_subjects(index_id, type)
    page.check(total)
    results = await index_service.list_index_subjects(
        index_id, page.limit, page.offset, type
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
