package person

import (
	"context"
	"errors"

	"github.com/samber/lo"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/search/searcher"
)

func (c *client) upsertDocument(ctx context.Context, id model.PersonID) error {
	s, err := c.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return nil
		}
		return errgo.Wrap(err, "personRepo.Get")
	}

	if s.Redirect != 0 {
		return c.OnDelete(ctx, id)
	}

	extracted := extract(&s)

	_, err = c.index.UpdateDocumentsWithContext(ctx, extracted, lo.ToPtr("id"))
	return err
}

func (c *client) OnAdded(ctx context.Context, id model.PersonID) error {
	return c.upsertDocument(ctx, id)
}

func (c *client) OnUpdate(ctx context.Context, id model.PersonID) error {
	return c.upsertDocument(ctx, id)
}

func (c *client) OnDelete(ctx context.Context, id model.PersonID) error {
	return searcher.DeleteDocument(ctx, c.index, uint32(id))
}
