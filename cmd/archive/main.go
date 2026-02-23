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

//nolint:forbidigo,funlen
package archive

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/trim21/errgo"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/dal"
	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/driver"
	"github.com/bangumi/server/internal/pkg/logger"
	subjectDto "github.com/bangumi/server/internal/subject"
)

const defaultStep = 50

var out string

var Command = &cobra.Command{
	Use:   "archive",
	Short: "create a wiki dump",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := mysql.SetLogger(logger.Std()); err != nil {
			return errgo.Wrap(err, "can't replace mysql driver's errLog")
		}

		fmt.Println("dumping data with args:", args)

		start(out)

		return nil
	},
}

func init() {
	Command.Flags().StringVar(&out, "out", "archive.zip", "zip file output location")
}

var ctx = context.Background() //nolint:gochecknoglobals

var maxSubjectID model.SubjectID     //nolint:gochecknoglobals
var maxCharacterID model.CharacterID //nolint:gochecknoglobals
var maxPersonID model.PersonID       //nolint:gochecknoglobals

func start(out string) {
	var q *query.Query
	err := fx.New(
		fx.NopLogger,
		fx.Provide(
			driver.NewMysqlDriver, dal.NewGormDB,

			config.NewAppConfig, logger.Copy,

			query.Use,
		),

		fx.Populate(&q),
	).Err()

	if err != nil {
		logger.Err(err, "failed to fill deps")
	}

	getMaxID(q)

	abs, err := filepath.Abs(out)
	if err != nil {
		logger.Fatal("failed to get output file full path", zap.Error(err))
	}

	fmt.Println(abs)

	f, err := os.Create(abs)
	if err != nil {
		panic(err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logger.Err(err, "failed to close of tile")
		}
	}(f)

	z := zip.NewWriter(f)
	defer func(z *zip.Writer) {
		err := z.Close()
		if err != nil {
			logger.Err(err, "failed to close zip writter")
		}
	}(z)

	for _, s := range []struct {
		FileName string
		Fn       func(q *query.Query, w io.Writer)
	}{
		{FileName: "subject.jsonlines", Fn: exportSubjects},
		{FileName: "person.jsonlines", Fn: exportPersons},
		{FileName: "character.jsonlines", Fn: exportCharacters},
		{FileName: "episode.jsonlines", Fn: exportEpisodes},
		{FileName: "subject-relations.jsonlines", Fn: exportSubjectRelations},
		{FileName: "subject-persons.jsonlines", Fn: exportSubjectPersonRelations},
		{FileName: "subject-characters.jsonlines", Fn: exportSubjectCharacterRelations},
		{FileName: "person-characters.jsonlines", Fn: exportPersonCharacterRelations},
		{FileName: "person-relations.jsonlines", Fn: exportPersonRelations},
	} {
		w, err := z.Create(s.FileName)
		if err != nil {
			panic(err)
		}

		s.Fn(q, w)
	}

	fmt.Println("finish exporting")
}

func getMaxID(q *query.Query) {
	lastSubject, err := q.WithContext(ctx).Subject.Order(q.Subject.ID.Desc()).Take()
	if err != nil {
		panic(err)
	}
	maxSubjectID = lastSubject.ID

	lastCharacter, err := q.WithContext(ctx).Character.Order(q.Character.ID.Desc()).Take()
	if err != nil {
		panic(err)
	}
	maxCharacterID = lastCharacter.ID

	lastPerson, err := q.WithContext(ctx).Person.Order(q.Person.ID.Desc()).Take()
	if err != nil {
		panic(err)
	}
	maxPersonID = lastPerson.ID
}

type Score struct {
	Field1  uint32 `json:"1"`
	Field2  uint32 `json:"2"`
	Field3  uint32 `json:"3"`
	Field4  uint32 `json:"4"`
	Field5  uint32 `json:"5"`
	Field6  uint32 `json:"6"`
	Field7  uint32 `json:"7"`
	Field8  uint32 `json:"8"`
	Field9  uint32 `json:"9"`
	Field10 uint32 `json:"10"`
}

type Favorite struct {
	Wish    uint32 `json:"wish"`
	Done    uint32 `json:"done"`
	Doing   uint32 `json:"doing"`
	OnHold  uint32 `json:"on_hold"`
	Dropped uint32 `json:"dropped"`
}

type Subject struct {
	ID       model.SubjectID   `json:"id"`
	Type     model.SubjectType `json:"type"`
	Name     string            `json:"name"`
	NameCN   string            `json:"name_cn"`
	Infobox  string            `json:"infobox"`
	Platform uint16            `json:"platform"`
	Summary  string            `json:"summary"`
	Nsfw     bool              `json:"nsfw"`

	Tags         []Tag    `json:"tags"`
	MetaTags     []string `json:"meta_tags"`
	Score        float64  `json:"score"`
	ScoreDetails Score    `json:"score_details"`
	Rank         uint32   `json:"rank"`
	Date         string   `json:"date"`
	Favorite     Favorite `json:"favorite"`

	Series bool `json:"series"`
}

type Tag struct {
	Name  string `json:"name"`
	Count uint   `json:"count"`
}

func exportSubjects(q *query.Query, w io.Writer) {
	for i := model.SubjectID(0); i < maxSubjectID; i += defaultStep {
		subjects, err := q.WithContext(ctx).Subject.Preload(q.Subject.Fields).
			Where(q.Subject.ID.Gt(i), q.Subject.ID.Lte(i+defaultStep), q.Subject.Ban.Eq(0)).Find()
		if err != nil {
			panic(err)
		}

		for _, subject := range subjects {
			tags, err := subjectDto.ParseTags(subject.Fields.Tags)
			if err != nil {
				tags = []model.Tag{}
			}

			sort.Slice(tags, func(i, j int) bool {
				return tags[i].Count >= tags[j].Count
			})

			tags = lo.Filter(lo.Slice(tags, 0, 11), func(item model.Tag, index int) bool { //nolint:mnd
				return utf8.RuneCountInString(item.Name) < 10 || item.Count >= 10
			})

			encodedTags := lo.Map(tags, func(item model.Tag, index int) Tag {
				return Tag{
					Name:  item.Name,
					Count: item.Count,
				}
			})

			f := subject.Fields
			var total = f.Rate1 + f.Rate2 + f.Rate3 + f.Rate4 + f.Rate5 + f.Rate6 + f.Rate7 + f.Rate8 + f.Rate9 + f.Rate10
			var score float64
			if total != 0 {
				score = float64(1*f.Rate1+2*f.Rate2+3*f.Rate3+4*f.Rate4+5*f.Rate5+
					6*f.Rate6+7*f.Rate7+8*f.Rate8+9*f.Rate9+10*f.Rate10) / float64(total)
			}

			encodedDate := ""
			if !subject.Fields.Date.IsZero() {
				encodedDate = subject.Fields.Date.Format("2006-01-02")
			}

			var metaTags = []string{}

			for _, v := range strings.Split(subject.FieldMetaTags, " ") {
				v = strings.TrimSpace(v)
				if v == "" {
					continue
				}

				metaTags = append(metaTags, v)
			}

			encode(w, Subject{
				ID:       subject.ID,
				Type:     subject.TypeID,
				Name:     string(subject.Name),
				NameCN:   string(subject.NameCN),
				Infobox:  string(subject.Infobox),
				Platform: subject.Platform,
				Summary:  subject.Summary,
				Nsfw:     subject.Nsfw,
				Rank:     subject.Fields.Rank,
				Tags:     encodedTags,
				MetaTags: metaTags,
				Score:    math.Round(score*10) / 10,
				ScoreDetails: Score{
					Field1:  subject.Fields.Rate1,
					Field2:  subject.Fields.Rate2,
					Field3:  subject.Fields.Rate3,
					Field4:  subject.Fields.Rate4,
					Field5:  subject.Fields.Rate5,
					Field6:  subject.Fields.Rate6,
					Field7:  subject.Fields.Rate7,
					Field8:  subject.Fields.Rate8,
					Field9:  subject.Fields.Rate9,
					Field10: subject.Fields.Rate10,
				},
				Date:   encodedDate,
				Series: subject.Series,
				Favorite: Favorite{
					Wish:    subject.Wish,
					Done:    subject.Done,
					Doing:   subject.Doing,
					OnHold:  subject.OnHold,
					Dropped: subject.Dropped,
				},
			})
		}
	}
}

type Person struct {
	ID       model.PersonID `json:"id"`
	Name     string         `json:"name"`
	Type     uint8          `json:"type"`
	Career   []string       `json:"career"`
	Infobox  string         `json:"infobox"`
	Summary  string         `json:"summary"`
	Comments uint32         `json:"comments"`
	Collects uint32         `json:"collects"`
}

func exportPersons(q *query.Query, w io.Writer) {
	for i := model.PersonID(0); i < maxPersonID; i += defaultStep {
		persons, err := q.WithContext(context.Background()).Person.
			Where(q.Person.ID.Gt(i), q.Person.ID.Lte(i+defaultStep)).Find()
		if err != nil {
			panic(err)
		}

		for _, p := range persons {
			encode(w, Person{
				ID:       p.ID,
				Name:     p.Name,
				Type:     p.Type,
				Career:   careers(p),
				Infobox:  p.Infobox,
				Summary:  p.Summary,
				Comments: p.Comment,
				Collects: p.Collects,
			})
		}
	}
}

func careers(p *dao.Person) []string {
	s := make([]string, 0, 7)

	if p.Writer {
		s = append(s, "writer")
	}

	if p.Producer {
		s = append(s, "producer")
	}

	if p.Mangaka {
		s = append(s, "mangaka")
	}

	if p.Artist {
		s = append(s, "artist")
	}

	if p.Seiyu {
		s = append(s, "seiyu")
	}

	if p.Illustrator {
		s = append(s, "illustrator")
	}

	if p.Actor {
		s = append(s, "actor")
	}

	return s
}

type Character struct {
	ID       model.CharacterID `json:"id"`
	Role     uint8             `json:"role"`
	Name     string            `json:"name"`
	Infobox  string            `json:"infobox"`
	Summary  string            `json:"summary"`
	Comments uint32            `json:"comments"`
	Collects uint32            `json:"collects"`
}

func exportCharacters(q *query.Query, w io.Writer) {
	for i := model.CharacterID(0); i < maxCharacterID; i += defaultStep {
		characters, err := q.WithContext(context.Background()).Character.
			Where(q.Character.ID.Gt(i), q.Character.ID.Lte(i+defaultStep)).Find()
		if err != nil {
			panic(err)
		}

		for _, c := range characters {
			encode(w, Character{
				ID:       c.ID,
				Name:     c.Name,
				Role:     c.Role,
				Infobox:  c.Infobox,
				Summary:  c.Summary,
				Comments: c.Comment,
				Collects: c.Collects,
			})
		}
	}
}

type Episode struct {
	ID          model.EpisodeID `json:"id"`
	Name        string          `json:"name"`
	NameCn      string          `json:"name_cn"`
	Description string          `json:"description"`
	AirDate     string          `json:"airdate"`
	Disc        uint8           `json:"disc"`
	Duration    string          `json:"duration"`
	SubjectID   model.SubjectID `json:"subject_id"`
	Sort        float32         `json:"sort"`
	Type        episode.Type    `json:"type"`
}

func exportEpisodes(q *query.Query, w io.Writer) {
	lastEpisode, err := q.WithContext(ctx).Episode.Order(q.Episode.ID.Desc()).Take()
	if err != nil {
		panic(err)
	}
	for i := model.EpisodeID(0); i < lastEpisode.ID; i += defaultStep {
		episodes, err := q.WithContext(context.Background()).Episode.
			Where(q.Episode.ID.Gt(i), q.Episode.ID.Lte(i+defaultStep), q.Episode.Ban.Eq(0)).Find()
		if err != nil {
			panic(err)
		}

		for _, e := range episodes {
			encode(w, Episode{
				ID:          e.ID,
				Name:        e.Name,
				NameCn:      e.NameCn,
				Sort:        e.Sort,
				SubjectID:   e.SubjectID,
				Duration:    e.Duration,
				Description: e.Desc,
				Type:        e.Type,
				AirDate:     e.Airdate,
				Disc:        e.Disc,
			})
		}
	}
}

type SubjectRelation struct {
	SubjectID        model.SubjectID `json:"subject_id"`
	RelationType     uint16          `json:"relation_type"`
	RelatedSubjectID model.SubjectID `json:"related_subject_id"`
	Order            uint16          `json:"order"`
}

func exportSubjectRelations(q *query.Query, w io.Writer) {
	for i := model.SubjectID(0); i < maxSubjectID; i += defaultStep {
		relations, err := q.WithContext(context.Background()).SubjectRelation.
			Order(q.SubjectRelation.SubjectID, q.SubjectRelation.SubjectID).
			Where(q.SubjectRelation.SubjectID.Gt(i), q.SubjectRelation.SubjectID.Lte(i+defaultStep)).Find()
		if err != nil {
			panic(err)
		}

		for _, rel := range relations {
			encode(w, SubjectRelation{
				SubjectID:        rel.SubjectID,
				RelationType:     rel.RelationType,
				RelatedSubjectID: rel.RelatedSubjectID,
				Order:            rel.Order,
			})
		}
	}
}

type SubjectPerson struct {
	PersonID  model.PersonID  `json:"person_id"`
	SubjectID model.SubjectID `json:"subject_id"`
	Position  uint16          `json:"position"`
	AppearEps string          `json:"appear_eps"`
}

func exportSubjectPersonRelations(q *query.Query, w io.Writer) {
	for i := model.SubjectID(0); i < maxSubjectID; i += defaultStep {
		relations, err := q.WithContext(context.Background()).PersonSubjects.
			Order(q.PersonSubjects.SubjectID, q.PersonSubjects.PersonID).
			Where(q.PersonSubjects.SubjectID.Gt(i), q.PersonSubjects.SubjectID.Lte(i+defaultStep)).Find()
		if err != nil {
			panic(err)
		}

		for _, rel := range relations {
			encode(w, SubjectPerson{
				PersonID:  rel.PersonID,
				SubjectID: rel.SubjectID,
				Position:  rel.PrsnPosition,
				AppearEps: rel.PrsnAppearEps,
			})
		}
	}
}

type SubjectCharacter struct {
	CharacterID model.CharacterID `json:"character_id"`
	SubjectID   model.SubjectID   `json:"subject_id"`
	Type        uint8             `json:"type"`
	Order       uint16            `json:"order"`
}

func exportSubjectCharacterRelations(q *query.Query, w io.Writer) {
	for i := model.SubjectID(0); i < maxSubjectID; i += defaultStep {
		relations, err := q.WithContext(context.Background()).CharacterSubjects.
			Order(q.CharacterSubjects.SubjectID, q.CharacterSubjects.CrtOrder).
			Where(q.CharacterSubjects.SubjectID.Gt(i), q.CharacterSubjects.SubjectID.Lte(i+defaultStep)).Find()
		if err != nil {
			panic(err)
		}

		for _, rel := range relations {
			encode(w, SubjectCharacter{
				CharacterID: rel.CharacterID,
				SubjectID:   rel.SubjectID,
				Type:        rel.CrtType,
				Order:       rel.CrtOrder,
			})
		}
	}
}

type PersonCharacter struct {
	PersonID    model.PersonID    `json:"person_id"`
	SubjectID   model.SubjectID   `json:"subject_id"`
	CharacterID model.CharacterID `json:"character_id"`
	Summary     string            `json:"summary"`
}

func exportPersonCharacterRelations(q *query.Query, w io.Writer) {
	for i := model.PersonID(0); i < maxPersonID; i += defaultStep {
		relations, err := q.WithContext(context.Background()).Cast.
			Order(q.Cast.PersonID, q.Cast.CharacterID).
			Where(q.Cast.PersonID.Gt(i), q.Cast.PersonID.Lte(i+defaultStep)).Find()
		if err != nil {
			panic(err)
		}

		for _, rel := range relations {
			encode(w, PersonCharacter{
				PersonID:    rel.PersonID,
				SubjectID:   rel.SubjectID,
				CharacterID: rel.CharacterID,
				Summary:     rel.Summary,
			})
		}
	}
}

type PersonRelation struct {
	PersonType      string         `json:"person_type"`
	PersonID        model.PersonID `json:"person_id"`
	RelatedPersonID model.PersonID `json:"related_person_id"`
	RelationType    uint32         `json:"relation_type"`
	Spoiler         bool           `json:"spoiler"`
	Ended           bool           `json:"ended"`
}

func exportPersonRelations(q *query.Query, w io.Writer) {
	for i := model.PersonID(0); i < maxPersonID; i += defaultStep {
		relations, err := q.WithContext(context.Background()).PersonRelation.
			Order(q.PersonRelation.PersonID, q.PersonRelation.PersonID).
			Where(q.PersonRelation.PersonID.Gt(i), q.PersonRelation.PersonID.Lte(i+defaultStep)).Find()
		if err != nil {
			panic(err)
		}

		for _, rel := range relations {
			encode(w, PersonRelation{
				PersonType:      rel.PersonType,
				PersonID:        rel.PersonID,
				RelatedPersonID: rel.RelatedPersonID,
				RelationType:    rel.RelationType,
				Spoiler:         rel.Spoiler,
				Ended:           rel.Ended,
			})
		}
	}
}

func encode(w io.Writer, object any) {
	if err := json.NewEncoder(w).Encode(object); err != nil {
		panic(err)
	}
}
