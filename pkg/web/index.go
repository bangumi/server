package web

import (
	"github.com/gofiber/fiber/v2"

	"app/ent"
	v0 "app/pkg/web/v0"
)

func SetupRouter(router fiber.Router, mysql *ent.Client) {
	v0.SetupRouter(router.Group("/v0"), mysql)
}
