package character

import (
	"context"
	"errors"
	"strconv"

	"github.com/meilisearch/meilisearch-go"
	"github.com/samber/lo"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
)

func (c *client) OnAdded(ctx context.Context, id model.CharacterID) error {
	s, err := c.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return nil
		}
		return errgo.Wrap(err, "characterRepo.Get")
	}

	if s.Redirect != 0 {
		return c.OnDelete(ctx, id)
	}

	extracted := extract(&s)

	_, err = c.index.UpdateDocumentsWithContext(ctx, extracted, &meilisearch.DocumentOptions{PrimaryKey: lo.ToPtr("id")})
	return err
}

func (c *client) OnUpdate(ctx context.Context, id model.CharacterID) error {
	s, err := c.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return nil
		}
		return errgo.Wrap(err, "characterRepo.Get")
	}

	if s.Redirect != 0 {
		return c.OnDelete(ctx, id)
	}

	extracted := extract(&s)

	_, err = c.index.UpdateDocumentsWithContext(ctx, extracted, &meilisearch.DocumentOptions{PrimaryKey: lo.ToPtr("id")})

	return err
}

func (c *client) OnDelete(ctx context.Context, id model.CharacterID) error {
	_, err := c.index.DeleteDocumentWithContext(ctx, strconv.FormatUint(uint64(id), 10), nil)

	return errgo.Wrap(err, "search")
}
