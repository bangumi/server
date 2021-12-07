package main

import (
	"encoding/json"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"

	"app/ent"
	"app/pkg/logger"
	"app/pkg/web"
)

func startHTTP(mysql *ent.Client) error {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		StrictRouting:         true,
		CaseSensitive:         true,
		GETOnly:               false,
		JSONEncoder: func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "  ") //nolint:wrapcheck
		},
	})

	app.Get("/", func(ctx *fiber.Ctx) error {
		logger.Info("test")

		return ctx.SendString("test")
	})

	web.SetupRouter(app, mysql)
	logger.Info("start serer")

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON("not found router")
	})

	return errors.Wrap(app.Listen(":80"), "failed to start http server")
}

func main() {
	if err := logger.Setup(); err != nil {
		log.Fatalln(err)
	}

	client, err := ent.Open("mysql", "userrr:passwdd@tcp(192.168.1.3:3308)/chii", ent.Debug())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	log.Panicln(startHTTP(client))
}
