from fastapi import Query, Depends, APIRouter
from pydantic import Field, BaseModel
from databases import Database

from pol import sa, res, curd
from pol.models import ErrorDetail
from pol.curd.ep import Ep
from pol.depends import get_db
from pol.db.const import EpType
from pol.db.tables import ChiiEpisode
from pol.api.v0.const import NotFoundDescription
from pol.api.v0.models import Paged
from pol.curd.exceptions import NotFoundError
from pol.api.v0.models.subject import Episode, EpisodeDetail

router = APIRouter(tags=["章节"])


class Pager(BaseModel):
    limit: int = Field(100, gt=0, le=200, description="最大值`200`")
    offset: int = Field(0, ge=0)


@router.get(
    "/episodes",
    response_model=Paged[Episode],
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
    if total == 0:
        raise res.HTTPException(
            status_code=404,
            title="Not Found",
            description=NotFoundDescription,
            detail={"subject_id": subject_id, "type": type},
        )

    first_episode = (
        await curd.ep.get_many(db, ChiiEpisode.ep_subject_id == subject_id, limit=1)
    )[0]

    return {
        "total": total,
        "limit": page.limit,
        "offset": page.offset,
        "data": [
            add_episode(x, first_episode.sort)
            for x in await curd.ep.get_many(
                db,
                *where,
                limit=page.limit,
                offset=page.offset,
            )
        ],
    }


def add_episode(e: Ep, start: float) -> dict:
    data = e.dict()
    if e.type == 0:
        data["ep"] = e.sort - start + 1
    else:
        data["ep"] = 0
    return data


@router.get(
    "/episodes/{episode_id}",
    response_model=EpisodeDetail,
    responses={
        404: res.response(model=ErrorDetail),
    },
    tags=["章节"],
)
async def get_episode(
    episode_id: int,
    db: Database = Depends(get_db),
):
    try:
        ep = await curd.ep.get_one(db, episode_id, ChiiEpisode.ep_ban == 0)
        first_episode = await curd.ep.get_many(
            db, ChiiEpisode.ep_subject_id == ep.subject_id, limit=1
        )
        return add_episode(ep, first_episode[0].sort)
    except (NotFoundError, IndexError):
        raise res.HTTPException(
            status_code=404,
            title="Not Found",
            description=NotFoundDescription,
            detail={"episode_id": episode_id},
        )
