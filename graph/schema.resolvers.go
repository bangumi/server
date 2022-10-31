package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"reflect"

	"go.uber.org/zap"

	"github.com/bangumi/server/graph/generated"
	"github.com/bangumi/server/graph/gql"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/gmap"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/subject"
)

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*gql.Me, error) {
	user := ctx.Value(CurrentUserKey)
	if user == nil {
		return nil, nil
	}

	if u, ok := user.(domain.Auth); ok {
		return &gql.Me{
			ID:       int(u.ID),
			UserName: fmt.Sprint("username", u.ID),
			NickName: fmt.Sprint("nickname", u.ID),
		}, nil
	}

	r.log.Error("unexpected user type", zap.String("type", reflect.TypeOf(user).String()))
	return nil, fmt.Errorf("BUG: unexpected user type")
}

// Subject is the resolver for the subject field.
func (r *queryResolver) Subject(ctx context.Context, id int) (*gql.Subject, error) {
	s, err := r.subject.Get(ctx, model.SubjectID(id), subject.Filter{})
	if err != nil {
		return nil, err
	}

	return &gql.Subject{
		ID:     int(s.ID),
		Name:   s.Name,
		NameCn: s.NameCN,
	}, nil
}

// Subjects is the resolver for the subjects field.
func (r *queryResolver) Subjects(ctx context.Context, id []int) ([]*gql.Subject, error) {
	s, err := r.subject.GetByIDs(ctx,
		slice.Map(id, func(i int) model.SubjectID { return model.SubjectID(i) }), subject.Filter{})
	if err != nil {
		return nil, err
	}

	return slice.Map(gmap.Values(s), func(i model.Subject) *gql.Subject {
		return &gql.Subject{
			ID:     int(i.ID),
			Name:   i.Name,
			NameCn: i.NameCN,
		}
	}), nil
}

// Episodes is the resolver for the episodes field.
func (r *subjectResolver) Episodes(
	ctx context.Context,
	obj *gql.Subject,
	limit int,
	offset int,
) ([]*gql.Episode, error) {
	if obj == nil {
		return nil, nil
	}

	episodes, err := r.episode.List(ctx, model.SubjectID(obj.ID), domain.EpisodeFilter{}, limit, offset)
	if err != nil {
		return nil, err
	}

	return slice.Map(episodes, func(e model.Episode) *gql.Episode {
		return &gql.Episode{ID: int(e.ID)}
	}), nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subject returns generated.SubjectResolver implementation.
func (r *Resolver) Subject() generated.SubjectResolver { return &subjectResolver{r} }

type queryResolver struct{ *Resolver }
type subjectResolver struct{ *Resolver }
