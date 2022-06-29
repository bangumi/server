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

package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
)

func (app App) UpdateEpisodeCollection(
	ctx context.Context,
	userID model.UserID,
	episodeID model.EpisodeID,
	collectionType model.EpisodeCollectionType,
) error {
	episode, err := app.episode.Get(ctx, episodeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrEpisodeNotFound
		}

		return errgo.Wrap(err, "episodeRepo.Get")
	}

	s, err := app.subject.Get(ctx, episode.SubjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			app.log.Error("unexpected err, missing subject for episode")
			return domain.ErrSubjectNotFound
		}

		return errgo.Wrap(err, "subjectRepo.Get")
	}

	if s.Redirect != 0 {
		return fmt.Errorf("%w: subject is removed", domain.ErrSubjectNotFound)
	}

	if err = validateSubjectCollectionRequest(s); err != nil {
		return err
	}

	_, err = app.collect.GetSubjectCollection(ctx, userID, s.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrSubjectNotCollected
		}

		return errgo.Wrap(err, "collection.GetSubjectCollection")
	}

	if err = app.collect.UpdateEpisodeCollection(ctx, userID, s.ID, episodeID, collectionType, time.Now()); err != nil {
		return errgo.Wrap(err, "failed to update episode collection")
	}

	return nil
}

func validateSubjectCollectionRequest(s model.Subject) error {
	if s.TypeID == model.SubjectTypeAnime || s.TypeID == model.SubjectTypeReal {
		return nil
	}

	return fmt.Errorf("%w: subject is not anime or real subject", domain.ErrInvalidInput)
}
