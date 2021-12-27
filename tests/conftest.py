from datetime import datetime
from collections import defaultdict

import redis
import pytest
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from pol import sa, config
from pol.db.const import SubjectType
from pol.db.tables import (
    ChiiMember,
    ChiiSubject,
    ChiiSubjectField,
    ChiiOauthAccessToken,
)

engine = create_engine(
    "mysql+pymysql://{}:{}@{}:{}/{}".format(
        config.MYSQL_USER,
        config.MYSQL_PASS,
        config.MYSQL_HOST,
        config.MYSQL_PORT,
        config.MYSQL_DB,
    )
)

Session = sessionmaker(bind=engine)


@pytest.fixture()
def db_session():
    db_session = Session()
    try:
        yield db_session
    except Exception:  # pragma: no cover
        db_session.rollback()
        raise
    finally:
        db_session.close()


@pytest.fixture(autouse=True)
def redis_client():
    with redis.Redis.from_url(config.REDIS_URI) as redis_client:
        redis_client.flushdb()
        try:
            yield redis_client
        finally:
            redis_client.flushdb()


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


@pytest.fixture()
def mock_access_token(db_session: Session):
    mock_user_id = set()
    mock_token = set()
    delete_query = defaultdict(list)

    def mock_id(
        user_id: int,
        access_token: str,
        expires=datetime.now(),
    ):
        delete_query[ChiiOauthAccessToken].append(
            ChiiOauthAccessToken.access_token == access_token
        )
        delete_query[ChiiMember].append(ChiiMember.uid == user_id)
        check_exist(db_session, delete_query)

        mock_user_id.add(user_id)
        mock_token.add(access_token)

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
            ChiiMember(uid=user_id, nickname="", avatar="", groupid=10, sign="")
        )
        db_session.commit()

    try:
        yield mock_id
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
