from fastapi import Depends, Path
from mypy.typeshed.stdlib.typing import List
from sqlalchemy.ext.asyncio import AsyncSession

from pol import res
from pol.api.v0.depends.auth import optional_user
from pol.api.v0.models.base import ListResponse, CursorPaged
from pol.api.v0.models.topic import Topic
from pol.depends import get_db, get_redis
from pol.http_cache.depends import CacheControl
from pol.models import ErrorDetail
from pol.permission import Role
from pol.redis.json_cache import JSONRedis
from subject import router, exception_404



@router.get(
    "/subjects/{subject_id}/topics",
    response_model=List[Topic],
    responses={
        404: res.response(model=ErrorDetail),
    },
    summary="all get topics related to a subject",
    description="returns the first 100 topics",
    tags=["章节"],
)
async def get_subject_topics(
    response: ListResponse[CursorPaged[Topic]],
    exc_404: res.HTTPException = Depends(exception_404),
    subject_id: int = Path(..., gt=0),
    user: Role = Depends(optional_user),
    db: AsyncSession = Depends(get_db),
    redis: JSONRedis = Depends(get_redis),
    cache_control: CacheControl = Depends(CacheControl),
)
