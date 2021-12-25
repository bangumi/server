from typing import Any, List, Tuple, Optional
from asyncio import gather
from functools import reduce

import sqlalchemy as sa
from databases import Database

from pol import curd
from pol.db.tables import ChiiMember, ChiiRevText, ChiiRevHistory
from pol.api.v0.utils import raise_offset_over_total
from pol.api.v0.models import Order, Pager


async def get_revisions(
    db: Database,
    filters: List[Any],
    columns: List[Any],
    join_args: List[Tuple[Any, Any]],
    page: Pager,
    uid: int = 0,
    order: Order = Order.desc,
):
    filters = [*filters]
    if uid > 0:
        filters.append(ChiiRevHistory.rev_creator == uid)
    count = await db.fetch_val(
        sa.select(sa.func.count(ChiiRevHistory.rev_id)).where(*filters)
    )
    if page.offset > count:
        raise_offset_over_total(count)
    columns = [
        ChiiRevHistory.rev_id,
        ChiiRevHistory.rev_type,
        ChiiRevHistory.rev_creator,
        ChiiRevHistory.rev_dateline,
        ChiiRevHistory.rev_edit_summary,
        *columns,
    ]
    join_args = [*join_args]
    if uid <= 0:
        columns.extend([ChiiMember.nickname, ChiiMember.avatar])
        join_args.append((ChiiMember, ChiiRevHistory.rev_creator == ChiiMember.uid))
    sort_field = ChiiRevHistory.rev_id
    if order == Order.desc:
        sort_field.desc()
    else:
        sort_field.asc()
    query = (
        sa.select(
            *columns,
        )
        .where(*filters)
        .order_by(sort_field)
        .limit(page.limit)
        .offset(page.offset)
    )
    query = reduce(lambda acc, item: acc.join(item[0], item[1]), join_args, query)
    revisions = [
        {
            "id": r["rev_id"],
            "type": r["rev_type"],
            "timestamp": r["rev_dateline"],
            "summary": r["rev_edit_summary"],
            "creator": {
                "id": r["rev_creator"],
                **(
                    {
                        "nickname": r["nickname"],
                        "avatar": r["avatar"],
                    }
                    if uid <= 0
                    else {}
                ),
            },
        }
        for r in await db.fetch_all(query)
    ]
    return {
        "limit": page.limit,
        "offset": page.offset,
        "data": revisions,
        "total": count,
    }


async def get_revision(
    db: Database,
    filters: List[Any],
    details: Optional[Any] = None,
):
    r = await curd.get_one(
        db,
        ChiiRevHistory,
        *filters,
        raise_404=True,
        details=details,
    )
    results: Tuple[ChiiMember, ChiiRevText] = await gather(
        curd.get_one(
            db,
            ChiiMember,
            ChiiMember.uid == r.rev_creator,
            raise_404=True,
            details={"rev_creator": r.rev_creator},
        ),
        curd.get_one(
            db,
            ChiiRevText,
            ChiiRevText.rev_text_id == r.rev_text_id,
            raise_404=True,
            details={"rev_text_id": r.rev_text_id},
        ),
    )
    user, text_item = results
    return {
        "id": r.rev_id,
        "type": r.rev_type,
        "timestamp": r.rev_dateline,
        "summary": r.rev_edit_summary,
        "data": text_item.rev_text,
        "creator": {
            "id": r.rev_creator,
            "nickname": user.nickname,
            "avatar": user.avatar,
        },
    }
