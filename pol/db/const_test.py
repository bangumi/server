from pol.db.const import StaffMap


def test_valid_type():
    keys = []
    for staff_job in StaffMap.values():
        for key, value in staff_job.items():
            keys.append(key)
            assert isinstance(key, int)
            assert value.cn or value.jp or value.en or value.rdf

    assert len(keys) == len(set(keys))
