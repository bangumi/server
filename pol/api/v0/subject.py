from typing import List

from fastapi import Path, Query, Depends, APIRouter
from pydantic import Field, BaseModel
from databases import Database
from starlette.responses import Response, RedirectResponse

from pol import res, curd, wiki
from pol.config import CACHE_KEY_PREFIX
from pol.models import ErrorDetail
from pol.depends import get_db, get_redis
from pol.db.const import PLATFORM_MAP, RELATION_MAP, EpType
from pol.db.tables import ChiiEpisode, ChiiSubject, ChiiSubjectRelations
from pol.db_models import sa
from pol.api.v0.const import NotFoundDescription
from pol.curd.exceptions import NotFoundError
from pol.redis.json_cache import JSONRedis
from pol.api.v0.models.subject import Subject, RelSubject, PagedEpisode

router = APIRouter(tags=["条目"])

api_base = "/v0/subjects"


async def basic_subject(db: Database, subject_id: int, *where) -> curd.subject.Subject:
    try:
        return await curd.subject.get_one(
            db,
            ChiiSubject.subject_id == subject_id,
            *where,
        )
    except NotFoundError:
        raise res.HTTPException(
            status_code=404,
            title="Not Found",
            description=NotFoundDescription,
            detail={"subject_id": subject_id},
        )


@router.get(
    "/subjects/{subject_id}",
    description="cache with 300s",
    response_model_by_alias=False,
    response_model=Subject,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject(
    response: Response,
    db: Database = Depends(get_db),
    subject_id: int = Path(..., gt=0),
    redis: JSONRedis = Depends(get_redis),
):
    cache_key = CACHE_KEY_PREFIX + f"subject:{subject_id}"
    if (value := await redis.get(cache_key)) is not None:
        response.headers["x-cache-status"] = "hit"
        return value
    else:
        response.headers["x-cache-status"] = "miss"

    subject = await basic_subject(db, subject_id)

    if subject.redirect:
        return RedirectResponse(f"{api_base}/{subject.redirect}")

    data = subject.dict(exclude={"dateline", "image", "dropped", "wish"})

    data["locked"] = subject.locked
    data["images"] = subject.images
    data["collection"] = subject.collection
    data["rating"] = subject.rating()
    data["platform"] = PLATFORM_MAP[subject.type].get(
        subject.platform, {"type_cn": ""}
    )["type_cn"]
    data["total_episodes"] = await db.fetch_val(
        sa.select(sa.func.count(ChiiEpisode.ep_id)).where(
            ChiiEpisode.ep_subject_id == subject_id
        )
    )

    try:
        data["infobox"] = wiki.parse(subject.infobox).info
    except wiki.WikiSyntaxError:
        data["infobox"] = None

    await redis.set_json(cache_key, value=data, ex=300)

    return data


class Pager(BaseModel):
    limit: int = Field(100, gt=0, le=200, description="最大值`200`")
    offset: int = Field(0, ge=0)


@router.get(
    "/episodes",
    response_model_by_alias=False,
    response_model=PagedEpisode,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_episodes(
    db: Database = Depends(get_db),
    subject_id: int = Query(..., gt=0),
    type: EpType = Query(None, description="`0`,`1`,`2`,`3`代表`本篇`，`sp`，`op`，`ed`"),
    page: Pager = Depends(),
):
    where = [
        ChiiEpisode.ep_subject_id == subject_id,
    ]

    if type is not None:
        where.append(ChiiEpisode.ep_type == type.value)

    total = await db.fetch_val(sa.select(sa.count(ChiiEpisode.ep_id)).where(*where))

    return {
        "total": total,
        "limit": page.limit,
        "offset": page.offset,
        "data": [
            x.dict()
            for x in await curd.ep.get_many(
                db,
                *where,
                limit=page.limit,
                offset=page.offset,
            )
        ],
    }


@router.get(
    "/subjects/{subject_id}/subjects",
    response_model_by_alias=False,
    response_model=List[RelSubject],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_relations(
    db: Database = Depends(get_db),
    subject_id: int = Path(..., gt=0),
):
    subject = await basic_subject(db, subject_id, ChiiSubject.subject_ban == 0)

    query = (
        sa.select(
            ChiiSubjectRelations, ChiiSubject.subject_name, ChiiSubject.subject_name_cn
        )
        .join(
            ChiiSubject, ChiiSubject.subject_id == ChiiSubjectRelations.rlt_subject_id
        )
        .where(ChiiSubjectRelations.rlt_subject_id == subject_id)
        .order_by(
            ChiiSubjectRelations.rlt_order, ChiiSubjectRelations.rlt_related_subject_id
        )
    )

    response = [
        {
            "id": r["rlt_related_subject_id"],
            "relation": RELATION_MAP[r["rlt_subject_type_id"]][
                r["rlt_relation_type"]
            ].get(),
            "name": r["subject_name"],
            "type": repr(r["rlt_related_subject_type_id"]),
            "name_cn": r["rlt_relation_type"],
            "images": subject.images,
        }
        for r in await db.fetch_all(query)
    ]

    return response
