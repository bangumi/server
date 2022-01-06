from typing import Any, List, Tuple, Optional
from asyncio import gather

from fastapi import Path, Depends, APIRouter
from starlette.responses import Response
from sqlalchemy.ext.asyncio import AsyncSession

from pol import sa, res, curd
from pol.models import ErrorDetail
from pol.router import ErrorCatchRoute
from pol.depends import get_db, get_redis
from pol.db.const import RevisionType
from pol.db.tables import ChiiMember, ChiiRevText, ChiiRevHistory
from pol.api.v0.utils import raise_not_found, raise_offset_over_total
from pol.api.v0.models import Paged, Pager
from pol.curd.exceptions import NotFoundError
from pol.redis.json_cache import JSONRedis
from pol.http_cache.depends import CacheControl
from pol.api.v0.models.revision import Revision, DetailedRevision

router = APIRouter(prefix="/revisions", tags=["编辑历史"], route_class=ErrorCatchRoute)


person_rev_type_filters = ChiiRevHistory.rev_type.in_(RevisionType.person_rev_types())

character_rev_type_filters = ChiiRevHistory.rev_type.in_(
    RevisionType.character_rev_types()
)


async def get_revisions(
    db: AsyncSession,
    filters: List[Any],
    page: Pager,
):
    total = await curd.count(db, ChiiRevHistory.rev_id, *filters)
    if total <= page.offset:
        raise_offset_over_total(total)
    columns = [
        ChiiRevHistory.rev_id,
        ChiiRevHistory.rev_type,
        ChiiRevHistory.rev_creator,
        ChiiRevHistory.rev_dateline,
        ChiiRevHistory.rev_edit_summary,
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
            "timestamp": r["rev_dateline"],
            "summary": r["rev_edit_summary"],
            "creator": {
                "id": r["rev_creator"],
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
    details: Optional[Any] = None,
):
    r = await curd.get_one(
        db,
        ChiiRevHistory,
        *filters,
        details=details,
    )
    results: Tuple[ChiiMember, ChiiRevText] = await gather(
        curd.get_one(
            db,
            ChiiMember,
            ChiiMember.uid == r.rev_creator,
            details={"rev_creator": r.rev_creator},
        ),
        curd.get_one(
            db,
            ChiiRevText,
            ChiiRevText.rev_text_id == r.rev_text_id,
            details={"rev_text_id": r.rev_text_id},
        ),
    )
    user, text_item = results
    return {
        "id": r.rev_id,
        "type": r.rev_type,
        "timestamp": r.rev_dateline,
        "summary": r.rev_edit_summary,
        "data": text_item.rev_text,
        "creator": {
            "id": r.rev_creator,
            "nickname": user.nickname,
            "avatar": user.avatar,
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
    response: Response,
    db: AsyncSession = Depends(get_db),
    revision_id: int = Path(..., gt=0),
    redis: JSONRedis = Depends(get_redis),
    cache_control: CacheControl = Depends(CacheControl),
):
    cache_control(300)
    cache_key = f"persons:revision:{revision_id}"
    data = await redis.get(cache_key)
    response.headers["x-cache-status"] = "miss" if data is None else "hit"
    if data is None:
        try:
            data = await get_revision(
                db,
                [
                    ChiiRevHistory.rev_id == revision_id,
                    person_rev_type_filters,
                ],
                details={
                    "rev_id": revision_id,
                },
            )
            await redis.set_json(cache_key, data, ex=60)
        except NotFoundError as e:
            raise_not_found(e.details)
    return data


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
    response: Response,
    db: AsyncSession = Depends(get_db),
    revision_id: int = Path(..., gt=0),
    redis: JSONRedis = Depends(get_redis),
    cache_control: CacheControl = Depends(CacheControl),
):
    cache_control(300)
    cache_key = f"characters:revision:{revision_id}"
    data = await redis.get(cache_key)
    response.headers["x-cache-status"] = "miss" if data is None else "hit"
    if data is None:
        try:
            data = await get_revision(
                db,
                [
                    ChiiRevHistory.rev_id == revision_id,
                    character_rev_type_filters,
                ],
                details={
                    "rev_id": revision_id,
                },
            )
            await redis.set_json(cache_key, data, ex=60)
        except NotFoundError as e:
            raise_not_found(e.details)
    return data
