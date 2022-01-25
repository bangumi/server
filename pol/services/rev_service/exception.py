from pol.curd import NotFoundError

__all__ = ["RevisionNotFound"]


class RevisionNotFound(NotFoundError):
    pass
