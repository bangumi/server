from fastapi import APIRouter

from . import subject

router = APIRouter()
router.include_router(subject.router)
