package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

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

	client, err := getEntClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	log.Panicln(startHTTP(client))
}

func getEntClient() (*ent.Client, error) {
	host := getEnv("MYSQL_HOST", "127.0.0.1")
	port := getEnv("MYSQL_PORT", "3306")
	user := mustGetEnv("MYSQL_USER")
	pass := mustGetEnv("MYSQL_PASS")
	db := getEnv("MYSQL_DB", "bangumi")

	debug, _ := strconv.ParseBool(getEnv("DB_DEBUG", "false"))

	var options = make([]ent.Option, 0, 0)

	if debug {
		options = append(options, ent.Debug())
	}

	return ent.Open(
		"mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, pass, host, port, db),
		options...,
	)
}

func getEnv(n, v string) string {
	if e, ok := os.LookupEnv(n); ok {
		return e
	}

	return v
}

func mustGetEnv(n string) string {
	if e, ok := os.LookupEnv(n); ok {
		return e
	}

	panic("you need to set env " + strconv.Quote(n))
}
