from typing import Any, List

from loguru import logger
from fastapi import Path, Query, Depends, APIRouter
from sqlalchemy.ext.asyncio import AsyncSession

from pol import res, curd
from pol.db import sa
from pol.res import ErrorDetail, not_found_exception
from pol.router import ErrorCatchRoute
from pol.depends import get_db
from pol.db.const import RevisionType
from pol.db.tables import ChiiMember, ChiiRevText, ChiiRevHistory, ChiiSubjectRevision
from pol.api.v0.models import Paged, Pager
from pol.curd.exceptions import NotFoundError
from pol.http_cache.depends import CacheControl
from pol.services.rev_service import RevisionService
from pol.services.user_service import UserService
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
    user_service: UserService,
    filters: List[Any],
    page: Pager,
):
    total = await curd.count(db, ChiiRevHistory.rev_id, *filters)

    page.check(total)

    query = (
        sa.select(ChiiRevHistory)
        .where(*filters)
        .order_by(ChiiRevHistory.rev_id.desc())
        .limit(page.limit)
        .offset(page.offset)
    )

    results: List[ChiiRevHistory] = list(await db.scalars(query))

    users = await user_service.get_users_by_id(x.rev_creator for x in results)

    revisions = [
        {
            "id": r.rev_id,
            "type": r.rev_type,
            "created_at": r.rev_dateline,
            "summary": r.rev_edit_summary,
            "creator": users.get(r.rev_creator),
        }
        for r in results
    ]
    return {
        "limit": page.limit,
        "offset": page.offset,
        "data": revisions,
        "total": total,
    }


async def get_revision(
    db: AsyncSession,
    user_service: UserService,
    filters: List[Any],
):
    r = await db.scalar(sa.get(ChiiRevHistory, *filters))
    if not r:
        raise NotFoundError

    text_item = await db.get(ChiiRevText, r.rev_text_id)
    if not text_item:
        raise NotFoundError
    user = await user_service.get_by_uid(r.rev_creator)

    return {
        "id": r.rev_id,
        "type": r.rev_type,
        "created_at": r.rev_dateline,
        "summary": r.rev_edit_summary,
        "data": text_item.rev_text,
        "creator": user,
    }


@router.get(
    "/persons",
    response_model=Paged[Revision],
    response_model_exclude_unset=True,
)
async def get_person_revisions(
    person_id: int = 0,
    db: AsyncSession = Depends(get_db),
    user_service: UserService = Depends(UserService.new),
    page: Pager = Depends(),
    cache_control: CacheControl = Depends(CacheControl),
):
    cache_control(300)
    filters = [person_rev_type_filters]
    if person_id > 0:
        filters.append(ChiiRevHistory.rev_mid == person_id)
    return await get_revisions(
        db,
        user_service,
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
    user_service: UserService = Depends(UserService.new),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(not_found_exception),
):
    cache_control(300)
    try:
        return await get_revision(
            db,
            user_service,
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
    user_service: UserService = Depends(UserService.new),
    page: Pager = Depends(),
    cache_control: CacheControl = Depends(CacheControl),
):
    cache_control(300)
    filters = [character_rev_type_filters]
    if character_id > 0:
        filters.append(ChiiRevHistory.rev_mid == character_id)
    return await get_revisions(
        db,
        user_service,
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
    user_service: UserService = Depends(UserService.new),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(not_found_exception),
):
    cache_control(300)
    try:
        return await get_revision(
            db,
            user_service,
            [ChiiRevHistory.rev_id == revision_id, character_rev_type_filters],
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
        .outerjoin(
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
            }
            if r["username"]
            else None,
        }
        for r in (await db.execute(query)).mappings().fetchall()
    ]
    return {
        "limit": page.limit,
        "offset": page.offset,
        "data": revisions,
        "total": total,
    }


@router.get(
    "/subjects/{revision_id}",
    response_model=DetailedRevision,
    response_model_exclude_unset=True,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_revision(
    db: AsyncSession = Depends(get_db),
    revision_id: int = Path(..., gt=0),
    user_service: UserService = Depends(UserService.new),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(not_found_exception),
):
    cache_control(300)

    r = await db.get(ChiiSubjectRevision, revision_id)
    if not r:
        raise not_found

    if r.rev_creator == 0:
        creator = None
    else:
        try:
            creator = await user_service.get_by_uid(r.rev_creator)
        except NotFoundError:
            logger.error(
                f"subject revision {r.rev_id} creator {r.rev_creator} does not exist"
            )
            raise not_found

    return {
        "id": r.rev_id,
        "type": r.rev_type,
        "created_at": r.rev_dateline,
        "summary": r.rev_edit_summary,
        "data": {
            "subject_id": r.rev_subject_id,
            "name": r.rev_name,
            "name_cn": r.rev_name_cn,
            "vote_field": r.rev_vote_field,
            "type": r.rev_type,
            "type_id": r.rev_type_id,
            "field_infobox": r.rev_field_infobox,
            "field_summary": r.rev_field_summary,
            "field_eps": r.rev_field_eps,
            "platform": r.rev_platform,
        },
        "creator": creator,
    }


@router.get(
    "/episodes",
    response_model=Paged[Revision],
    response_model_exclude_unset=True,
)
async def get_episode_revisions(
    episode_id: int = Query(..., gt=0),
    rev_service: RevisionService = Depends(RevisionService.new),
    user_service: UserService = Depends(UserService.new),
    page: Pager = Depends(),
    cache_control: CacheControl = Depends(CacheControl),
):
    cache_control(300)

    total = await rev_service.count_ep_history(episode_id)

    page.check(total)

    results = await rev_service.list_ep_history(
        ep_id=episode_id, limit=page.limit, offset=page.offset
    )

    users = await user_service.get_users_by_id(x.creator_id for x in results)

    revisions = [
        {
            "id": r.id,
            "type": r.type,
            "created_at": r.creator_id,
            "summary": r.summary,
            "creator": users[r.creator_id],
        }
        for r in results
    ]
    return {
        "limit": page.limit,
        "offset": page.offset,
        "data": revisions,
        "total": total,
    }


@router.get(
    "/episodes/{revision_id}",
    response_model=DetailedRevision,
    response_model_exclude_unset=True,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_episode_revision(
    revision_id: int = Path(..., gt=0),
    rev_service: RevisionService = Depends(RevisionService.new),
    user_service: UserService = Depends(UserService.new),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(not_found_exception),
):
    cache_control(300)

    try:
        rev = await rev_service.get_ep_history(revision_id)
    except RevisionService.NotFoundError:
        raise not_found

    try:
        user = await user_service.get_by_uid(rev.creator_id)
    except NotFoundError:
        logger.error(
            f"episode revision {id} creator {uid} does not exist",
            id=rev.id,
            uid=rev.creator_id,
        )
        raise not_found

    return {
        "id": rev.id,
        "type": RevisionType.ep,
        "created_at": rev.created_at,
        "summary": rev.summary,
        "data": {
            "eids": rev.episode_ids,
            "ep_infobox": rev.infobox,
            "subject_id": rev.subject_id,
            "version": rev.version,
        },
        "creator": user,
    }
