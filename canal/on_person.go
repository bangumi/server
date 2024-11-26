package canal

import (
	"context"
	"encoding/json"

	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/search"
)

type PersonKey struct {
	ID model.PersonID `json:"prsn_id"`
}

func (e *eventHandler) OnPerson(ctx context.Context, key json.RawMessage, payload Payload) error {
	var k PersonKey
	if err := json.Unmarshal(key, &k); err != nil {
		return err
	}
	return e.onPersonChange(ctx, k.ID, payload.Op)
}

func (e *eventHandler) onPersonChange(ctx context.Context, personID model.PersonID, op string) error {
	switch op {
	case opCreate:
		if err := e.search.EventAdded(ctx, personID, search.SearchTargetPerson); err != nil {
			return errgo.Wrap(err, "search.OnPersonAdded")
		}
	case opUpdate, opSnapshot:
		if err := e.search.EventUpdate(ctx, personID, search.SearchTargetPerson); err != nil {
			return errgo.Wrap(err, "search.OnPersonUpdate")
		}
	case opDelete:
		if err := e.search.EventDelete(ctx, personID, search.SearchTargetPerson); err != nil {
			return errgo.Wrap(err, "search.OnPersonDelete")
		}
	default:
		e.log.Warn("unexpected operator", zap.String("op", op))
	}
	return nil
}
