from fastapi import APIRouter

from . import person, subject, character

router = APIRouter()
router.include_router(person.router)
router.include_router(character.router)
router.include_router(subject.router)
