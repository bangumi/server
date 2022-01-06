"""
create a public archive at `--out`
"""
import os
import zipfile
from typing import IO
from argparse import ArgumentParser

import orjson
from loguru import logger
from sqlalchemy.orm import Session

from pol import sa
from pol.db.tables import (
    ChiiPerson,
    ChiiEpisode,
    ChiiSubject,
    ChiiCharacter,
    ChiiCrtCastIndex,
    ChiiPersonCsIndex,
    ChiiCrtSubjectIndex,
    ChiiSubjectRelations,
)
from pol.api.v0.utils import get_career


@logger.catch()
def main():
    parser = ArgumentParser()
    parser.add_argument("--out", default="./data/export.zip")

    args = parser.parse_args()
    logger.info("dump database to {}", os.path.abspath(args.out))

    SessionMaker = sa.sync_session_maker()
    with zipfile.ZipFile(
        args.out, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=7
    ) as zip_file:
        logger.info("dumping subjects")
        with zip_file.open("subject.jsonlines", "w") as f, SessionMaker() as session:
            export_subjects(f, session)

        logger.info("dumping subject relations")
        with zip_file.open(
            "subject-relations.jsonlines", "w"
        ) as f, SessionMaker() as session:
            export_subject_self_relation(f, session)

        logger.info("dumping characters")
        with zip_file.open("character.jsonlines", "w") as f, SessionMaker() as session:
            export_characters(f, session)

        logger.info("dumping persons")
        with zip_file.open("person.jsonlines", "w") as f, SessionMaker() as session:
            export_persons(f, session)

        logger.info("dumping subject persons")
        with zip_file.open(
            "subject-persons.jsonlines", "w"
        ) as f, SessionMaker() as session:
            export_subject_persons(f, session)

        logger.info("dumping subject characters")
        with zip_file.open(
            "subject-characters.jsonlines", "w"
        ) as f, SessionMaker() as session:
            export_subject_characters(f, session)

        logger.info("dumping person character")
        with zip_file.open(
            "person-characters.jsonlines", "w"
        ) as f, SessionMaker() as session:
            export_person_characters(f, session)

        logger.info("dumping episodes")
        with zip_file.open("episodes.jsonlines", "w") as f, SessionMaker() as session:
            export_episodes(f, session)


def export_person_characters(f: IO[bytes], session: Session):
    for cast in session.scalars(sa.select(ChiiCrtCastIndex)):
        cast: ChiiCrtCastIndex
        f.write(
            orjson.dumps(
                {
                    "person_id": cast.prsn_id,
                    "subject_id": cast.subject_id,
                    "character_id": cast.crt_id,
                    "summary": cast.summary,
                }
            )
        )
        f.write(b"\n")


def export_subject_persons(f: IO[bytes], session: Session):
    for person_subject in session.scalars(sa.select(ChiiPersonCsIndex)):
        person_subject: ChiiPersonCsIndex
        f.write(
            orjson.dumps(
                {
                    "person_id": person_subject.prsn_id,
                    "subject_id": person_subject.subject_id,
                    "position": person_subject.prsn_position,
                }
            )
        )
        f.write(b"\n")


def export_subject_characters(f: IO[bytes], session: Session):
    for crt_subject in session.scalars(sa.select(ChiiCrtSubjectIndex)):
        crt_subject: ChiiCrtSubjectIndex
        f.write(
            orjson.dumps(
                {
                    "character_id": crt_subject.crt_id,
                    "subject_id": crt_subject.subject_id,
                    "type": crt_subject.crt_type,
                    "order": crt_subject.crt_order,
                }
            )
        )
        f.write(b"\n")


def export_persons(f: IO[bytes], session: Session):
    for person in session.scalars(
        sa.select(ChiiPerson).where(ChiiPerson.prsn_ban == 0)
    ):
        person: ChiiPerson
        f.write(
            orjson.dumps(
                {
                    "id": person.prsn_id,
                    "name": person.prsn_name,
                    "type": person.prsn_type,
                    "career": get_career(person),
                    "infobox": person.prsn_infobox,
                    "summary": person.prsn_summary,
                }
            )
        )
        f.write(b"\n")


def export_characters(f: IO[bytes], session: Session):
    for character in session.scalars(
        sa.select(ChiiCharacter).where(ChiiCharacter.crt_ban == 0)
    ):
        character: ChiiCharacter
        f.write(
            orjson.dumps(
                {
                    "id": character.crt_id,
                    "role": character.crt_role,
                    "name": character.crt_name,
                    "infobox": character.crt_infobox,
                    "summary": character.crt_summary,
                }
            )
        )
        f.write(b"\n")


def export_subject_self_relation(f: IO[bytes], session: Session):
    for relation in session.scalars(sa.select(ChiiSubjectRelations)):
        relation: ChiiSubjectRelations
        f.write(
            orjson.dumps(
                {
                    "subject_id": relation.rlt_subject_id,
                    "relation_type": relation.rlt_relation_type,
                    "related_subject_id": relation.rlt_related_subject_id,
                    "order": relation.rlt_order,
                }
            )
        )
        f.write(b"\n")


def export_subjects(f: IO[bytes], session: Session):
    chunk = 50
    max_subject_id = session.scalar(
        sa.select(ChiiSubject.subject_id)
        .order_by(ChiiSubject.subject_id.desc())
        .limit(1)
    )

    for id in range(1, max_subject_id + chunk, chunk):
        for subject in session.scalars(
            sa.select(ChiiSubject).where(
                ChiiSubject.subject_id >= id,
                ChiiSubject.subject_id < id + chunk,
                ChiiSubject.subject_ban == 0,
            )
        ):
            subject: ChiiSubject
            f.write(
                orjson.dumps(
                    {
                        "id": subject.subject_id,
                        "name": subject.subject_name,
                        "name_cn": subject.subject_name_cn,
                        "infobox": subject.field_infobox,
                        "platform": subject.subject_platform,
                        "summary": subject.field_summary,
                        "nsfw": subject.subject_nsfw,
                    }
                )
            )
            f.write(b"\n")


def export_episodes(f: IO[bytes], session: Session):
    chunk = 50
    max_subject_id = session.scalar(
        sa.select(ChiiSubject.subject_id)
        .order_by(ChiiSubject.subject_id.desc())
        .limit(1)
    )

    for id in range(1, max_subject_id + chunk, chunk):
        for episode in session.scalars(
            sa.select(ChiiEpisode).where(
                ChiiEpisode.ep_subject_id >= id,
                ChiiEpisode.ep_subject_id < id + chunk,
                ChiiEpisode.ep_ban == 0,
            )
        ):
            episode: ChiiEpisode
            f.write(
                orjson.dumps(
                    {
                        "id": episode.ep_id,
                        "name": episode.ep_name,
                        "name_cn": episode.ep_name_cn,
                        "subject_id": episode.ep_subject_id,
                        "description": episode.ep_desc,
                        "type": episode.ep_type,
                        "airdate": episode.ep_airdate,
                        "disc": episode.ep_disc,
                    }
                )
            )
            f.write(b"\n")


if __name__ == "__main__":
    main()
