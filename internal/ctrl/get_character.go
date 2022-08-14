// SPDX-License-Identifier: AGPL-3.0-only
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>

package ctrl

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/ctrl/internal/cachekey"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func (ctl Ctrl) GetCharacter(ctx context.Context, user domain.Auth, id model.CharacterID) (model.Character, error) {
	character, err := ctl.getCharacter(ctx, id)
	if err != nil {
		return model.Character{}, err
	}

	if !auth.AllowReadCharacter(user, character) {
		return model.Character{}, domain.ErrCharacterNotFound
	}

	return character, nil
}

func (ctl Ctrl) GetCharacterNoRedirect(
	ctx context.Context, user domain.Auth, id model.CharacterID,
) (model.Character, error) {
	character, err := ctl.GetCharacter(ctx, user, id)
	if err != nil {
		return model.Character{}, err
	}

	if character.Redirect != 0 {
		return model.Character{}, domain.ErrCharacterNotFound
	}

	return character, nil
}

func (ctl Ctrl) getCharacter(ctx context.Context, id model.CharacterID) (model.Character, error) {
	var key = cachekey.Character(id)

	// try to read from cache
	var r model.Character
	ok, err := ctl.cache.Get(ctx, key, &r)
	if err != nil {
		return r, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		return r, nil
	}

	r, err = ctl.character.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return r, domain.ErrPersonNotFound
		}

		return r, errgo.Wrap(err, "personRepo.Get")
	}

	if e := ctl.cache.Set(ctx, key, r, time.Minute); e != nil {
		ctl.log.Error("can't set response to cache", zap.Error(e))
	}

	return r, nil
}
