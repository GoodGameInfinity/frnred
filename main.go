package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"flag"
	"fmt"

	"os"
	"strings"

	"github.com/dchest/uniuri"

	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/keyauth"
	"github.com/jxskiss/base62"

	"go.frnsrv.ru/frnred/query"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type addUrl struct {
	Url string
}

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

	adder := app.Group("/add")
	adder.Use(keyauth.New(keyauth.Config{
		KeyLookup: "cookie:access_token",
		Validator: func(c fiber.Ctx, key string) (bool, error) {
			hashedKey := sha256.Sum256([]byte(key))

			if _, err := queries.FindKey(ctx, string(hashedKey[:])); err != nil {
				if err == sql.ErrNoRows {
					return false, keyauth.ErrMissingOrMalformedAPIKey
				} else {
					return false, c.SendStatus(fiber.StatusInternalServerError)
				}
			}
			return true, nil
		},
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().To(rootURL)
	})

	// Standart shortener behaviour
	// l stands for link
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

	adder.Post("/", func(c fiber.Ctx) error {
		u := query.CreateUrlParams{}

		// Generating a new ID
		var try string
		for {
			try = uniuri.NewLen(8)
			if _, err := queries.GetUrl(ctx, try); err == sql.ErrNoRows { // Check if it exists **just in case**
				u.ID = try
				break
			} else if err != nil {
				log.Fatal(err.Error())
			}
		}

		// Appending the specified URL
		var nu addUrl
		log.Print(c.Bind(), c.Bind().JSON(&nu))
		c.Bind().JSON(&nu)
		u.Url = nu.Url

		url, err := queries.CreateUrl(ctx, u)
		if err != nil {
			log.Fatal(err.Error())
		}

		return c.JSON(url)
	})

	// Vanity URLs (always works no matter what)
	// n stands for name
	app.Get("/v/:n", func(c fiber.Ctx) error {
		v, err := queries.GetVanity(ctx, c.Params("n"))
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

	adder.Post("/v", func(c fiber.Ctx) error {
		u := query.CreateVanityParams{}

		log.Print(c.Bind(), c.Bind().JSON(&u))
		c.Bind().JSON(&u)

		url, err := queries.CreateVanity(ctx, u)
		if err != nil {
			log.Fatal(err.Error())
		}

		return c.JSON(url)
	})

	// Algorithmic URLs
	// c stands for code
	app.Get("/a/:c", func(c fiber.Ctx) error {
		url, err := base62.Decode([]byte(c.Params("c")))
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		return c.Redirect().Status(fiber.StatusPermanentRedirect).To("https://" + string(url))
	})

	// HTTP edition for the whoever actually needs it
	app.Get("/at/:c", func(c fiber.Ctx) error {
		url, err := base62.Decode([]byte(c.Params("c")))
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		return c.Redirect().Status(fiber.StatusPermanentRedirect).To("http://" + string(url))
	})

	app.Hooks().OnListen(func(listenData fiber.ListenData) error {
		log.Println("Server is up and running!")
		log.Println("Try " + "http://" + listenData.Host + ":" + listenData.Port + "/a/" + base62.EncodeToString([]byte("example.com")))
		// Create a new admin key if there are no keys
		if _, err := queries.CheckKey(ctx); err != nil {
			if err == sql.ErrNoRows {
				key := uniuri.NewLen(16)
				hash := sha256.Sum256([]byte(key))
				queries.CreateKey(ctx, query.CreateKeyParams{ID: key, Hashed: string(hash[:]), Admin: sql.NullBool{Bool: true, Valid: true}})
				log.Println("Created a new admin API key: " + key)
			}
		}
		return nil
	})

	if err := app.Listen(appAddr, fiber.ListenConfig{EnablePrefork: true}); err != nil {
		log.Fatal("Webapp stopped with error: ", err.Error(), "\nExiting!")
	}
}
