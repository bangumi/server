from typing import List, Optional

from fastapi import Path, Depends, Request, APIRouter
from starlette.responses import Response, RedirectResponse
from sqlalchemy.ext.asyncio import AsyncSession

from pol import sa, res, curd, wiki
from pol.utils import subject_images
from pol.config import CACHE_KEY_PREFIX
from pol.models import ErrorDetail
from pol.depends import get_redis, get_session
from pol.db.const import (
    PLATFORM_MAP,
    RELATION_MAP,
    StaffMap,
    SubjectType,
    get_character_rel,
)
from pol.db.tables import (
    ChiiEpisode,
    ChiiSubject,
    ChiiPersonCsIndex,
    ChiiCrtSubjectIndex,
    ChiiSubjectRelations,
)
from pol.permission import Role
from pol.api.v0.const import NotFoundDescription
from pol.api.v0.utils import get_career, person_images, short_description
from pol.api.v0.models import RelatedPerson, RelatedCharacter
from pol.redis.json_cache import JSONRedis
from pol.api.v0.depends.auth import optional_user
from pol.api.v0.models.subject import Subject, RelatedSubject

router = APIRouter(tags=["条目"])

api_base = "/v0/subjects"


async def exception_404(request: Request):
    detail = dict(request.query_params)
    detail.update(request.path_params)
    return res.HTTPException(
        status_code=404,
        title="Not Found",
        description=NotFoundDescription,
        detail=detail,
    )


@router.get(
    "/subjects/{subject_id}",
    description="cache with 300s",
    response_model=Subject,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject(
    response: Response,
    exc_404: res.HTTPException = Depends(exception_404),
    subject_id: int = Path(..., gt=0),
    user: Role = Depends(optional_user),
    db_session: AsyncSession = Depends(get_session),
    redis: JSONRedis = Depends(get_redis),
):
    cache_key = CACHE_KEY_PREFIX + f"subject:{subject_id}"
    if s := await redis.get_with_model(cache_key, Subject):
        if s.nsfw and not user.allow_nsfw():
            raise exc_404
        response.headers["x-cache-status"] = "hit"
        return s
    else:
        response.headers["x-cache-status"] = "miss"

    subject: Optional[ChiiSubject] = await db_session.get(
        ChiiSubject, subject_id, options=[sa.joinedload(ChiiSubject.fields)]
    )
    if subject is None:
        raise exc_404

    if subject.fields.field_redirect:
        return RedirectResponse(f"{api_base}/{subject.fields.field_redirect}")

    if subject.ban:
        raise exc_404

    data = {
        "id": subject.subject_id,
        "name": subject.subject_name,
        "name_cn": subject.subject_name_cn,
        "type": subject.subject_type_id,
        "summary": subject.field_summary,
        "eps": subject.field_eps,
        "volumes": subject.field_volumes,
        "locked": subject.locked,
        "images": subject_images(subject.subject_image),
        "nsfw": subject.subject_nsfw,
        "collection": {
            "wish": subject.subject_wish,
            "collect": subject.subject_collect,
            "doing": subject.subject_doing,
            "on_hold": subject.subject_on_hold,
            "dropped": subject.subject_dropped,
        },
        "rating": subject.fields.rating(),
        "platform": PLATFORM_MAP[subject.subject_type_id].get(
            subject.subject_platform, {"type_cn": ""}
        )["type_cn"],
        "total_episodes": await curd.count(
            db_session, ChiiEpisode.ep_subject_id == subject_id
        ),
        "tags": subject.fields.tags(),
    }

    try:
        data["infobox"] = wiki.parse(subject.field_infobox).info
    except wiki.WikiSyntaxError:
        data["infobox"] = None

    await redis.set_json(cache_key, value=data, ex=300)

    if subject.subject_nsfw and not user.allow_nsfw():
        raise exc_404

    return data


@router.get(
    "/subjects/{subject_id}/persons",
    response_model=List[RelatedPerson],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_persons(
    db_session: AsyncSession = Depends(get_session),
    exc_404: res.HTTPException = Depends(exception_404),
    subject_id: int = Path(..., gt=0),
):
    subject: ChiiSubject = await db_session.scalar(
        sa.select(ChiiSubject)
        .options(
            sa.selectinload(ChiiSubject.persons).joinedload(ChiiPersonCsIndex.person)
        )
        .where(ChiiSubject.subject_id == subject_id, ChiiSubject.subject_ban == 0)
    )

    if not subject:
        raise exc_404

    persons = []

    for rel in subject.persons:
        p = rel.person
        persons.append(
            {
                "id": p.prsn_id,
                "name": p.prsn_name,
                "type": p.prsn_type,
                "relation": StaffMap[rel.subject_type_id][rel.prsn_position].str(),
                "career": get_career(p),
                "short_summary": short_description(p.prsn_summary),
                "images": person_images(p.prsn_img),
            }
        )
    return persons


@router.get(
    "/subjects/{subject_id}/characters",
    response_model=List[RelatedCharacter],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_characters(
    db_session: AsyncSession = Depends(get_session),
    exc_404: res.HTTPException = Depends(exception_404),
    subject_id: int = Path(..., gt=0),
):
    subject: ChiiSubject = await db_session.scalar(
        sa.select(ChiiSubject)
        .options(
            sa.selectinload(ChiiSubject.characters).joinedload(
                ChiiCrtSubjectIndex.character
            )
        )
        .where(ChiiSubject.subject_id == subject_id, ChiiSubject.subject_ban == 0)
    )

    if not subject:
        raise exc_404

    characters = []
    for rel in subject.characters:
        r = rel.character
        characters.append(
            {
                "id": r.crt_id,
                "name": r.crt_name,
                "relation": get_character_rel(rel.crt_type),
                "type": r.crt_role,
                "short_summary": short_description(r.crt_summary),
                "images": person_images(r.crt_img),
            }
        )
    return characters


@router.get(
    "/subjects/{subject_id}/subjects",
    response_model=List[RelatedSubject],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_relations(
    exc_404: res.HTTPException = Depends(exception_404),
    subject_id: int = Path(..., gt=0),
    db_session: AsyncSession = Depends(get_session),
):
    if not await db_session.scalar(
        sa.select(ChiiSubject.subject_id).where(
            ChiiSubject.subject_id == subject_id, ChiiSubject.subject_ban == 0
        )
    ):
        raise exc_404

    relations: List[ChiiSubjectRelations] = await db_session.scalars(
        sa.select(ChiiSubjectRelations)
        .options(sa.selectinload(ChiiSubjectRelations.dst_subject))
        .where(
            ChiiSubjectRelations.rlt_subject_id == subject_id,
        )
        .order_by(
            ChiiSubjectRelations.rlt_order, ChiiSubjectRelations.rlt_related_subject_id
        )
    )

    if not relations:
        raise exc_404

    response = []

    for r in relations:
        s = r.dst_subject
        relation = RELATION_MAP[r.rlt_related_subject_type_id].get(r.rlt_relation_type)

        if relation is None or r.rlt_related_subject_type_id == 1:
            rel = SubjectType(r.rlt_related_subject_type_id).str()
        else:
            rel = relation.str()

        response.append(
            {
                "id": r.rlt_related_subject_id,
                "relation": rel,
                "name": s.subject_name,
                "type": r.rlt_related_subject_type_id,
                "name_cn": s.subject_name_cn,
                "images": subject_images(s.subject_image),
            }
        )

    return response
