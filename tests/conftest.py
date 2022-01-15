from abc import abstractmethod
from typing import Protocol, Generator
from datetime import datetime
from unittest import mock
from contextlib import asynccontextmanager
from collections import defaultdict

import redis
import pytest
from _pytest.nodes import Node
from sqlalchemy.orm import Session, sessionmaker
from starlette.testclient import TestClient
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine

import pol.server
from pol import config
from pol.db import sa
from tests.base import async_lambda
from pol.depends import get_db, get_redis
from pol.db.const import Gender, BloodType, PersonType, SubjectType, CollectionType
from pol.db.tables import (
    ChiiMember,
    ChiiPerson,
    ChiiSubject,
    ChiiPersonField,
    ChiiSubjectField,
    ChiiSubjectInterest,
    ChiiOauthAccessToken,
)


def pytest_addoption(parser):
    parser.addoption("--e2e", action="store_true", help="Run E2E tests")
    parser.addoption(
        "--database", action="store_true", help="Enable tests require database"
    )
    parser.addoption("--redis", action="store_true", help="Enable tests require redis")


def pytest_configure(config):
    # register an additional marker
    config.addinivalue_line(
        "markers",
        "env(name1, ...name2): mark test to run only on named environment",
    )


def pytest_runtest_setup(item: Node):
    mark = item.get_closest_marker(name="env")
    if not mark:
        return
    for name in mark.args:
        if not item.config.getoption(name):
            pytest.skip(f"test skipped without flag --{name}")


DBSession = sa.sync_session_maker()


@pytest.fixture()
def app():
    return pol.server.app


class MockAsyncSession(Protocol):
    get: mock.AsyncMock
    scalar: mock.AsyncMock
    scalars: mock.AsyncMock


@pytest.fixture()
def mock_db(app) -> MockAsyncSession:
    """mock mock AsyncSession, also override dependency `get_db` for all router"""
    db = mock.Mock()
    db.get = mock.AsyncMock(return_value=None)
    db.scalar = mock.AsyncMock(return_value=None)
    db.scalars = mock.AsyncMock(return_value=None)
    app.dependency_overrides[get_db] = async_lambda(db)
    return db


class MockRedis(Protocol):
    get: mock.AsyncMock
    set_json: mock.AsyncMock
    get_with_model: mock.AsyncMock


@pytest.fixture()
def mock_redis(app) -> MockRedis:
    """mock redis client, also override dependency `get_redis` for all router"""
    r = mock.Mock()
    r.set_json = mock.AsyncMock()
    r.get = mock.AsyncMock(return_value=None)
    r.get_with_model = mock.AsyncMock(return_value=None)
    app.dependency_overrides[get_redis] = async_lambda(r)
    return r


@pytest.fixture()
def db_session():
    db_session = DBSession()
    try:
        yield db_session
    except Exception:  # pragma: no cover
        db_session.rollback()
        raise
    finally:
        db_session.close()


@pytest.fixture()
def AsyncSessionMaker():
    @asynccontextmanager
    async def get():
        engine = create_async_engine(
            "mysql+aiomysql://{}:{}@{}:{}/{}".format(
                config.MYSQL_USER,
                config.MYSQL_PASS,
                config.MYSQL_HOST,
                config.MYSQL_PORT,
                config.MYSQL_DB,
            ),
            pool_recycle=14400,
            pool_size=10,
            max_overflow=20,
        )

        SS = sessionmaker(engine, expire_on_commit=False, class_=AsyncSession)

        async with SS() as s:
            yield s

        # ensure to dispose the engine after usage.
        # otherwise asyncio will raise a RuntimeError
        await engine.dispose()

    return get


@pytest.fixture()
def redis_client():
    with redis.Redis.from_url(config.REDIS_URI) as redis_client:
        redis_client.flushdb()
        try:
            yield redis_client
        finally:
            redis_client.flushdb()


access_token = "a_development_access_token"


@pytest.fixture()
def client(app):
    with TestClient(app) as test_client:
        yield test_client
    app.dependency_overrides = {}


@pytest.fixture()
def auth_header():
    return {"Authorization": f"Bearer {access_token}"}


@pytest.fixture()
def mock_subject(db_session: Session):
    mock_subject_id = set()
    delete_query = defaultdict(list)

    def mock_id(
        subject_id: int,
        subject_name="",
        subject_type=SubjectType.anime,
        subject_name_cn="",
        subject_uid=1,
        subject_creator=1,
        subject_image="",
        field_infobox="",
        field_summary="",
        field_5="",
        subject_idx_cn="",
        subject_airtime=1,
        subject_nsfw=0,
        subject_ban=0,
        field_tags="",
        field_airtime=1,
        field_year=2007,
        field_mon=1,
        field_week_day=3,
        field_date=datetime.now().astimezone().date(),
    ):
        delete_query[ChiiSubject].append(ChiiSubject.subject_id == subject_id)
        delete_query[ChiiSubjectField].append(ChiiSubjectField.field_sid == subject_id)

        check_exist(db_session, delete_query)
        mock_subject_id.add(subject_id)

        db_session.add(
            ChiiSubject(
                subject_id=subject_id,
                subject_name=subject_name,
                subject_name_cn=subject_name_cn,
                subject_uid=subject_uid,
                subject_type_id=subject_type,
                subject_creator=subject_creator,
                subject_image=subject_image,
                field_infobox=field_infobox,
                field_summary=field_summary,
                field_5=field_5,
                subject_idx_cn=subject_idx_cn,
                subject_airtime=subject_airtime,
                subject_nsfw=subject_nsfw,
                subject_ban=subject_ban,
            )
        )
        db_session.add(
            ChiiSubjectField(
                field_sid=subject_id,
                field_tags=field_tags,
                field_airtime=field_airtime,
                field_year=field_year,
                field_mon=field_mon,
                field_week_day=field_week_day,
                field_date=field_date,
            )
        )
        db_session.commit()

    try:
        yield mock_id
    finally:
        for table, where in delete_query.items():
            db_session.execute(sa.delete(table).where(sa.or_(*where)))
        db_session.commit()


class MockUser(Protocol):
    @abstractmethod
    def __call__(
        self,
        user_id: int,
        access_token: str = "",
        username: str = "",
        expires: datetime = datetime.now(),
    ) -> None:
        pass


@pytest.fixture()
def mock_person(db_session: Session):
    mock_person_id = set()
    delete_query = defaultdict(list)

    def mock_id(
        prsn_id: int,
        prsn_name="",
        prsn_type=PersonType.person,
        prsn_infobox="",
        prsn_producer=0,
        prsn_mangaka=0,
        prsn_artist=1,
        prsn_seiyu=0,
        prsn_writer=0,
        prsn_illustrator=0,
        prsn_actor=0,
        prsn_summary="",
        prsn_img="",
        prsn_img_anidb="",
        prsn_comment=0,
        prsn_collects=0,
        prsn_dateline=0,
        prsn_lastpost=0,
        prsn_lock=0,
        prsn_anidb_id=0,
        prsn_ban=0,
        prsn_redirect=0,
        prsn_nsfw=0,
        gender=Gender.male,
        bloodtype=BloodType.o,
        birth_year=2000,
        birth_mon=1,
        birth_day=1,
    ):
        delete_query[ChiiPerson].append(ChiiPerson.prsn_id == prsn_id)
        delete_query[ChiiPersonField].append(ChiiPersonField.prsn_id == prsn_id)

        check_exist(db_session, delete_query)
        mock_person_id.add(prsn_id)

        db_session.add(
            ChiiPerson(
                prsn_id=prsn_id,
                prsn_name=prsn_name,
                prsn_type=prsn_type,
                prsn_infobox=prsn_infobox,
                prsn_producer=prsn_producer,
                prsn_mangaka=prsn_mangaka,
                prsn_artist=prsn_artist,
                prsn_seiyu=prsn_seiyu,
                prsn_writer=prsn_writer,
                prsn_illustrator=prsn_illustrator,
                prsn_actor=prsn_actor,
                prsn_summary=prsn_summary,
                prsn_img=prsn_img,
                prsn_img_anidb=prsn_img_anidb,
                prsn_comment=prsn_comment,
                prsn_collects=prsn_collects,
                prsn_dateline=prsn_dateline,
                prsn_lastpost=prsn_lastpost,
                prsn_lock=prsn_lock,
                prsn_anidb_id=prsn_anidb_id,
                prsn_ban=prsn_ban,
                prsn_redirect=prsn_redirect,
                prsn_nsfw=prsn_nsfw,
            )
        )
        db_session.add(
            ChiiPersonField(
                prsn_cat="prsn",
                prsn_id=prsn_id,
                gender=gender,
                bloodtype=bloodtype,
                birth_year=birth_year,
                birth_mon=birth_mon,
                birth_day=birth_day,
            )
        )
        db_session.commit()

    try:
        yield mock_id
    finally:
        for table, where in delete_query.items():
            db_session.execute(sa.delete(table).where(sa.or_(*where)))
        db_session.commit()


@pytest.fixture()
def mock_user(db_session: Session) -> Generator[MockUser, None, None]:
    mock_user_id = set()
    mock_token = set()
    delete_query = defaultdict(list)

    def mock_id(
        user_id: int,
        access_token: str = "",
        username: str = "",
        expires=datetime.now(),
    ):
        if user_id in mock_user_id and (not access_token or access_token in mock_token):
            return
        mock_user_id.add(user_id)
        mock_token.add(access_token)
        if access_token:
            delete_query[ChiiOauthAccessToken].append(
                ChiiOauthAccessToken.access_token == access_token
            )
        delete_query[ChiiMember].append(ChiiMember.uid == user_id)
        try:
            check_exist(db_session, delete_query)
        except ValueError as e:
            print(e)
            return

        if access_token:
            db_session.add(
                ChiiOauthAccessToken(
                    access_token=access_token,
                    client_id="",
                    user_id=user_id,
                    expires=expires,
                    scope=None,
                )
            )

        db_session.add(
            ChiiMember(
                uid=user_id,
                username=username or f"mock_{user_id}",
                nickname="",
                avatar="",
                groupid=10,
                sign="",
            )
        )
        db_session.commit()

    try:
        yield mock_id
    finally:
        for table, where in delete_query.items():
            db_session.execute(sa.delete(table).where(sa.or_(*where)))
        db_session.commit()


@pytest.fixture()
def mock_user_collection(db_session: Session):
    mock_ids = set()
    delete_query = defaultdict(list)

    def mock_collection(
        id: int,
        user_id: int,
        subject_id: int,
        subject_type=SubjectType.anime,
        type=CollectionType.doing,
        private=False,
    ):
        delete_query[ChiiSubjectInterest].append(ChiiSubjectInterest.id == id)
        check_exist(db_session, delete_query)

        mock_ids.add(id)

        db_session.add(
            ChiiSubjectInterest(
                id=id,
                user_id=user_id,
                subject_id=subject_id,
                subject_type=subject_type,
                type=type,
                private=private,
            )
        )

        db_session.commit()

    try:
        yield mock_collection
    finally:
        for table, where in delete_query.items():
            db_session.execute(sa.delete(table).where(sa.or_(*where)))
        db_session.commit()


def check_exist(db_session: Session, quries):
    for Table, q in quries.items():
        query = q[-1]
        r = db_session.execute(sa.select(Table).where(query)).all()
        if r:  # pragma: no cover
            raise ValueError(
                f"record {query.left.name} == {repr(query.right.value)}"
                f" exists in table {Table}, can't mock it"
            )

    for Table, q in quries.items():
        db_session.execute(sa.delete(Table).where(q[-1]))
