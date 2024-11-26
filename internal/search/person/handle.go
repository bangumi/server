package person

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

//nolint:funlen
func (c *client) Handle(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "TODO:")
}
