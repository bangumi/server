package graph

import (
	"github.com/jmoiron/sqlx"

	"github.com/bangumi/server/dal/query"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	db *sqlx.DB
	q  *query.Query
}

func NewResolver(db *sqlx.DB, q *query.Query) (*Resolver, error) {
	return &Resolver{db: db, q: q}, nil
}
