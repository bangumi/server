import pydantic
from fastapi import Depends, APIRouter

from pol import res
from pol.models import ErrorDetail
from .depends.auth import User, get_current_user

router = APIRouter()


class Me(pydantic.BaseModel):
    pass


@router.get(
    "/me",
    description="cache with 300s",
    response_model_by_alias=False,
    response_model=User,
    responses={
        403: res.response(model=ErrorDetail, description="unauthorized"),
    },
)
async def get_subject(
    user: User = Depends(get_current_user),
):
    return user.dict(by_alias=True)
