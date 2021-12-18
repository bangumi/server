from fastapi import Query, Depends, APIRouter
from pydantic import Field, BaseModel
from databases import Database

from pol import res, curd
from pol.models import ErrorDetail
from pol.depends import get_db
from pol.db.const import EpType
from pol.db.tables import ChiiEpisode
from pol.db_models import sa
from pol.api.v0.const import NotFoundDescription
from pol.curd.exceptions import NotFoundError
from pol.api.v0.models.subject import PagedEpisode, EpisodeDetail

router = APIRouter(tags=["章节"])


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
    tags=["章节"],
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
    "/episodes/{episode_id}",
    response_model_by_alias=False,
    response_model=EpisodeDetail,
    responses={
        404: res.response(model=ErrorDetail),
    },
    tags=["章节"],
)
async def get_episodes(
    episode_id: int,
    db: Database = Depends(get_db),
):
    try:
        return (await curd.ep.get_one(db, episode_id, ChiiEpisode.ep_ban == 0)).dict()
    except NotFoundError:
        raise res.HTTPException(
            status_code=404,
            title="Not Found",
            description=NotFoundDescription,
            detail={"episode_id": episode_id},
        )
