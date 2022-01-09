from typing import Any, List, Tuple
from asyncio import gather

from fastapi import Path, Depends, APIRouter
from sqlalchemy.ext.asyncio import AsyncSession

from pol import sa, res, curd
from pol.res import ErrorDetail, not_found_exception
from pol.router import ErrorCatchRoute
from pol.depends import get_db
from pol.db.const import RevisionType
from pol.db.tables import ChiiMember, ChiiRevText, ChiiRevHistory, ChiiSubjectRevision
from pol.api.v0.models import Paged, Pager
from pol.curd.exceptions import NotFoundError
from pol.http_cache.depends import CacheControl
from pol.api.v0.models.revision import Revision, DetailedRevision

router = APIRouter(prefix="/revisions", tags=["编辑历史"], route_class=ErrorCatchRoute)

person_rev_type_filters = ChiiRevHistory.rev_type.in_(RevisionType.person_rev_types())

character_rev_type_filters = ChiiRevHistory.rev_type.in_(
    RevisionType.character_rev_types()
)

subject_rev_type_filters = ChiiRevHistory.rev_type.in_(RevisionType.subject_rev_types())

episode_rev_type_filters = ChiiRevHistory.rev_type.in_(RevisionType.episode_rev_types())


async def get_revisions(
    db: AsyncSession,
    filters: List[Any],
    page: Pager,
):
    total = await curd.count(db, ChiiRevHistory.rev_id, *filters)

    page.check(total)

    columns = [
        ChiiRevHistory.rev_id,
        ChiiRevHistory.rev_type,
        ChiiRevHistory.rev_creator,
        ChiiRevHistory.rev_dateline,
        ChiiRevHistory.rev_edit_summary,
        ChiiMember.username,
        ChiiMember.nickname,
    ]

    query = (
        sa.select(
            *columns,
        )
        .join(ChiiMember, ChiiRevHistory.rev_creator == ChiiMember.uid)
        .where(*filters)
        .order_by(ChiiRevHistory.rev_id.desc())
        .limit(page.limit)
        .offset(page.offset)
    )

    revisions = [
        {
            "id": r["rev_id"],
            "type": r["rev_type"],
            "created_at": r["rev_dateline"],
            "summary": r["rev_edit_summary"],
            "creator": {
                "username": r["username"],
                "nickname": r["nickname"],
            },
        }
        for r in (await db.execute(query)).mappings().fetchall()
    ]
    return {
        "limit": page.limit,
        "offset": page.offset,
        "data": revisions,
        "total": total,
    }


async def get_revision(
    db: AsyncSession,
    filters: List[Any],
):
    r = await curd.get_one(
        db,
        ChiiRevHistory,
        *filters,
    )
    results: Tuple[ChiiMember, ChiiRevText] = await gather(
        curd.get_one(
            db,
            ChiiMember,
            ChiiMember.uid == r.rev_creator,
        ),
        curd.get_one(
            db,
            ChiiRevText,
            ChiiRevText.rev_text_id == r.rev_text_id,
        ),
    )
    user, text_item = results
    return {
        "id": r.rev_id,
        "type": r.rev_type,
        "created_at": r.rev_dateline,
        "summary": r.rev_edit_summary,
        "data": text_item.rev_text,
        "creator": {
            "username": user.username,
            "nickname": user.nickname,
        },
    }


@router.get(
    "/persons",
    response_model=Paged[Revision],
    response_model_exclude_unset=True,
)
async def get_person_revisions(
    person_id: int = 0,
    db: AsyncSession = Depends(get_db),
    page: Pager = Depends(),
    cache_control: CacheControl = Depends(CacheControl),
):
    cache_control(300)
    filters = [person_rev_type_filters]
    if person_id > 0:
        filters.append(ChiiRevHistory.rev_mid == person_id)
    return await get_revisions(
        db,
        filters,
        page,
    )


@router.get(
    "/persons/{revision_id}",
    response_model=DetailedRevision,
    response_model_exclude_unset=True,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_person_revision(
    db: AsyncSession = Depends(get_db),
    revision_id: int = Path(..., gt=0),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(not_found_exception),
):
    cache_control(300)
    try:
        return await get_revision(
            db,
            [ChiiRevHistory.rev_id == revision_id, person_rev_type_filters],
        )
    except NotFoundError:
        raise not_found


@router.get(
    "/characters",
    response_model=Paged[Revision],
    response_model_exclude_unset=True,
)
async def get_character_revisions(
    character_id: int = 0,
    db: AsyncSession = Depends(get_db),
    page: Pager = Depends(),
    cache_control: CacheControl = Depends(CacheControl),
):
    cache_control(300)
    filters = [character_rev_type_filters]
    if character_id > 0:
        filters.append(ChiiRevHistory.rev_mid == character_id)
    return await get_revisions(
        db,
        filters,
        page,
    )


@router.get(
    "/characters/{revision_id}",
    response_model=DetailedRevision,
    response_model_exclude_unset=True,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_character_revision(
    db: AsyncSession = Depends(get_db),
    revision_id: int = Path(..., gt=0),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(not_found_exception),
):
    cache_control(300)
    try:
        return await get_revision(
            db, [ChiiRevHistory.rev_id == revision_id, character_rev_type_filters]
        )
    except NotFoundError:
        raise not_found


@router.get(
    "/subjects",
    response_model=Paged[Revision],
    response_model_exclude_unset=True,
)
async def get_subject_revisions(
    subject_id: int = 0,
    db: AsyncSession = Depends(get_db),
    page: Pager = Depends(),
    cache_control: CacheControl = Depends(CacheControl),
):
    cache_control(300)

    filters = []
    if subject_id > 0:
        filters.append(ChiiSubjectRevision.rev_subject_id == subject_id)
    total = await curd.count(db, ChiiSubjectRevision.rev_subject_id, *filters)
    page.check(total)

    query = (
        sa.select(
            ChiiSubjectRevision.rev_id.label("rev_id"),
            ChiiSubjectRevision.rev_type,
            ChiiSubjectRevision.rev_creator,
            ChiiSubjectRevision.rev_dateline,
            ChiiSubjectRevision.rev_edit_summary,
            ChiiMember.username,
            ChiiMember.nickname,
        )
        .join(
            ChiiMember,
            ChiiSubjectRevision.rev_creator == ChiiMember.uid,
        )
        .where(*filters)
        .order_by(ChiiSubjectRevision.rev_dateline.desc())
        .limit(page.limit)
        .offset(page.offset)
    )

    revisions = [
        {
            "id": r["rev_id"],
            "type": r["rev_type"],
            "created_at": r["rev_dateline"],
            "summary": r["rev_edit_summary"],
            "creator": {
                "username": r["username"],
                "nickname": r["nickname"],
            },
        }
        for r in (await db.execute(query)).mappings().fetchall()
    ]
    return {
        "limit": page.limit,
        "offset": page.offset,
        "data": revisions,
        "total": total,
    }
