package v0

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"

	"app/ent"
	"app/ent/characterfields"
	"app/ent/personcsindex"
	"app/pkg/logger"
	"app/pkg/wiki"
)

func SetupRouter(router fiber.Router, mysql *ent.Client) {
	router.Use(func(c *fiber.Ctx) error {
		logger.Info(c.Path())

		err := c.Next()
		if err != nil {
			if errors.Is(err, fiber.ErrNotFound) {
				return c.JSON(Error{Error: "not found", Message: "can't find the resource you queried"})
			}

			return c.Status(fiber.StatusInternalServerError).
				JSON(Error{Error: "unexpected error", Message: err.Error()})
		}

		return nil
	})

	router.Get("/characterfields/:prsn_id", characterfieldsGetter(characterfields.PrsnCatCrt, mysql))
	router.Get("/person/:prsn_id", prsnGetter(characterfields.PrsnCatPrsn, mysql))
}

func characterfieldsGetter(t characterfields.PrsnCat, mysql *ent.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := ctxGetPositiveInt(c, "prsn_id")
		if err != nil {
			return err
		}

		prsn, err := mysql.CharacterFields.Query().
			Where(characterfields.PrsnCatEQ(t), characterfields.IDEQ(uint8(id))).
			First(c.Context())

		if err != nil {
			if ent.IsNotFound(err) {
				return fiber.ErrNotFound
			}

			return errors.Wrap(err, "failed to execute sql")
		}

		return c.JSON(prsn)
	}
}

func prsnGetter(t characterfields.PrsnCat, mysql *ent.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := ctxGetPositiveInt(c, "prsn_id")
		if err != nil {
			return err
		}

		p, err := mysql.Person.Get(c.Context(), id)
		if err != nil {
			return errNotFoundTo404(err)
		}

		if p.Redirect != 0 {
			return c.Redirect(strconv.Itoa(p.Redirect))
		}

		prsn, err := mysql.CharacterFields.Query().
			Where(characterfields.PrsnCatEQ(t), characterfields.IDEQ(uint8(id))).
			First(c.Context())

		if err != nil {
			return errNotFoundTo404(err)
		}

		cs, err := mysql.PersonCsIndex.Query().Select(personcsindex.FieldSubjectID).Unique(true).
			Where(personcsindex.PrsnTypeEQ(personcsindex.PrsnType(t)), personcsindex.ID(prsn.ID)).
			Order(ent.Asc(personcsindex.FieldSubjectID)).All(c.Context())
		if err != nil {
			return errors.Wrap(err, "sql error")
		}

		pp := PersonDTO{
			ID:         prsn.ID,
			URL:        fmt.Sprintf("https://bgm.tv/person/%d", prsn.ID),
			SubjectIDs: make([]int, len(cs)),
			Summary:    p.Summary,
			Img:        p.Img,
			Fields:     prsn,
		}

		for i, c := range cs {
			pp.SubjectIDs[i] = c.SubjectID
		}

		w, err := wiki.Parse(p.Infobox)
		if err == nil {
			pp.Wiki = w
		}

		return c.JSON(pp)
	}
}

type PersonDTO struct {
	Wiki       wiki.Wiki            `json:"wiki"`
	URL        string               `json:"#url"`
	Summary    string               `json:"summary"`
	Img        string               `json:"img"`
	Fields     *ent.CharacterFields `json:"parsed"`
	SubjectIDs []int                `json:"subject_ids"`
	ID         uint8                `json:"id"`
}
