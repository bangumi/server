import datetime
from typing import List, Optional

from fastapi import Path, Depends, APIRouter
from starlette.responses import Response
from sqlalchemy.ext.asyncio import AsyncSession

from pol import sa, res, curd, wiki
from pol.res import ErrorDetail, not_found_exception
from pol.utils import subject_images
from pol.config import CACHE_KEY_PREFIX
from pol.router import ErrorCatchRoute
from pol.depends import get_db, get_redis
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
from pol.api.v0.utils import get_career, person_images, short_description
from pol.api.v0.models import RelatedPerson, RelatedCharacter
from pol.redis.json_cache import JSONRedis
from pol.http_cache.depends import CacheControl
from pol.api.v0.depends.auth import optional_user
from pol.api.v0.models.subject import Subject, RelatedSubject

router = APIRouter(tags=["条目"], route_class=ErrorCatchRoute)

api_base = "/v0/subjects"


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
    not_found: res.HTTPException = Depends(not_found_exception),
    subject_id: int = Path(..., gt=0),
    user: Role = Depends(optional_user),
    db: AsyncSession = Depends(get_db),
    redis: JSONRedis = Depends(get_redis),
    cache_control: CacheControl = Depends(CacheControl),
):
    """cache and permission wrapper"""
    cache_key = CACHE_KEY_PREFIX + f"subject:{subject_id}"
    s = await redis.get_with_model(cache_key, Subject)
    if s:
        response.headers["x-cache-status"] = "hit"
        nsfw = s.nsfw
        data = s.dict()
    else:
        # now fetch real data
        response.headers["x-cache-status"] = "miss"
        data = await _get_subject(
            cache_control, not_found=not_found, subject_id=subject_id, db=db
        )
        await redis.set_json(cache_key, value=data, ex=300)
        nsfw = data["nsfw"]

    if nsfw and not user.allow_nsfw():
        raise not_found

    return data


async def _get_subject(
    cache_control: CacheControl,
    not_found: res.HTTPException = Depends(not_found_exception),
    subject_id: int = Path(..., gt=0),
    db: AsyncSession = Depends(get_db),
):
    subject: Optional[ChiiSubject] = await db.get(
        ChiiSubject, subject_id, options=[sa.joinedload(ChiiSubject.fields)]
    )
    if subject is None:
        raise not_found

    if not subject.subject_nsfw:
        cache_control(300)

    if subject.fields.field_redirect:
        raise res.HTTPRedirect(f"{api_base}/{subject.fields.field_redirect}")

    if subject.ban:
        raise not_found

    date = None
    v = subject.fields.field_date
    if isinstance(v, datetime.date):
        date = f"{v.year:04d}-{v.month:02d}-{v.day:02d}"

    data = {
        "id": subject.subject_id,
        "name": subject.subject_name,
        "name_cn": subject.subject_name_cn,
        "date": date,
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
        "total_episodes": await curd.count(db, ChiiEpisode.ep_subject_id == subject_id),
        "tags": subject.fields.tags(),
    }

    try:
        data["infobox"] = wiki.parse(subject.field_infobox).info
    except wiki.WikiSyntaxError:
        data["infobox"] = None

    return data


@router.get(
    "/subjects/{subject_id}/persons",
    response_model=List[RelatedPerson],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_subject_persons(
    db: AsyncSession = Depends(get_db),
    not_found: res.HTTPException = Depends(not_found_exception),
    subject_id: int = Path(..., gt=0),
):
    subject: ChiiSubject = await db.scalar(
        sa.select(ChiiSubject)
        .options(
            sa.selectinload(ChiiSubject.persons).joinedload(ChiiPersonCsIndex.person)
        )
        .where(ChiiSubject.subject_id == subject_id, ChiiSubject.subject_ban == 0)
    )

    if not subject:
        raise not_found

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
    db: AsyncSession = Depends(get_db),
    not_found: res.HTTPException = Depends(not_found_exception),
    subject_id: int = Path(..., gt=0),
):
    subject: ChiiSubject = await db.scalar(
        sa.select(ChiiSubject)
        .options(
            sa.selectinload(ChiiSubject.characters).joinedload(
                ChiiCrtSubjectIndex.character
            )
        )
        .where(ChiiSubject.subject_id == subject_id, ChiiSubject.subject_ban == 0)
    )

    if not subject:
        raise not_found

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
    not_found: res.HTTPException = Depends(not_found_exception),
    subject_id: int = Path(..., gt=0),
    db: AsyncSession = Depends(get_db),
):
    subject: Optional[ChiiSubject] = await db.scalar(
        sa.select(ChiiSubject)
        .where(ChiiSubject.subject_id == subject_id, ChiiSubject.subject_ban == 0)
        .options(
            sa.selectinload(ChiiSubject.relations).selectinload(
                ChiiSubjectRelations.dst_subject
            )
        )
    )

    if not subject:
        raise not_found

    response = []

    for r in subject.relations:
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
