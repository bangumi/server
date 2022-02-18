from typing import List
from asyncio import gather

from fastapi import Path, Depends, APIRouter
from starlette.responses import Response, RedirectResponse

from pol import res
from pol.res import ErrorDetail, not_found_exception
from pol.config import CACHE_KEY_PREFIX
from pol.router import ErrorCatchRoute
from pol.depends import get_redis
from pol.api.v0.utils import subject_images
from pol.api.v0.models import RelatedSubject, CharacterDetail, CharacterPerson
from pol.redis.json_cache import JSONRedis
from pol.services.person_service import PersonService
from pol.services.subject_service import SubjectService
from pol.services.character_service import CharacterService

router = APIRouter(tags=["角色"], route_class=ErrorCatchRoute)

api_base = "/v0/characters"


@router.get(
    "/characters/{character_id}",
    description="cache with 60s",
    response_model=CharacterDetail,
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_character_detail(
    response: Response,
    not_found: res.HTTPException = Depends(not_found_exception),
    character_service: CharacterService = Depends(CharacterService.new),
    character_id: int = Path(..., gt=0),
    redis: JSONRedis = Depends(get_redis),
):
    cache_key = CACHE_KEY_PREFIX + f"character:{character_id}"
    if value := await redis.get_with_model(cache_key, CharacterDetail):
        response.headers["x-cache-status"] = "hit"
        return value

    try:
        character = await character_service.get_by_id(character_id)
    except CharacterService.NotFoundError:
        raise not_found

    if character.redirect:
        return RedirectResponse(f"{api_base}/{character.redirect}")

    data = character.dict()
    response.headers["x-cache-status"] = "miss"
    await redis.set_json(cache_key, value=data, ex=60)

    return data


@router.get(
    "/characters/{character_id}/subjects",
    summary="get character related subjects",
    response_model=List[RelatedSubject],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_person_subjects(
    not_found: res.HTTPException = Depends(not_found_exception),
    character_service: CharacterService = Depends(CharacterService.new),
    subject_service: SubjectService = Depends(SubjectService.new),
    character_id: int = Path(..., gt=0),
):
    try:
        character_subjects = await character_service.list_subjects_by_id(character_id)
    except CharacterService.NotFoundError:
        raise not_found

    subjects = await subject_service.get_by_ids(
        *[s.id for s in character_subjects], include_nsfw=True
    )

    return [
        {
            "id": s.id,
            "name": ss.name,
            "name_cn": ss.name_cn,
            "staff": s.staff,
            "image": images["grid"] if images else None,
        }
        for s, ss, images in (
            (s, subjects[s.id], subject_images(subjects[s.id].image))
            for s in character_subjects
            if s.id in subjects
        )
    ]


@router.get(
    "/characters/{character_id}/persons",
    summary="get character related persons",
    response_model=List[CharacterPerson],
    responses={
        404: res.response(model=ErrorDetail),
    },
)
async def get_character_persons(
    character_service: CharacterService = Depends(CharacterService.new),
    person_serivce: PersonService = Depends(PersonService.new),
    subject_service: SubjectService = Depends(SubjectService.new),
    character_id: int = Path(..., gt=0),
):
    await character_service.get_by_id(character_id)

    person_map = await character_service.get_persons_by_ids(
        id for id in (character_id,)
    )

    person_to_subject = [
        (person.id, person.subject_id)
        for person_list in person_map.values()
        for person in person_list
    ]

    subjects, persons = await gather(
        subject_service.get_by_ids(
            *(id for _, id in person_to_subject), include_nsfw=True
        ),
        person_serivce.get_by_ids(id for id, _ in person_to_subject),
    )

    persons = [
        {
            "id": r.id,
            "name": r.name,
            "type": r.type,
            "images": r.images,
            "subject_id": subject.id,
            "subject_name": subject.name,
            "subject_name_cn": subject.name_cn,
        }
        for r, subject in (
            (persons[person_id], subjects[subject_id])
            for person_id, subject_id in person_to_subject
            if person_id in persons and subject_id in subjects
        )
    ]

    return persons
