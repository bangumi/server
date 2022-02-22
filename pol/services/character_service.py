from typing import Dict, List, Iterator, Optional, cast

from fastapi import Depends
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession

from pol import wiki
from pol.db import sa
from pol.depends import get_db
from pol.db.const import Gender, get_character_rel
from pol.db.tables import ChiiCharacter, ChiiCrtCastIndex, ChiiCrtSubjectIndex
from pol.wiki.parser import WikiSyntaxError
from pol.api.v0.utils import person_images, short_description
from pol.api.v0.models import CharacterDetail
from pol.curd.exceptions import NotFoundError


class CharacterNotFound(NotFoundError):
    pass


class Character(CharacterDetail):
    redirect: int


class CharacterPerson(BaseModel):
    id: int
    subject_id: int
    summary: str


class CharacterSubject(BaseModel):
    id: int
    staff: str


class CharacterService:

    __slots__ = ("_db",)
    _db: AsyncSession
    NotFoundError = CharacterNotFound

    @classmethod
    def new(cls, session: AsyncSession = Depends(get_db)):
        return cls(session)

    def __init__(self, db: AsyncSession):
        self._db = db

    async def get_by_id(self, id: int):
        character: Optional[ChiiCharacter] = await self._db.scalar(
            sa.select(ChiiCharacter)
            .options(sa.joinedload(ChiiCharacter.field))
            .where(ChiiCharacter.crt_id == id)
        )
        if not character or (not character.crt_redirect and character.crt_ban):
            raise self.NotFoundError
        return self._convert_to_orm(character)

    async def get_by_ids(self, ids: Iterator[int]):
        characters = cast(
            Iterator[ChiiCharacter],
            await self._db.scalars(
                sa.select(ChiiCharacter)
                .options(sa.joinedload(ChiiCharacter.field))
                .where(ChiiCharacter.crt_id.in_(ids), ChiiCharacter.crt_ban == 0)
            ),
        )
        return {cast(int, c.crt_id): self._convert_to_orm(c) for c in characters}

    def _convert_to_orm(self, character: ChiiCharacter):
        res = {
            "id": character.crt_id,
            "name": character.crt_name,
            "type": character.crt_role,
            "redirect": character.crt_redirect,
            "images": person_images(character.crt_img),
            "summary": short_description(character.crt_summary),
            "locked": character.crt_lock,
            "stat": {
                "comments": character.crt_comment,
                "collects": character.crt_collects,
            },
        }

        if not character.crt_redirect:
            field = character.field
            if field is not None:
                if field.gender:
                    res["gender"] = Gender(field.gender).str()
                res["blood_type"] = field.bloodtype or None
                res["birth_year"] = field.birth_year or None
                res["birth_mon"] = field.birth_mon or None
                res["birth_day"] = field.birth_day or None
            try:
                res["infobox"] = wiki.parse(character.crt_infobox).info
            except WikiSyntaxError:
                pass
        return Character(**res)

    async def get_persons_by_ids(self, ids: Iterator[int], subject_id: int = 0):
        where = [ChiiCrtCastIndex.crt_id.in_(ids)]
        if subject_id:
            where.append(ChiiCrtCastIndex.subject_id == subject_id)
        res: Dict[int, List[CharacterPerson]] = {}
        for person in cast(
            Iterator[ChiiCrtCastIndex],
            await self._db.scalars(sa.select(ChiiCrtCastIndex).where(*where)),
        ):
            id = cast(int, person.crt_id)
            if id not in res:
                res[id] = []
            res[id].append(
                CharacterPerson(
                    id=person.prsn_id,
                    subject_id=person.subject_id,
                    summary=person.summary,
                )
            )
        return res

    async def list_subjects_by_id(self, id: int):
        character: Optional[ChiiCharacter] = await self._db.scalar(
            sa.select(ChiiCharacter).where(
                ChiiCharacter.crt_id == id, ChiiCharacter.crt_ban == 0
            )
        )
        if not character:
            raise self.NotFoundError
        res = cast(
            Iterator[ChiiCrtSubjectIndex],
            await self._db.scalars(
                sa.select(ChiiCrtSubjectIndex).where(ChiiCrtSubjectIndex.crt_id == id)
            ),
        )

        return [
            CharacterSubject(id=i.subject_id, staff=get_character_rel(i.crt_type))
            for i in res
        ]
