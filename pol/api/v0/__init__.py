from fastapi import APIRouter

from . import person

router = APIRouter()
router.include_router(person.router)
