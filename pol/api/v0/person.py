from fastapi import Depends, APIRouter, HTTPException
from databases import Database
from starlette.responses import RedirectResponse

from pol import res, curd, wiki
from pol.utils import imgUrl
from pol.api.v0 import models
from pol.models import ErrorDetail
from pol.depends import get_db
from pol.db.tables import ChiiPerson
from pol.curd.exceptions import NotFoundError

router = APIRouter()


@router.get(
    "/person/{person_id}",
    response_model=models.Person,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def bgm_ip_map(
    person_id: int,
    db: Database = Depends(get_db),
):
    try:
        person = await curd.get_one(db, ChiiPerson, ChiiPerson.prsn_id == person_id)
    except NotFoundError:
        raise HTTPException(status_code=404, detail="person not found")

    if person.prsn_redirect:
        return RedirectResponse(str(person.prsn_redirect))

    m = models.Person(
        id=person.prsn_id,
        name=person.prsn_name,
        type=person.prsn_type,
        infobox=person.prsn_infobox,
        role=person.role(),
        summary=person.prsn_summary,
        img=imgUrl(person.prsn_img),
    )

    try:
        m.wiki = wiki.parse(person.prsn_infobox).info
    except wiki.WikiSyntaxError:
        pass

    return m
