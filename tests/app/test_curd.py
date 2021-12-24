from sqlalchemy.orm import Session

from pol.curd import get_many
from tests.base import async_test
from pol.db.mysql import database
from pol.db.tables import ChiiSubject


@async_test
async def test_curd_get_many(db_session: Session):
    async with database:
        results = await get_many(database, ChiiSubject)
        assert [x.subject_id for x in results] == [
            x.subject_id for x in db_session.query(ChiiSubject).all()
        ]


@async_test
async def test_curd_get_many_args(db_session: Session):
    async with database:
        results = await get_many(
            database,
            ChiiSubject,
            limit=2,
            offset=3,
            order=ChiiSubject.subject_name.desc(),
        )
        expected = [
            x.subject_id
            for x in db_session.query(ChiiSubject)
            .order_by(ChiiSubject.subject_name.desc())
            .limit(2)
            .offset(3)
            .all()
        ]

        assert [x.subject_id for x in results] == expected


@async_test
async def test_curd_get_many_args(db_session: Session):
    async with database:
        results = await get_many(database, ChiiSubject, ChiiSubject.subject_id != 8)
        for s in results:
            assert s.subject_id != 8
