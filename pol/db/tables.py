from typing import List

from sqlalchemy import (
    TIMESTAMP,
    Date,
    Enum,
    Float,
    Index,
    Table,
    Column,
    String,
    ForeignKey,
    text,
)
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.mysql import (
    CHAR,
    ENUM,
    TEXT,
    YEAR,
    INTEGER,
    TINYINT,
    VARCHAR,
    SMALLINT,
    MEDIUMINT,
    MEDIUMBLOB,
    MEDIUMTEXT,
)
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()
metadata = Base.metadata


class ChiiCharacter(Base):
    __tablename__ = "chii_characters"

    crt_id = Column(MEDIUMINT(8), primary_key=True)
    crt_name = Column(String(255, "utf8_unicode_ci"), nullable=False)
    crt_role = Column(TINYINT(4), nullable=False, index=True, comment="角色，机体，组织。。")
    crt_infobox = Column(MEDIUMTEXT, nullable=False)
    crt_summary = Column(MEDIUMTEXT, nullable=False)
    crt_img = Column(String(255, "utf8_unicode_ci"), nullable=False)
    crt_comment = Column(MEDIUMINT(9), nullable=False, server_default=text("'0'"))
    crt_collects = Column(MEDIUMINT(8), nullable=False)
    crt_dateline = Column(INTEGER(10), nullable=False)
    crt_lastpost = Column(INTEGER(11), nullable=False)
    crt_lock = Column(
        TINYINT(4), nullable=False, index=True, server_default=text("'0'")
    )
    crt_img_anidb = Column(VARCHAR(255), nullable=False)
    crt_anidb_id = Column(MEDIUMINT(8), nullable=False)
    crt_ban = Column(TINYINT(3), nullable=False, index=True, server_default=text("'0'"))
    crt_redirect = Column(INTEGER(10), nullable=False, server_default=text("'0'"))
    crt_nsfw = Column(TINYINT(1), nullable=False)

    subjects: List["ChiiCrtSubjectIndex"] = relationship(
        "ChiiCrtSubjectIndex",
        lazy="raise_on_sql",
        # secondary=ChiiPersonCsIndex,
        back_populates="character",
    )


class ChiiCrtCastIndex(Base):
    __tablename__ = "chii_crt_cast_index"

    crt_id = Column(MEDIUMINT(9), primary_key=True, nullable=False)
    prsn_id = Column(MEDIUMINT(9), primary_key=True, nullable=False, index=True)
    subject_id = Column(MEDIUMINT(9), primary_key=True, nullable=False, index=True)
    subject_type_id = Column(
        TINYINT(3), nullable=False, index=True, comment="根据人物归类查询角色，动画，书籍，游戏"
    )
    summary = Column(
        String(255, "utf8_unicode_ci"), nullable=False, comment="幼年，男乱马，女乱马，变身形态，少女形态。。"
    )


class ChiiCrtSubjectIndex(Base):
    __tablename__ = "chii_crt_subject_index"

    crt_id = Column(
        MEDIUMINT(9),
        ForeignKey("chii_characters.crt_id"),
        primary_key=True,
        nullable=False,
    )
    subject_id = Column(
        MEDIUMINT(9),
        ForeignKey("chii_subjects.subject_id"),
        primary_key=True,
        nullable=False,
        index=True,
    )
    subject_type_id = Column(TINYINT(4), nullable=False, index=True)
    crt_type = Column(TINYINT(4), nullable=False, index=True, comment="主角，配角")
    ctr_appear_eps = Column(MEDIUMTEXT, nullable=False, comment="可选，角色出场的的章节")
    crt_order = Column(TINYINT(3), nullable=False)

    character: "ChiiCharacter" = relationship(
        "ChiiCharacter",
        lazy="raise",
        innerjoin=True,
    )
    subject: "ChiiSubject" = relationship("ChiiSubject", lazy="raise", innerjoin=True)


class ChiiEpRevision(Base):
    __tablename__ = "chii_ep_revisions"
    __table_args__ = (Index("rev_sid", "rev_sid", "rev_creator"),)

    ep_rev_id = Column(MEDIUMINT(8), primary_key=True)
    rev_sid = Column(MEDIUMINT(8), nullable=False)
    rev_eids = Column(String(255), nullable=False)
    rev_ep_infobox = Column(MEDIUMTEXT, nullable=False)
    rev_creator = Column(MEDIUMINT(8), nullable=False)
    rev_version = Column(TINYINT(1), nullable=False, server_default=text("'0'"))
    rev_dateline = Column(INTEGER(10), nullable=False)
    rev_edit_summary = Column(String(200), nullable=False)


class ChiiEpisode(Base):
    __tablename__ = "chii_episodes"
    __table_args__ = (Index("ep_subject_id_2", "ep_subject_id", "ep_ban", "ep_sort"),)

    ep_id = Column(MEDIUMINT(8), primary_key=True)
    ep_subject_id = Column(MEDIUMINT(8), nullable=False, index=True)
    ep_sort = Column(Float, nullable=False, index=True, server_default=text("'0'"))
    ep_type = Column(TINYINT(1), nullable=False)
    ep_disc = Column(
        TINYINT(3),
        nullable=False,
        index=True,
        server_default=text("'0'"),
        comment="碟片数",
    )
    ep_name = Column(String(80), nullable=False)
    ep_name_cn = Column(String(80), nullable=False)
    ep_rate = Column(TINYINT(3), nullable=False)
    ep_duration = Column(String(80), nullable=False)
    ep_airdate = Column(String(80), nullable=False)
    ep_online = Column(MEDIUMTEXT, nullable=False)
    ep_comment = Column(MEDIUMINT(8), nullable=False)
    ep_resources = Column(MEDIUMINT(8), nullable=False)
    ep_desc = Column(MEDIUMTEXT, nullable=False)
    ep_dateline = Column(INTEGER(10), nullable=False)
    ep_lastpost = Column(INTEGER(10), nullable=False, index=True)
    ep_lock = Column(TINYINT(3), nullable=False, server_default=text("'0'"))
    ep_ban = Column(TINYINT(3), nullable=False, index=True, server_default=text("'0'"))


class ChiiMemberfield(Base):
    __tablename__ = "chii_memberfields"

    uid = Column(MEDIUMINT(8), primary_key=True, server_default=text("'0'"))
    site = Column(VARCHAR(75), nullable=False, server_default=text("''"))
    location = Column(VARCHAR(30), nullable=False, server_default=text("''"))
    bio = Column(TEXT, nullable=False)
    privacy = Column(MEDIUMTEXT, nullable=False)
    blocklist = Column(MEDIUMTEXT, nullable=False)


class ChiiMember(Base):
    __tablename__ = "chii_members"

    uid = Column(MEDIUMINT(8), primary_key=True)
    username = Column(CHAR(15), nullable=False, unique=True, server_default=text("''"))
    nickname = Column(String(30), nullable=False)
    avatar = Column(VARCHAR(255), nullable=False)
    groupid = Column(SMALLINT(6), nullable=False, server_default=text("'0'"))
    regdate = Column(INTEGER(10), nullable=False, server_default=text("'0'"))
    lastvisit = Column(INTEGER(10), nullable=False, server_default=text("'0'"))
    lastactivity = Column(INTEGER(10), nullable=False, server_default=text("'0'"))
    lastpost = Column(INTEGER(10), nullable=False, server_default=text("'0'"))
    dateformat = Column(CHAR(10), nullable=False, server_default=text("''"))
    timeformat = Column(TINYINT(1), nullable=False, server_default=text("'0'"))
    timeoffset = Column(CHAR(4), nullable=False, server_default=text("''"))
    newpm = Column(TINYINT(1), nullable=False, server_default=text("'0'"))
    new_notify = Column(
        SMALLINT(6), nullable=False, server_default=text("'0'"), comment="新提醒"
    )
    sign = Column(VARCHAR(255), nullable=False)

    @classmethod
    def all_column(cls):
        """
        This table contains only column with read permission,
        don't `select *` on this table.
        """
        return cls.__table__.columns


class ChiiOauthAccessToken(Base):
    __tablename__ = "chii_oauth_access_tokens"

    access_token = Column(String(40, "utf8_unicode_ci"), primary_key=True)
    client_id = Column(String(80, "utf8_unicode_ci"), nullable=False)
    user_id = Column(String(80, "utf8_unicode_ci"))
    expires = Column(
        TIMESTAMP,
        nullable=False,
        server_default=text("CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"),
    )
    scope = Column(String(4000, "utf8_unicode_ci"))


t_chii_person_alias = Table(
    "chii_person_alias",
    metadata,
    Column("prsn_cat", ENUM("prsn", "crt"), nullable=False),
    Column("prsn_id", MEDIUMINT(9), nullable=False, index=True),
    Column("alias_name", String(255, "utf8_unicode_ci"), nullable=False),
    Column("alias_type", TINYINT(4), nullable=False),
    Column("alias_key", String(10, "utf8_unicode_ci"), nullable=False),
    Index("prsn_cat", "prsn_cat", "prsn_id"),
)


class ChiiPersonCollect(Base):
    __tablename__ = "chii_person_collects"
    __table_args__ = (
        Index("prsn_clt_cat", "prsn_clt_cat", "prsn_clt_mid"),
        {"comment": "人物收藏"},
    )

    prsn_clt_id = Column(MEDIUMINT(8), primary_key=True)
    prsn_clt_cat = Column(Enum("prsn", "crt"), nullable=False)
    prsn_clt_mid = Column(MEDIUMINT(8), nullable=False, index=True)
    prsn_clt_uid = Column(MEDIUMINT(8), nullable=False, index=True)
    prsn_clt_dateline = Column(INTEGER(10), nullable=False)


class ChiiPersonCsIndex(Base):
    __tablename__ = "chii_person_cs_index"
    __table_args__ = {"comment": "subjects' credits/creator & staff (c&s)index"}

    prsn_type = Column(ENUM("prsn", "crt"), primary_key=True, nullable=False)
    prsn_id = Column(
        MEDIUMINT(9),
        ForeignKey("chii_persons.prsn_id"),
        primary_key=True,
        nullable=False,
        index=True,
    )
    prsn_position = Column(
        SMALLINT(5), primary_key=True, nullable=False, index=True, comment="监督，原案，脚本,.."
    )
    subject_id = Column(
        MEDIUMINT(9),
        ForeignKey("chii_subjects.subject_id"),
        primary_key=True,
        nullable=False,
        index=True,
    )
    subject_type_id = Column(TINYINT(4), nullable=False, index=True)
    summary = Column(MEDIUMTEXT, nullable=False)
    prsn_appear_eps = Column(MEDIUMTEXT, nullable=False, comment="可选，人物参与的章节")

    person: "ChiiPerson" = relationship(
        "ChiiPerson",
        lazy="raise",
        innerjoin=True,
    )
    subject: "ChiiSubject" = relationship("ChiiSubject", lazy="raise", innerjoin=True)


class ChiiPersonField(Base):
    __tablename__ = "chii_person_fields"
    __table_args__ = {"extend_existing": True}

    prsn_id = Column(INTEGER(8), primary_key=True, nullable=False, index=True)
    prsn_cat = Column(ENUM("prsn", "crt"), nullable=False)
    gender = Column(TINYINT(4), nullable=False)
    bloodtype = Column(TINYINT(4), nullable=False)
    birth_year = Column(YEAR(4), nullable=False)
    birth_mon = Column(TINYINT(2), nullable=False)
    birth_day = Column(TINYINT(2), nullable=False)
    __mapper_args__ = {"polymorphic_on": prsn_cat, "polymorphic_identity": "prsn"}


class ChiiCharacterField(ChiiPersonField):
    __mapper_args__ = {"polymorphic_identity": "crt"}


t_chii_person_relationship = Table(
    "chii_person_relationship",
    metadata,
    Column("prsn_type", ENUM("prsn", "crt"), nullable=False),
    Column("prsn_id", MEDIUMINT(9), nullable=False),
    Column("relat_prsn_type", ENUM("prsn", "crt"), nullable=False),
    Column("relat_prsn_id", MEDIUMINT(9), nullable=False),
    Column("relat_type", SMALLINT(6), nullable=False, comment="任职于，从属,聘用,嫁给，"),
    Index("relat_prsn_type", "relat_prsn_type", "relat_prsn_id"),
    Index("prsn_type", "prsn_type", "prsn_id"),
)


class ChiiPerson(Base):
    __tablename__ = "chii_persons"
    __table_args__ = {"comment": "（现实）人物表"}

    prsn_id = Column(MEDIUMINT(8), primary_key=True)
    prsn_name = Column(String(255, "utf8_unicode_ci"), nullable=False)
    prsn_type = Column(TINYINT(4), nullable=False, index=True, comment="个人，公司，组合")
    prsn_infobox = Column(MEDIUMTEXT, nullable=False)
    prsn_producer = Column(TINYINT(1), nullable=False, index=True)
    prsn_mangaka = Column(TINYINT(1), nullable=False, index=True)
    prsn_artist = Column(TINYINT(1), nullable=False, index=True)
    prsn_seiyu = Column(TINYINT(1), nullable=False, index=True)
    prsn_writer = Column(
        TINYINT(4), nullable=False, index=True, server_default=text("'0'"), comment="作家"
    )
    prsn_illustrator = Column(
        TINYINT(4), nullable=False, index=True, server_default=text("'0'"), comment="绘师"
    )
    prsn_actor = Column(TINYINT(1), nullable=False, index=True, comment="演员")
    prsn_summary = Column(MEDIUMTEXT, nullable=False)
    prsn_img = Column(String(255, "utf8_unicode_ci"), nullable=False)
    prsn_img_anidb = Column(VARCHAR(255), nullable=False)
    prsn_comment = Column(MEDIUMINT(9), nullable=False)
    prsn_collects = Column(MEDIUMINT(8), nullable=False)
    prsn_dateline = Column(INTEGER(10), nullable=False)
    prsn_lastpost = Column(INTEGER(11), nullable=False)
    prsn_lock = Column(TINYINT(4), nullable=False, index=True)
    prsn_anidb_id = Column(MEDIUMINT(8), nullable=False)
    prsn_ban = Column(
        TINYINT(3), nullable=False, index=True, server_default=text("'0'")
    )
    prsn_redirect = Column(INTEGER(10), nullable=False, server_default=text("'0'"))
    prsn_nsfw = Column(TINYINT(1), nullable=False)

    subjects: List["ChiiPersonCsIndex"] = relationship(
        "ChiiPersonCsIndex",
        lazy="raise_on_sql",
        # secondary=ChiiPersonCsIndex,
        back_populates="person",
    )


class ChiiRevHistory(Base):
    __tablename__ = "chii_rev_history"
    __table_args__ = (
        Index("rev_crt_id", "rev_type", "rev_mid"),
        Index("rev_id", "rev_id", "rev_type", "rev_creator"),
    )

    rev_id = Column(MEDIUMINT(8), primary_key=True)
    rev_type = Column(TINYINT(3), nullable=False, comment="条目，角色，人物")
    rev_mid = Column(MEDIUMINT(8), nullable=False, comment="对应条目，人物的ID")
    rev_text_id = Column(MEDIUMINT(9), nullable=False)
    rev_dateline = Column(INTEGER(10), nullable=False)
    rev_creator = Column(MEDIUMINT(8), nullable=False, index=True)
    rev_edit_summary = Column(String(200, "utf8_unicode_ci"), nullable=False)


class ChiiRevText(Base):
    __tablename__ = "chii_rev_text"

    rev_text_id = Column(MEDIUMINT(9), primary_key=True)
    rev_text = Column(MEDIUMBLOB, nullable=False)


t_chii_subject_alias = Table(
    "chii_subject_alias",
    metadata,
    Column("subject_id", INTEGER(10), nullable=False, index=True),
    Column("alias_name", String(255), nullable=False),
    Column(
        "subject_type_id",
        TINYINT(3),
        nullable=False,
        server_default=text("'0'"),
        comment="所属条目的类型",
    ),
    Column(
        "alias_type",
        TINYINT(3),
        nullable=False,
        server_default=text("'0'"),
        comment="是别名还是条目名",
    ),
    Column("alias_key", VARCHAR(10), nullable=False),
)


class ChiiSubjectField(Base):
    __tablename__ = "chii_subject_fields"
    __table_args__ = (
        Index("query_date", "field_sid", "field_date"),
        Index("field_year_mon", "field_year", "field_mon"),
    )

    field_sid = Column(MEDIUMINT(8), primary_key=True)
    field_tid = Column(
        SMALLINT(6), nullable=False, index=True, server_default=text("'0'")
    )
    field_tags = Column(MEDIUMTEXT, nullable=False)
    field_rate_1 = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    field_rate_2 = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    field_rate_3 = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    field_rate_4 = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    field_rate_5 = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    field_rate_6 = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    field_rate_7 = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    field_rate_8 = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    field_rate_9 = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    field_rate_10 = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    field_airtime = Column(TINYINT(1), nullable=False, index=True)
    field_rank = Column(
        INTEGER(10), nullable=False, index=True, server_default=text("'0'")
    )
    field_year = Column(YEAR(4), nullable=False, index=True, comment="放送年份")
    field_mon = Column(TINYINT(2), nullable=False, comment="放送月份")
    field_week_day = Column(TINYINT(1), nullable=False, comment="放送日(星期X)")
    field_date = Column(Date, nullable=False, index=True, comment="放送日期")
    field_redirect = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))


class ChiiSubjectRelations(Base):
    """
    这个表带有 comment，也没有主键，所以生成器用的是 `Table` 而不是现在的class。
    """

    __tablename__ = "chii_subject_relations"
    __table_args__ = (
        Index(
            "rlt_relation_type",
            "rlt_relation_type",
            "rlt_subject_id",
            "rlt_related_subject_id",
        ),
        Index(
            "rlt_subject_id",
            "rlt_subject_id",
            "rlt_related_subject_id",
            "rlt_vice_versa",
            unique=True,
        ),
        Index(
            "rlt_related_subject_type_id", "rlt_related_subject_type_id", "rlt_order"
        ),
    )
    rlt_subject_id = Column(
        "rlt_subject_id", MEDIUMINT(8), nullable=False, comment="关联主 ID"
    )
    rlt_subject_type_id = Column(
        "rlt_subject_type_id", TINYINT(3), nullable=False, index=True
    )
    rlt_relation_type = Column(
        "rlt_relation_type", SMALLINT(5), nullable=False, comment="关联类型"
    )
    rlt_related_subject_id = Column(
        "rlt_related_subject_id", MEDIUMINT(8), nullable=False, comment="关联目标 ID"
    )
    rlt_related_subject_type_id = Column(
        "rlt_related_subject_type_id", TINYINT(3), nullable=False, comment="关联目标类型"
    )
    rlt_vice_versa = Column("rlt_vice_versa", TINYINT(1), nullable=False)
    rlt_order = Column("rlt_order", TINYINT(3), nullable=False, comment="关联排序")

    __mapper_args__ = {
        "primary_key": [rlt_subject_id, rlt_related_subject_id, rlt_vice_versa]
    }


class ChiiSubjectRevision(Base):
    __tablename__ = "chii_subject_revisions"
    __table_args__ = (
        Index("rev_subject_id", "rev_subject_id", "rev_creator"),
        Index("rev_creator", "rev_creator", "rev_id"),
    )

    rev_id = Column(MEDIUMINT(8), primary_key=True)
    rev_type = Column(
        TINYINT(3),
        nullable=False,
        index=True,
        server_default=text("'1'"),
        comment="修订类型",
    )
    rev_subject_id = Column(MEDIUMINT(8), nullable=False)
    rev_type_id = Column(SMALLINT(6), nullable=False, server_default=text("'0'"))
    rev_creator = Column(MEDIUMINT(8), nullable=False)
    rev_dateline = Column(
        INTEGER(10), nullable=False, index=True, server_default=text("'0'")
    )
    rev_name = Column(String(80), nullable=False)
    rev_name_cn = Column(String(80), nullable=False)
    rev_field_infobox = Column(MEDIUMTEXT, nullable=False)
    rev_field_summary = Column(MEDIUMTEXT, nullable=False)
    rev_vote_field = Column(MEDIUMTEXT, nullable=False)
    rev_field_eps = Column(MEDIUMINT(8), nullable=False)
    rev_edit_summary = Column(String(200), nullable=False)
    rev_platform = Column(SMALLINT(6), nullable=False)


class ChiiSubject(Base):
    __tablename__ = "chii_subjects"
    __table_args__ = (
        Index(
            "order_by_name",
            "subject_ban",
            "subject_type_id",
            "subject_series",
            "subject_platform",
            "subject_name",
        ),
        Index(
            "browser",
            "subject_ban",
            "subject_type_id",
            "subject_series",
            "subject_platform",
        ),
        Index("subject_idx_cn", "subject_idx_cn", "subject_type_id"),
    )

    subject_id = Column(MEDIUMINT(8), primary_key=True)
    subject_type_id = Column(
        SMALLINT(6), nullable=False, index=True, server_default=text("'0'")
    )
    subject_name = Column(String(80), nullable=False, index=True)
    subject_name_cn = Column(String(80), nullable=False, index=True)
    subject_uid = Column(String(20), nullable=False, comment="isbn / imdb")
    subject_creator = Column(MEDIUMINT(8), nullable=False, index=True)
    subject_dateline = Column(INTEGER(10), nullable=False, server_default=text("'0'"))
    subject_image = Column(String(255), nullable=False)
    subject_platform = Column(
        SMALLINT(6), nullable=False, index=True, server_default=text("'0'")
    )
    field_infobox = Column(MEDIUMTEXT, nullable=False)
    field_summary = Column(MEDIUMTEXT, nullable=False, comment="summary")
    field_5 = Column(MEDIUMTEXT, nullable=False, comment="author summary")
    field_volumes = Column(
        MEDIUMINT(8), nullable=False, server_default=text("'0'"), comment="卷数"
    )
    field_eps = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    subject_wish = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    subject_collect = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    subject_doing = Column(MEDIUMINT(8), nullable=False, server_default=text("'0'"))
    subject_on_hold = Column(
        MEDIUMINT(8), nullable=False, server_default=text("'0'"), comment="搁置人数"
    )
    subject_dropped = Column(
        MEDIUMINT(8), nullable=False, server_default=text("'0'"), comment="抛弃人数"
    )
    subject_series = Column(
        TINYINT(1), nullable=False, index=True, server_default=text("'0'")
    )
    subject_series_entry = Column(
        MEDIUMINT(8), nullable=False, index=True, server_default=text("'0'")
    )
    subject_idx_cn = Column(String(1), nullable=False)
    subject_airtime = Column(TINYINT(1), nullable=False, index=True)
    subject_nsfw = Column(TINYINT(1), nullable=False, index=True)
    subject_ban = Column(
        TINYINT(1), nullable=False, index=True, server_default=text("'0'")
    )

    persons: List[ChiiPersonCsIndex] = relationship(
        "ChiiPersonCsIndex",
        lazy="raise",
        # secondary=ChiiPersonCsIndex,
        back_populates="subject",
    )
    characters: List[ChiiCrtSubjectIndex] = relationship(
        "ChiiCrtSubjectIndex",
        lazy="raise",
        # secondary=ChiiPersonCsIndex,
        back_populates="subject",
    )
