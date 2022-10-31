package graph

import (
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/subject"
)

var userKeyStub = "user_key"

var CurrentUserKey = &userKeyStub

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	log     *zap.Logger
	subject subject.Repo
	episode domain.EpisodeRepo
}

func NewResolver(
	episode domain.EpisodeRepo,
	subjectRepo subject.Repo,
	log *zap.Logger,
) Resolver {
	return Resolver{
		episode: episode,
		subject: subjectRepo,
		log:     log.Named("GraphQL.resolver"),
	}
}
