import pytest

from .parser import WikiSyntaxError, kv, parse


def test_key_value():
    raw = """{{Infobox animanga/TVAnime
|中文名= Code Geass 反叛的鲁路修R2
|其他=
|Copyright= （C）2006 SUNRISE inc./MBS
}}"""
    assert parse(raw).info == [
        kv("中文名", "Code Geass 反叛的鲁路修R2"),
        kv("其他"),
        kv("Copyright", "（C）2006 SUNRISE inc./MBS"),
    ]


def test_key_array():
    raw = """{{Infobox animanga/TVAnime
|中文名= Code Geass 反叛的鲁路修R2
|别名={
[叛逆的鲁路修R2]
[Code Geass: Hangyaku no Lelouch R2]
[叛逆的勒鲁什R2]
[叛逆的鲁鲁修R2]
[コードギアス 反逆のルルーシュR2]
[Code Geass: Lelouch of the Rebellion R2]
[叛逆的勒路什R2]
}
|话数=25
}}"""
    assert parse(raw).info == [
        kv("中文名", "Code Geass 反叛的鲁路修R2"),
        kv(
            "别名",
            [
                v("叛逆的鲁路修R2"),
                v("Code Geass: Hangyaku no Lelouch R2"),
                v("叛逆的勒鲁什R2"),
                v("叛逆的鲁鲁修R2"),
                v("コードギアス 反逆のルルーシュR2"),
                v("Code Geass: Lelouch of the Rebellion R2"),
                v("叛逆的勒路什R2"),
            ],
        ),
        kv("话数", "25"),
    ]


def test_new_line_array_body():
    raw = """{{Infobox Game
|中文名= 足球经理2009
|别名={
}
|平台=
{
[PC]
[Mac]
[PSP]
}
}}
"""
    assert parse(raw).info == [
        kv("中文名", "足球经理2009"),
        kv("别名", []),
        kv(
            "平台",
            [
                {"v": "PC"},
                {"v": "Mac"},
                {"v": "PSP"},
            ],
        ),
    ]


def test_array_with_key():
    raw = """{{Infobox Game
|平台={
[1|PC]
[2|Mac]
[PSP]
}
}}
"""
    assert parse(raw).info == [
        kv(
            "平台",
            [
                {"k": "1", "v": "PC"},
                {"k": "2", "v": "Mac"},
                {"v": "PSP"},
            ],
        )
    ]


def test_example_2():
    raw = """{{Infobox Album
|中文名=
|别名={
}
|版本特性=
|发售日期= 2008.9.24
|价格=
|播放时长= 89M
|录音= GENEON ENTERTAINMENT,INC(PLC)(M)
|碟片数量= 2CD
|艺术家= {
KOTOKO
MELL
詩月カオリ
川田まみ
島みやえい子
}
}}"""
    with pytest.raises(WikiSyntaxError, match="wiki item should"):
        parse(raw)


def test_example_1():
    raw = """{{Infobox real/Television
{
|中文名= 最强的名医 2013 SP
|别名={
}
|集数= 1
|放送星期={
[土曜日]
[日曜日]
}
|开始={
[2013年6月1日]
[2020年4月19日]
}
|结束= 剧情
|类型=
}}"""
    with pytest.raises(WikiSyntaxError, match='unexpected "{"'):
        parse(raw)


def test_a():
    raw = """{{Infobox animanga/Manga
|中文名= 火之鳥 復刻版 08 亂世篇 下·羽衣篇
|别名={
}
|出版社={
[朝日新聞出版]
[台灣東販]
}
|价格={
[￥ 1,365]
[NT$240]
}
|其他出版社=
|连载杂志=
|发售日={
[2009-8-20]
[2012-12-27]
}
|册数=
|页数={
[327]
[332]
}
|话数=
|ISBN={
[4022140291]
[9789862519295]
}
|其他=
}}"""

    parse(raw)


def test_no_error():
    raw = """{{Infobox animanga/Book
|中文名=
|别名= {

}
|作者=荒川 弘
|译者=
|出版社=スクウェア・エニックス
|价格=￥ 420
|插图=
|连载杂志=
|发售日=2006-03-22
|页数=185
|话数=
|ISBN=475751638X
|其他=
}}"""
    parse(raw)


def test_multi_array():
    raw = """{{Infobox Album
|中文名=
|别名={
}
|版本特性=
|发售日期= 2014-04-27 (M3-33)
|价格= 1000 yen
|播放时长=
|录音= ハッカドロップ。
|碟片数量= 1
|Vocal= 紗智
|Illustration= 真琴
|Lyrics={
[紗智]
[水城さえ]
[タキモトショウ]
[りでる]
[mintea]
[Cororo]
}
|Music={
[KA=YA]
[三滝航]
[塵屑れお]
[Yuy]
[タキモトショウ]
[りでる]
[sonoa]
[Cororo]
}
|特设= http://haccadrop.chu.jp/favoritte/?spm=a1z1s.6659513.0.0.cV15ab
}}"""
    assert parse(raw).info == [
        kv("中文名"),
        kv("别名", []),
        kv("版本特性"),
        kv("发售日期", "2014-04-27 (M3-33)"),
        kv("价格", "1000 yen"),
        kv("播放时长"),
        kv("录音", "ハッカドロップ。"),
        kv("碟片数量", "1"),
        kv("Vocal", "紗智"),
        kv("Illustration", "真琴"),
        kv(
            "Lyrics",
            [v(x) for x in ["紗智", "水城さえ", "タキモトショウ", "りでる", "mintea", "Cororo"]],
        ),
        kv(
            "Music",
            [
                v("KA=YA"),
                v("三滝航"),
                v("塵屑れお"),
                v("Yuy"),
                v("タキモトショウ"),
                v("りでる"),
                v("sonoa"),
                v("Cororo"),
            ],
        ),
        kv("特设", "http://haccadrop.chu.jp/favoritte/?spm=a1z1s.6659513.0.0.cV15ab"),
    ]


def test_error_message_missing_key():
    raw = """{{Infobox animanga/Book
|中文名=
|别名= {

}
|作者=
荒川 弘
|译者=
}}"""
    with pytest.raises(WikiSyntaxError, match="missing key or unexpected line break "):
        parse(raw)


def v(vv: str) -> dict:
    return {"v": vv}
