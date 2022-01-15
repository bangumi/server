from fastapi import APIRouter

from . import me, user, topic, person, episode, subject, revision, character

router = APIRouter()
router.include_router(subject.router)
router.include_router(episode.router)
router.include_router(character.router)
router.include_router(person.router)
router.include_router(me.router)
router.include_router(user.router)
router.include_router(topic.router)
router.include_router(revision.router)
