package web

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/jmoiron/sqlx"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/bangumi/server/graph"
)

func gql(db *sqlx.DB) (*handler.Server, error) {
	r, err := graph.NewResolver(db)
	if err != nil {
		return nil, err
	}
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: r}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](10000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](1000),
	})

	return srv, nil
}
