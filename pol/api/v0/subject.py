from fastapi import Path, Depends, APIRouter
from databases import Database
from starlette.responses import Response, RedirectResponse

from pol import res, curd
from pol.models import ErrorDetail
from pol.depends import get_db
from pol.db.tables import ChiiSubject
from pol.api.v0.const import NotFoundDescription
from pol.curd.exceptions import NotFoundError

router = APIRouter(tags=["条目"])

api_base = "/v0/subjects"


async def basic_subject(db: Database, subject_id: int) -> curd.subject.Subject:
    try:
        return await curd.subject.get_one(db, ChiiSubject.subject_id == subject_id)
    except NotFoundError:
        raise res.HTTPException(
            status_code=404,
            title="Not Found",
            description=NotFoundDescription,
            detail={"character_id": "character_id"},
        )


@router.get(
    "/subjects/{subject_id}",
    description="cache with 300s",
    response_model_by_alias=False,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject(
    response: Response,
    db: Database = Depends(get_db),
    subject_id: int = Path(..., gt=0),
    # redis: JSONRedis = Depends(get_redis),
):
    # cache_key = CACHE_KEY_PREFIX + f"subject:{subject_id}"
    # if (value := await redis.get(cache_key)) is not None:
    #     response.headers["x-cache-status"] = "hit"
    #     return value
    # else:
    #     response.headers["x-cache-status"] = "miss"

    subject = await basic_subject(db, subject_id)

    if subject.redirect:
        return RedirectResponse(f"{api_base}/{subject.redirect}")

    data = {
        "id": subject.id,
        "name": subject.name,
        "date": subject.date,
    }

    # await redis.set_json(cache_key, value=data, ex=300)

    return data
