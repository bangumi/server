from fastapi import Path, Depends
from subject import router, exception_404
from fastapi.exceptions import RequestValidationError
from sqlalchemy.ext.asyncio import AsyncSession
from mypy.typeshed.stdlib.typing import List

from pol import res
from pol.models import ErrorDetail
from pol.depends import get_db
from pol.curd.user import User
from pol.db.tables import ChiiSubject
from pol.api.v0.models.base import OffsetPagedResponse
from pol.api.v0.depends.auth import optional_user
from pol.api.v0.models.topic import Topic
from pol.api.v0.depends.page.offset import OffsetPage

MAX_PAGE = 1000
MAX_PAGE_SIZE = 100


def get_subject_topics_page(*args) -> OffsetPage:
    page = OffsetPage.from_request(*args)

    if page.page > MAX_PAGE:
        raise RequestValidationError()


@router.get(
    "/subjects/{subject_id}/topics",
    response_model=List[Topic],
    responses={
        400: res.response(model=ErrorDetail),
        404: res.response(model=ErrorDetail),
    },
    summary="get a list of topics related to a subject; result is paginated",
    tags=["章节"],
)
async def get_subject_topics(
    response: OffsetPagedResponse[Topic],
    exc_404: res.HTTPException = Depends(exception_404),
    subject_id: int = Path(..., gt=0),
    user: User = Depends(optional_user),
    db: AsyncSession = Depends(get_db),
    page_params: OffsetPage = Depends(get_subject_topics_page),
):
    # todo: cache metadata of each topic (reply count, last reply timestamp)

    # sanity check

    subject = (
        await db.query(ChiiSubject).filter(ChiiSubject.subject_id == subject_id).one()
    )

    if subject is None:
        raise exc_404

    # check access permissions
    if subject.subject_nsfw and not user.to_role().allow_nsfw():
        raise exc_404

    # fetch topics
    # order_by, page_cond = page_to_sort(page_params)

    # ChiiSubjectTopic.sbj_tpc_id

    # TopicPaginationKey[page_params.key]
    #
    # topic = db.query(ChiiSubjectTopic).filter(
    #     and_(ChiiSubjectTopic.sbj_tpc_id > page_params.cursor)
    # )

    # todo: load topic, check perm of each topic, then return the list
