from pathlib import Path

import pytest

from pol.compat.phpseralize import loads, dict_to_list

fixtures_path = Path(__file__).parent.joinpath("fixtures")


def test_loads():
    assert loads(b"a:4:{i:1;i:2;i:30;i:2;i:20;i:2;i:21;i:0;}") == {
        1: 2,
        30: 2,
        20: 2,
        21: 0,
    }


def test_loads_2():
    assert dict_to_list(
        loads(fixtures_path.joinpath("subject_8_tags.txt").read_bytes().strip())
    ) == [
        {"result": "1645", "tag_name": "叛逆的鲁鲁修"},
        {"result": "1229", "tag_name": "SUNRISE"},
        {"result": "936", "tag_name": "反逆のルルーシュ"},
        {"result": "721", "tag_name": "还是死妹控"},
        {"result": "664", "tag_name": "TV"},
        {"result": "603", "tag_name": "妹控"},
        {"result": "569", "tag_name": "codegeass"},
        {"result": "523", "tag_name": "谷口悟朗"},
        {"result": "453", "tag_name": "鲁路修"},
        {"result": "427", "tag_name": "R2"},
        {"result": "409", "tag_name": "2008"},
        {"result": "385", "tag_name": "原创"},
        {"result": "357", "tag_name": "2008年4月"},
        {"result": "174", "tag_name": "大河内一楼"},
        {"result": "151", "tag_name": "日升"},
        {"result": "120", "tag_name": "萝卜"},
        {"result": "111", "tag_name": "机战"},
        {"result": "104", "tag_name": "狗得鸡鸭死"},
        {"result": "94", "tag_name": "福山润"},
        {"result": "84", "tag_name": "露露胸"},
        {"result": "69", "tag_name": "CLAMP"},
        {"result": "67", "tag_name": "科幻"},
        {"result": "62", "tag_name": "鲁鲁修"},
        {"result": "57", "tag_name": "GEASS"},
        {"result": "54", "tag_name": "神作"},
        {"result": "49", "tag_name": "战斗"},
        {"result": "41", "tag_name": "战争"},
        {"result": "40", "tag_name": "裸露修的跌二次KUSO"},
        {"result": "37", "tag_name": "中二"},
        {"result": "34", "tag_name": "樱井孝宏"},
    ]


def test_loads_null():
    assert loads(fixtures_path.joinpath("with_null.txt").read_bytes().strip()) == {
        1: None,
        2: {0: 1, 1: 4.5, 2: 3},
    }


def test_loads_disallow_object():
    with pytest.raises(ValueError, match="php object"):
        loads(fixtures_path.joinpath("disallow_object.txt").read_bytes().strip())


def test_loads_bool():
    assert loads(fixtures_path.joinpath("bool.txt").read_bytes().strip()) == {1: True}
