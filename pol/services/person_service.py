from typing import Iterator, cast

from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from pol.db import sa
from pol.depends import get_db
from pol.db.tables import ChiiPerson
from pol.api.v0.utils import get_career, person_images, short_description
from pol.api.v0.models import Person
from pol.curd.exceptions import NotFoundError


class PersonNotFound(NotFoundError):
    pass


class PersonService:

    __slots__ = ("_db",)
    _db: AsyncSession
    NotFoundError = PersonNotFound

    @classmethod
    def new(cls, session: AsyncSession = Depends(get_db)):
        return cls(session)

    def __init__(self, db: AsyncSession):
        self._db = db

    async def get_by_ids(self, ids: Iterator[int]):
        persons = cast(
            Iterator[ChiiPerson],
            await self._db.scalars(
                sa.select(ChiiPerson).where(
                    ChiiPerson.prsn_id.in_(ids), ChiiPerson.prsn_ban == 0
                )
            ),
        )

        return {
            cast(int, person.prsn_id): Person(
                id=person.prsn_id,
                name=person.prsn_name,
                type=person.prsn_type,
                career=get_career(person),
                short_summary=short_description(person.prsn_summary),
                images=person_images(person.prsn_img),
                locked=person.prsn_lock,
            )
            for person in persons
        }
