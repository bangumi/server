package v0

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"

	"app/ent"
)

func ctxGetPositiveInt(c *fiber.Ctx, name string) (int, error) {
	v := c.Params(name)
	if v == "" {
		return 0, ErrMissingParam
	}
	id, err := strconv.Atoi(v)
	if err != nil || id <= 0 {
		return 0, errors.Wrapf(ErrQuery, "%s is not positive integer", v)
	}

	return id, nil
}

func errNotFoundTo404(err error) error {
	if ent.IsNotFound(err) {
		return fiber.ErrNotFound
	}

	return errors.Wrap(err, "failed to execute sql")
}
