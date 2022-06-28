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

package collection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
)

func NewService(
	r domain.CollectionRepo,
	s domain.SubjectRepo,
	episode domain.EpisodeRepo,
) (domain.CollectionService, error) {
	return service{subject: s, repo: r, episode: episode}, nil
}

type service struct {
	subject domain.SubjectRepo
	repo    domain.CollectionRepo
	episode domain.EpisodeRepo
}

func (s service) CountSubjectCollections(
	ctx context.Context,
	userID model.UserID,
	subjectType model.SubjectType,
	collectionType model.CollectionType,
	showPrivate bool,
) (int64, error) {
	return s.repo.CountSubjectCollections(ctx, userID, subjectType, collectionType, showPrivate) //nolint:wrapcheck
}

func (s service) ListSubjectCollection(
	ctx context.Context,
	userID model.UserID,
	subjectType model.SubjectType,
	collectionType model.CollectionType,
	showPrivate bool,
	limit, offset int,
) ([]model.SubjectCollection, error) {
	//nolint:wrapcheck
	return s.repo.ListSubjectCollection(ctx, userID, subjectType, collectionType, showPrivate, limit, offset)
}

func (s service) GetSubjectCollection(
	ctx context.Context,
	userID model.UserID,
	subjectID model.SubjectID,
) (model.SubjectCollection, error) {
	return s.repo.GetSubjectCollection(ctx, userID, subjectID) //nolint:wrapcheck
}

func (s service) UpdateEpisodeCollection(
	ctx context.Context,
	userID model.UserID,
	subjectID model.SubjectID,
	episodeID model.EpisodeID,
	collectionType model.EpisodeCollectionType,
) error {
	_, err := s.subject.Get(ctx, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrSubjectNotFound
		}

		return errgo.Wrap(err, "subjectRepo.Get")
	}

	episode, err := s.episode.Get(ctx, episodeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrEpisodeNotFound
		}

		return errgo.Wrap(err, "episodeRepo.Get")
	}

	if episode.SubjectID != subjectID {
		return fmt.Errorf("%w: episode(%d) is not belong to subject(%d)", domain.ErrInvalidInput, episodeID, subjectID)
	}

	err = s.repo.UpdateEpisodeCollection(ctx, userID, subjectID, episodeID, collectionType, time.Now())
	return errgo.Wrap(err, "collectionRepo.UpdateEpisodeCollection")
}
