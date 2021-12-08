from fastapi import APIRouter

from . import v0

router = APIRouter()
router.include_router(v0.router, prefix="/v0")
