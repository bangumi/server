from typing import Optional

from fastapi import Path, Depends
from subject import router, exception_404
from sqlalchemy.ext.asyncio import AsyncSession
from mypy.typeshed.stdlib.typing import List

from pol import sa, res
from pol.models import ErrorDetail
from pol.depends import get_db
from pol.db.tables import ChiiSubject
from pol.permission.roles import Role
from pol.api.v0.models.base import ResponseCursorPaged
from pol.api.v0.depends.auth import optional_user
from pol.api.v0.models.topic import Topic


@router.get(
    "/subjects/{subject_id}/topics",
    response_model=List[Topic],
    responses={
        404: res.response(model=ErrorDetail),
    },
    summary="get a list of topics related to a subject; result is paginated",
    tags=["章节"],
)
async def get_subject_topics(
    response: ResponseCursorPaged[Topic, int],
    exc_404: res.HTTPException = Depends(exception_404),
    subject_id: int = Path(..., gt=0),
    role: Role = Depends(optional_user),
    db: AsyncSession = Depends(get_db),
):
    # todo: cache metadata of each topic (reply count, last reply timestamp)

    subject: Optional[ChiiSubject] = await db.get(
        ChiiSubject, subject_id, options=[sa.joinedload(ChiiSubject.fields)]
    )
    if subject is None:
        raise exc_404

    # check access permissions
    if subject.subject_nsfw and not role.allow_nsfw():
        raise exc_404

    # todo: load topic, check perm of each topic, then return the list
