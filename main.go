package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"os"
	"strings"

	"log"

	"github.com/gofiber/fiber/v3"

	"go.frnsrv.ru/frnred/query"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var ctx = context.Background()

// Not so complicated, right?
func main() {
	var dbUrl string
	flag.StringVar(&dbUrl, "db", "file:./frnred.db", "Database connection string (file: or libsql:// for libsql | postgres:// | | sql:// etc.)")
	// Currently unused
	//var analiticUrl string
	//flag.StringVar(&dbUrl, "adb", "", "Analytics database connection string (clickhouse://). Adding such a connection string also enables analytics.")

	var appAddr string
	flag.StringVar(&appAddr, "addr", "0.0.0.0:8080", "Application address string (0.0.0.0:8080)")

	var rootURL string
	flag.StringVar(&rootURL, "root", "https://friendsserver.ru", "Root URL redirect (https://friendsserver.ru)")

	// Currently unused
	//var webEnable bool
	//flag.BoolVar(&webEnable, "web", false, "Enable WebUI (true/false)")

	var help bool
	flag.BoolVar(&help, "help", false, "Shows this message")

	flag.Parse()

	if help {
		fmt.Println("Usage: frnred [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("\nThis program is a URL shortener designed for frnsrv.ru.")
		fmt.Println("Licensed under the MIT-0, copyright 2025 GoodGameInfinity.")
		os.Exit(0)
	}

	// Database
	var db *sql.DB

	if strings.HasPrefix(dbUrl, "host=") || strings.HasPrefix(dbUrl, "postgres://") || strings.HasPrefix(dbUrl, "postgresql://") { // Postgres
		var err error
		db, err = sql.Open("postgres", dbUrl)
		if err != nil {
			log.Fatal("failed to connect to Postgres database: ", err.Error())
		}
	} else if strings.HasPrefix(dbUrl, "libsql://") || strings.HasPrefix(dbUrl, "file:") { // libSQL
		var err error
		db, err = sql.Open("libsql", dbUrl)
		if err != nil {
			log.Fatal("failed to connect to libSQL database: ", err.Error())
		}
	} else if strings.HasPrefix(dbUrl, "sql://") { // MySQL or other compatable systems
		var err error
		db, err = sql.Open("mysql", dbUrl)
		if err != nil {
			log.Fatal("failed to connect to MySQL-compatable database: ", err.Error())
		}
	} else {
		log.Fatal("bad database connection string: ", dbUrl, `

		All available connection strings: 
			Postgres-compatable | prefix: host= OR prefix: postgres:// OR prefix: postgresql://
			libSQL              | prefix: file: OR prefix: libsql://
			MySQL-compatable    | prefix: sql://
		`)
	}
	queries := query.New(db)

	// Redirect + API methods
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().To(rootURL)
	})

	// Standart shortener behaviour
	app.Get("/:l", func(c fiber.Ctx) error {
		lp := c.Params("l")
		l, err := queries.GetUrl(ctx, lp)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				{
					// Try a vanity URL
					v, errr := queries.GetVanity(ctx, lp)
					if errr != nil {
						switch errr {
						case sql.ErrNoRows:
							{
								return c.SendStatus(fiber.StatusNotFound)
							}
						default:
							{
								return c.SendStatus(fiber.StatusInternalServerError)
							}
						}
					}
					return c.Redirect().To(v.Url)
				}
			default:
				{
					return c.SendStatus(fiber.StatusInternalServerError)
				}
			}
		}
		return c.Redirect().To(l.Url)
	})

	// Vanity URLs (always works no matter what)
	app.Get("/v/:name", func(c fiber.Ctx) error {
		v, err := queries.GetVanity(ctx, c.Params("name"))
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				{
					return c.SendStatus(fiber.StatusNotFound)
				}
			default:
				{
					return c.SendStatus(fiber.StatusInternalServerError)
				}
			}
		}
		return c.Redirect().To(v.Url)
	})

	app.Hooks().OnListen(func(listenData fiber.ListenData) error {
		log.Println("Server is up and running!")
		return nil
	})

	if err := app.Listen(appAddr, fiber.ListenConfig{EnablePrefork: true}); err != nil {
		log.Fatal("Webapp stopped with error: ", err.Error(), "\nExiting!")
	}
}
