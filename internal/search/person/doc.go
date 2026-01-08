package person

import (
	"strconv"

	wiki "github.com/bangumi/wiki-parser-go"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/search/searcher"
)

type document struct {
	ID      model.PersonID `json:"id"`
	Name    string         `json:"name" searchable:"true"`
	Aliases []string       `json:"aliases,omitempty" searchable:"true"`
	Comment uint32         `json:"comment" sortable:"true"`
	Collect uint32         `json:"collect" sortable:"true"`
	Career  []string       `json:"career,omitempty" filterable:"true"`
}

func (d *document) GetID() string {
	return strconv.FormatUint(uint64(d.ID), 10)
}

func rankRule() *[]string {
	return &[]string{
		// 相似度最优先
		"exactness",
		"words",
		"typo",
		"proximity",
		"attribute",
		"sort",
		"id:asc",
		"comment:desc",
		"collect:desc",
	}
}

func extract(c *model.Person) searcher.Document {
	w := wiki.ParseOmitError(c.Infobox)

	return &document{
		ID:      c.ID,
		Name:    c.Name,
		Aliases: searcher.ExtractAliases(w),
		Comment: c.CommentCount,
		Collect: c.CollectCount,
		Career:  c.Careers(),
	}
}
