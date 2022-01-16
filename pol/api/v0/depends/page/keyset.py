from pol.api.v0.depends.page.base import BasePage


class KeysetPage(BasePage):
    """
    pagination by keyset (better perf, but only navigable by prev and next page)
    https://github.com/djrobstep/sqlakeyset
    """

    # todo: implementation
    def from_request(self, *args) -> BasePage:
        raise Exception("not implemented")

    def modify_query(self, query) -> None:
        raise Exception("not implemented")
