import pytest

from pol.db.const import Gender, StaffMap


def test_valid_type():
    keys = []
    for staff_job in StaffMap.values():
        for key, value in staff_job.items():
            keys.append(key)
            assert isinstance(key, int)
            assert value.cn or value.jp or value.en or value.rdf

    assert len(keys) == len(set(keys))


def test_gender():
    assert Gender(1).str() == "male"
    assert Gender(2).str() == "female"
    with pytest.raises(ValueError):
        Gender(3).str()
