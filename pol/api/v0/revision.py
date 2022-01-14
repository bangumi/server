from loguru import logger
from fastapi import Path, Query, Depends, APIRouter

from pol import res
from pol.res import ErrorDetail, not_found_exception
from pol.router import ErrorCatchRoute
from pol.db.const import RevisionType
from pol.api.v0.models import Paged, Pager
from pol.curd.exceptions import NotFoundError
from pol.http_cache.depends import CacheControl
from pol.services.rev_service import RevisionService
from pol.services.user_service import UserService
from pol.api.v0.models.revision import Revision, DetailedRevision

router = APIRouter(prefix="/revisions", tags=["编辑历史"], route_class=ErrorCatchRoute)


@router.get(
    "/persons",
    response_model=Paged[Revision],
    response_model_exclude_unset=True,
)
async def get_person_revisions(
    person_id: int = Query(..., gt=0),
    rev_service: RevisionService = Depends(RevisionService.new),
    user_service: UserService = Depends(UserService.new),
    page: Pager = Depends(),
    cache_control: CacheControl = Depends(CacheControl),
):
    cache_control(300)
    total = await rev_service.count_person_history(person_id)
    page.check(total)

    results = await rev_service.list_person_history(
        person_id, limit=page.limit, offset=page.offset
    )

    users = await user_service.get_users_by_id(x.creator_id for x in results)

    revisions = [
        {
            "id": r.id,
            "type": r.type,
            "created_at": r.created_at,
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
    "/persons/{revision_id}",
    response_model=DetailedRevision,
    response_model_exclude_unset=True,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_person_revision(
    revision_id: int = Path(..., gt=0),
    rev_service: RevisionService = Depends(RevisionService.new),
    user_service: UserService = Depends(UserService.new),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(not_found_exception),
):
    cache_control(300)
    try:
        rev = await rev_service.get_person_history(revision_id)
    except RevisionService.NotFoundError:
        raise not_found
    try:
        user = await user_service.get_by_uid(rev.creator_id)
    except UserService.NotFoundError:
        logger.error(
            "can't find user {uid} for revision {rev_id}",
            uid=rev.creator_id,
            rev_id=revision_id,
        )
        raise ValueError("can't find user")
    return {
        "id": revision_id,
        "type": rev.type,
        "creator": user,
        "summary": rev.summary,
        "created_at": rev.created_at,
        "data": rev.data,
    }


@router.get(
    "/characters",
    response_model=Paged[Revision],
    response_model_exclude_unset=True,
)
async def get_character_revisions(
    character_id: int = Query(..., gt=0),
    rev_service: RevisionService = Depends(RevisionService.new),
    user_service: UserService = Depends(UserService.new),
    page: Pager = Depends(),
    cache_control: CacheControl = Depends(CacheControl),
):
    cache_control(300)
    total = await rev_service.count_character_history(character_id)
    page.check(total)

    results = await rev_service.list_character_history(
        character_id, limit=page.limit, offset=page.offset
    )

    users = await user_service.get_users_by_id(x.creator_id for x in results)

    revisions = [
        {
            "id": r.id,
            "type": r.type,
            "created_at": r.created_at,
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
    "/characters/{revision_id}",
    response_model=DetailedRevision,
    response_model_exclude_unset=True,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_character_revision(
    revision_id: int = Path(..., gt=0),
    rev_service: RevisionService = Depends(RevisionService.new),
    user_service: UserService = Depends(UserService.new),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(not_found_exception),
):
    cache_control(300)
    try:
        rev = await rev_service.get_character_history(revision_id)
    except RevisionService.NotFoundError:
        raise not_found
    try:
        user = await user_service.get_by_uid(rev.creator_id)
    except UserService.NotFoundError:
        logger.error(
            "can't find user {uid} for revision {rev_id}",
            uid=rev.creator_id,
            rev_id=revision_id,
        )
        raise ValueError("can't find user")
    return {
        "id": revision_id,
        "type": rev.type,
        "creator": user,
        "summary": rev.summary,
        "created_at": rev.created_at,
        "data": rev.data,
    }


@router.get(
    "/subjects",
    response_model=Paged[Revision],
    response_model_exclude_unset=True,
)
async def get_subject_revisions(
    subject_id: int = Query(..., gt=0),
    rev_service: RevisionService = Depends(RevisionService.new),
    user_service: UserService = Depends(UserService.new),
    page: Pager = Depends(),
    cache_control: CacheControl = Depends(CacheControl),
):
    cache_control(300)

    total = await rev_service.count_subject_history(subject_id)

    page.check(total)

    results = await rev_service.list_subject_history(
        subject_id, offset=page.offset, limit=page.limit
    )

    users = await user_service.get_users_by_id(
        x.creator_id for x in results if x.creator_id > 0
    )

    revisions = [
        {
            "id": r.id,
            "type": r.type,
            "created_at": r.created_at,
            "summary": r.summary,
            "creator": users[r.creator_id] if r.creator_id else None,
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
    "/subjects/{revision_id}",
    response_model=DetailedRevision,
    response_model_exclude_unset=True,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_revision(
    revision_id: int = Path(..., gt=0),
    rev_service: RevisionService = Depends(RevisionService.new),
    user_service: UserService = Depends(UserService.new),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(not_found_exception),
):
    cache_control(300)

    try:
        r = await rev_service.get_subject_history(revision_id)
    except RevisionService.NotFoundError:
        raise not_found

    if r.creator_id == 0:
        creator = None
    else:
        try:
            creator = await user_service.get_by_uid(r.creator_id)
        except NotFoundError:
            logger.error(
                "subject revision {rev_id} creator {user_id} does not exist",
                rev_id=revision_id,
                user_id=r.creator_id,
            )
            raise not_found

    return {
        "id": r.id,
        "type": r.type,
        "created_at": r.created_at,
        "summary": r.summary,
        "data": r.data,
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
            "episode revision {id} creator {uid} does not exist",
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
