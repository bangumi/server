from typing import Any, List, Tuple, Optional
from asyncio import gather
from functools import reduce

from fastapi import Path, Depends, APIRouter
from databases import Database
from starlette.responses import Response

from pol import sa, res, curd
from pol.models import ErrorDetail
from pol.depends import get_db, get_redis
from pol.db.const import RevisionType
from pol.db.tables import ChiiMember, ChiiRevText, ChiiRevHistory
from pol.api.v0.utils import raise_not_found, raise_offset_over_total
from pol.api.v0.models import Order, Paged, Pager
from pol.curd.exceptions import NotFoundError
from pol.redis.json_cache import JSONRedis, cache
from pol.api.v0.models.revision import Revision, DetailedRevision

router = APIRouter(prefix="/revisions", tags=["编辑历史"])

person_rev_types = [
    t.value
    for t in [
        RevisionType.person,
        RevisionType.person_cast_relation,
        RevisionType.person_subject_relation,
        RevisionType.person_erase,
        RevisionType.person_merge,
    ]
]

person_rev_type_filters = ChiiRevHistory.rev_type.in_(person_rev_types)

character_rev_types = [
    t.value
    for t in [
        RevisionType.character,
        RevisionType.character_subject_relation,
        RevisionType.character_cast_relation,
        RevisionType.character_erase,
        RevisionType.character_merge,
    ]
]
character_rev_type_filters = ChiiRevHistory.rev_type.in_(character_rev_types)


async def get_revisions(
    db: Database,
    filters: List[Any],
    columns: List[Any],
    join_args: List[Tuple[Any, Any]],
    page: Pager,
    uid: int = 0,
    order: Order = Order.desc,
):
    filters = [*filters]
    if uid > 0:
        filters.append(ChiiRevHistory.rev_creator == uid)
    total = await curd.count(db, ChiiRevHistory.rev_id, *filters)
    if total <= page.offset:
        raise_offset_over_total(total)
    columns = [
        ChiiRevHistory.rev_id,
        ChiiRevHistory.rev_type,
        ChiiRevHistory.rev_creator,
        ChiiRevHistory.rev_dateline,
        ChiiRevHistory.rev_edit_summary,
        *columns,
    ]
    join_args = [*join_args]
    if uid <= 0:
        columns.extend([ChiiMember.nickname, ChiiMember.avatar])
        join_args.append((ChiiMember, ChiiRevHistory.rev_creator == ChiiMember.uid))
    sort_field = ChiiRevHistory.rev_id
    if order == Order.desc:
        sort_field.desc()
    else:
        sort_field.asc()
    query = (
        sa.select(
            *columns,
        )
        .where(*filters)
        .order_by(sort_field)
        .limit(page.limit)
        .offset(page.offset)
    )
    query = reduce(lambda acc, item: acc.join(item[0], item[1]), join_args, query)
    revisions = [
        {
            "id": r["rev_id"],
            "type": r["rev_type"],
            "timestamp": r["rev_dateline"],
            "summary": r["rev_edit_summary"],
            "creator": {
                "id": r["rev_creator"],
                **(
                    {
                        "nickname": r["nickname"],
                        "avatar": r["avatar"],
                    }
                    if uid <= 0
                    else {}
                ),
            },
        }
        for r in await db.fetch_all(query)
    ]
    return {
        "limit": page.limit,
        "offset": page.offset,
        "data": revisions,
        "total": total,
    }


async def get_revision(
    db: Database,
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
    db: Database = Depends(get_db),
    page: Pager = Depends(),
    uid: int = 0,
    order: Order = Order.desc,
):
    filters = [person_rev_type_filters]
    if person_id > 0:
        filters.append(ChiiRevHistory.rev_mid == person_id)
    return await get_revisions(
        db,
        filters,
        [],
        [],
        page,
        uid=uid,
        order=order,
    )


@router.get(
    "/persons/{revision_id}",
    response_model=DetailedRevision,
    response_model_exclude_unset=True,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
@cache(lambda revision_id, **_: f"persons:revision:{revision_id}")
async def get_person_revision(
    response: Response,
    db: Database = Depends(get_db),
    revision_id: int = Path(..., gt=0),
    redis: JSONRedis = Depends(get_redis),
):

    try:
        data = await get_revision(
            db,
            [
                ChiiRevHistory.rev_id == revision_id,
                person_rev_type_filters,
            ],
            details={
                "rev_id": revision_id,
                "rev_type": f"({','.join([str(t) for t in person_rev_types])})",
            },
        )
        return data
    except NotFoundError as e:
        raise_not_found(e.details)


@router.get(
    "/characters",
    response_model=Paged[Revision],
    response_model_exclude_unset=True,
)
async def get_character_revisions(
    character_id: int = 0,
    db: Database = Depends(get_db),
    page: Pager = Depends(),
    uid: int = 0,
    order: Order = Order.desc,
):
    filters = [character_rev_type_filters]
    if character_id > 0:
        filters.append(ChiiRevHistory.rev_mid == character_id)
    return await get_revisions(
        db,
        filters,
        [],
        [],
        page,
        uid=uid,
        order=order,
    )


@router.get(
    "/characters/{revision_id}",
    response_model=DetailedRevision,
    response_model_exclude_unset=True,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
@cache(lambda revision_id, **_: f"characters:revision:{revision_id}")
async def get_character_revision(
    response: Response,
    db: Database = Depends(get_db),
    revision_id: int = Path(..., gt=0),
    redis: JSONRedis = Depends(get_redis),
):
    try:
        data = await get_revision(
            db,
            [
                ChiiRevHistory.rev_id == revision_id,
                character_rev_type_filters,
            ],
            details={
                "rev_id": revision_id,
                "rev_type": f"({','.join([str(t) for t in character_rev_types])})",
            },
        )
        return data
    except NotFoundError as e:
        raise_not_found(e.details)
