from fastapi import APIRouter

from . import person, character

router = APIRouter()
router.include_router(person.router)
router.include_router(character.router)
