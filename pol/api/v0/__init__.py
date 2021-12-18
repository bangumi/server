from fastapi import APIRouter

from . import person, episode, subject, character

router = APIRouter()
router.include_router(subject.router)
router.include_router(episode.router)
router.include_router(character.router)
router.include_router(person.router)
