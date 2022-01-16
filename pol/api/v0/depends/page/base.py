from __future__ import annotations

import abc

from pydantic.main import BaseModel


class BasePage(BaseModel):
    @classmethod
    def from_request(cls, *args) -> BasePage:
        raise Exception("not implemented!")

    @abc.abstractmethod
    def modify_query(self, query) -> None:
        """
        modifies the query with pagination parameters
        :param query: sqlalchemy query object
        :return:
        """
