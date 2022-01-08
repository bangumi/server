from fastapi import APIRouter

from . import me, person, episode, subject, revision, character

router = APIRouter()
router.include_router(subject.router)
router.include_router(episode.router)
router.include_router(character.router)
router.include_router(person.router)
router.include_router(revision.router)
router.include_router(me.router)
