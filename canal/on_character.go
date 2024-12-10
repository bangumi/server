package canal

import (
	"context"
	"encoding/json"

	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/search"
)

type CharacterKey struct {
	ID model.CharacterID `json:"crt_id"`
}

func (e *eventHandler) OnCharacter(ctx context.Context, key json.RawMessage, payload Payload) error {
	var k CharacterKey
	if err := json.Unmarshal(key, &k); err != nil {
		return err
	}
	return e.onCharacterChange(ctx, k.ID, payload.Op)
}

func (e *eventHandler) onCharacterChange(ctx context.Context, characterID model.CharacterID, op string) error {
	switch op {
	case opCreate:
		if err := e.search.EventAdded(ctx, characterID, search.SearchTargetCharacter); err != nil {
			return errgo.Wrap(err, "search.OnCharacterAdded")
		}
	case opUpdate, opSnapshot:
		if err := e.search.EventUpdate(ctx, characterID, search.SearchTargetCharacter); err != nil {
			return errgo.Wrap(err, "search.OnCharacterUpdate")
		}
	case opDelete:
		if err := e.search.EventDelete(ctx, characterID, search.SearchTargetCharacter); err != nil {
			return errgo.Wrap(err, "search.OnCharacterDelete")
		}
	default:
		e.log.Warn("unexpected operator", zap.String("op", op))
	}

	return nil
}
