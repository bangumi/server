from fastapi import APIRouter

from . import collection

router = APIRouter()
router.include_router(collection.router)
