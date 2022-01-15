from fastapi import Path, Depends, APIRouter

from pol.models import Subject
from pol.router import ErrorCatchRoute
from pol.permission import Role
from pol.api.v0.models import Paged, Pager
from pol.api.v0.depends import get_subject
from pol.api.v0.topic.res import Topic, TopicDetail
from pol.api.v0.depends.auth import optional_user

router = APIRouter(
    tags=["社区"],
    route_class=ErrorCatchRoute,
    redirect_slashes=False,
)


@router.get(
    "/subjects/{subject_id}/topics",
    summary="获取条目讨论帖列表",
    response_model=Paged[Topic],
)
async def get_topics(
    page: Pager = Depends(),
    user: Role = Depends(optional_user),
    subject: Subject = Depends(get_subject),
):
    total = ...
    data = ...

    return {
        "total": total,
        **page.dict(),
        "data": data,
    }


@router.get(
    "/subjects/{subject_id}/topics/{topic_id}",
    summary="获取条目讨论帖内容",
    response_model=TopicDetail,
)
async def get_topic(
    topic_id: int = Path(..., gt=0),
    user: Role = Depends(optional_user),
    _: Subject = Depends(get_subject),
):
    data = ...
    return data
