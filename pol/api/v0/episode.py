from typing import Optional

from loguru import logger
from fastapi import Query, Depends, APIRouter
from pydantic import Field
from sqlalchemy.ext.asyncio import AsyncSession

from pol import res
from pol.db import sa
from pol.res import ErrorDetail, not_found_exception
from pol.router import ErrorCatchRoute
from pol.depends import get_db
from pol.db.const import EpType
from pol.db.tables import ChiiEpisode, ChiiSubject
from pol.permission import Role
from pol.api.v0.models import Paged
from pol.api.v0.models import Pager as _Pager
from pol.models.subject import Ep
from pol.http_cache.depends import CacheControl
from pol.api.v0.depends.auth import optional_user
from pol.api.v0.models.subject import Episode, EpisodeDetail

router = APIRouter(tags=["章节"], route_class=ErrorCatchRoute)


class Pager(_Pager):
    limit: int = Field(100, gt=0, le=200, description="最大值`200`")


@router.get(
    "/episodes",
    response_model=Paged[Episode],
    responses={
        404: res.response(model=ErrorDetail),
    },
    tags=["章节"],
)
async def get_episodes(
    db: AsyncSession = Depends(get_db),
    subject_id: int = Query(..., gt=0),
    type: EpType = Query(None, description="`0`,`1`,`2`,`3`代表`本篇`，`sp`，`op`，`ed`"),
    page: Pager = Depends(),
    cache_control: CacheControl = Depends(CacheControl),
    user: Role = Depends(optional_user),
):
    subject = await db.get(ChiiSubject, subject_id)
    if not subject:
        cache_control(300)
        return page.dict()

    if subject.subject_nsfw:
        if not user.allow_nsfw():
            return page.dict()
    else:
        cache_control(300)

    where = [
        ChiiEpisode.ep_subject_id == subject_id,
        ChiiEpisode.ep_ban == 0,
    ]

    if type is not None:
        where.append(ChiiEpisode.ep_type == type.value)

    total = await db.scalar(sa.select(sa.count(1)).where(*where))

    if total == 0:
        return page.dict()

    page.check(total)

    first_episode: ChiiEpisode = await db.scalar(
        sa.select(ChiiEpisode)
        .where(
            ChiiEpisode.ep_subject_id == subject_id,
            ChiiEpisode.ep_type == EpType.normal,
            ChiiEpisode.ep_ban == 0,
        )
        .limit(1)
    )

    return {
        "total": total,
        "limit": page.limit,
        "offset": page.offset,
        "data": [
            add_episode(Ep.from_orm(x), first_episode.ep_sort)
            for x in await db.scalars(
                sa.select(ChiiEpisode)
                .where(*where)
                .limit(page.limit)
                .offset(page.offset)
            )
        ],
    }


def add_episode(e: Ep, start: float) -> dict:
    data = e.dict(by_alias=False)
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
    db: AsyncSession = Depends(get_db),
    cache_control: CacheControl = Depends(CacheControl),
    not_found: res.HTTPException = Depends(not_found_exception),
    user: Role = Depends(optional_user),
):
    ep: Optional[ChiiEpisode] = await db.get(ChiiEpisode, episode_id)
    if ep is None:
        cache_control(300)
        raise not_found

    subject = await db.scalar(
        sa.get(ChiiSubject, ChiiSubject.subject_id == ep.ep_subject_id)
    )
    if not subject:  # pragma: no cover
        logger.error(
            "detached episode {}, missing subject {}", ep.ep_id, ep.ep_subject_id
        )
        cache_control(300)
        raise not_found

    if subject.subject_nsfw:
        if not user.allow_nsfw():
            raise not_found
    else:
        cache_control(300)

    first_episode: ChiiEpisode = await db.scalar(
        sa.get(
            ChiiEpisode,
            ChiiEpisode.ep_subject_id == ep.ep_subject_id,
            ChiiEpisode.ep_type == EpType.normal,
            ChiiEpisode.ep_ban == 0,
        )
    )

    if first_episode:
        first = first_episode.ep_sort
    else:
        first = 1

    return add_episode(Ep.from_orm(ep), first)
