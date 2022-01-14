from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from pol.depends import get_db
from pol.services.rev_service.ep_rev import _EpisodeRevsionService
from pol.services.rev_service.exception import RevisionNotFound
from pol.services.rev_service.person_rev import _PersonRevisionService
from pol.services.rev_service.subject_rev import _SubjectRevisionService
from pol.services.rev_service.character_rev import _CharacterRevisionService


class RevisionService(
    _EpisodeRevsionService,
    _SubjectRevisionService,
    _PersonRevisionService,
    _CharacterRevisionService,
):
    __slots__ = ("_db",)
    _db: AsyncSession
    NotFoundError = RevisionNotFound

    @classmethod
    async def new(cls, session: AsyncSession = Depends(get_db)):
        return cls(session)

    def __init__(self, db: AsyncSession):
        self._db = db
