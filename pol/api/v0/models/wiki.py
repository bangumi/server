from typing import List, Union

from pydantic import BaseModel


class V(BaseModel):
    v: str


class KV(BaseModel):
    k: str
    v: str


class Item(BaseModel):
    key: str
    value: Union[str, List[Union[KV, V]]]


Wiki = List[Item]
