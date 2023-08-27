package main

import (
	"embed"
	"log"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"
)

//go:embed views/*
var viewsStaticDir embed.FS

//go:embed public/*
var publicStaticDir embed.FS

type Paste struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

const createPasteTableSql = `CREATE TABLE IF NOT EXISTS pastes(id INTEGER PRIMARY KEY, title TEXT, paste_text TEXT NOT NULL);`
const getMaxPasteIdSql = `SELECT MAX(id) FROM pastes;`
const insertPasteTableSql = `INSERT INTO pastes (id, title, paste_text) VALUES (?, ?, ?);`
const getPasteSql = `SELECT title, paste_text FROM pastes WHERE id = ?;`

func main() {
	db := sqlx.MustConnect("sqlite", "pb.db?_pragma=journal_mode(WAL)")

	engine := html.NewFileSystem(http.FS(viewsStaticDir), ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// use embedded public directory
	app.Use("/public", filesystem.New(filesystem.Config{
		Root:       http.FS(publicStaticDir),
		PathPrefix: "public",
	}))
	db.Exec(createPasteTableSql)
	db.Exec(insertPasteTableSql, -1, "seed", "seed")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("views/home", "")
	})

	app.Get("/paste/:id", func(c *fiber.Ctx) error {
		sid := c.Params("id", "")
		id, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			log.Printf("GET /paste/:id (%s) error: %s", sid, err.Error())
			return c.Status(fiber.StatusBadRequest).JSON("invalid id " + sid)
		}
		var title, text string
		row := db.QueryRow(getPasteSql, id)
		err = row.Scan(&title, &text)
		if err != nil {
			log.Printf("GET /paste/:id (%s) error: %s", sid, err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON("failed to find paste with id " + sid)
		}
		return c.JSON(Paste{Title: title, Text: text})
	})
	pasteHandler := func(c *fiber.Ctx) (error, int) {
		p := new(Paste)
		if err := c.BodyParser(p); err != nil {
			log.Printf("POST /paste error: %s", err.Error())
			return err, -1
		}
		if p.Title == "" {
			log.Printf("POST /paste error: title is empty")
			return c.Status(fiber.StatusBadRequest).JSON("title is empty"), -1
		}
		if p.Text == "" {
			log.Printf("POST /paste error: text is empty")
			return c.Status(fiber.StatusBadRequest).JSON("text is empty"), -1
		}
		row := db.QueryRow(getMaxPasteIdSql)
		var curId int
		err := row.Scan(&curId)
		if err != nil {
			log.Printf("POST /paste error: %s", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON("failed to get id"), -1
		}
		// Increment for new paste
		curId++
		_, err = db.Exec(insertPasteTableSql, curId, p.Title, p.Text)
		if err != nil {
			log.Printf("POST /paste error: %s", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON("failed to insert paste"), -1
		}
		return c.JSON(curId), curId
	}
	app.Post("/paste", func(c *fiber.Ctx) error {
		err, _ := pasteHandler(c)
		return err
	})
	app.Post("/paste-ui", func(c *fiber.Ctx) error {
		_, id := pasteHandler(c)
		if id == -1 {
			// Return the error on the ui
			return c.Render("views/home", fiber.Map{"Error": "Some error happened."})
		}
		var title, text string
		row := db.QueryRow(getPasteSql, id)
		err := row.Scan(&title, &text)
		if err != nil {
			return c.Render("views/home", fiber.Map{"Error": "Failed to get paste."})
		}
		return c.Render("views/paste", fiber.Map{"Title": title, "Text": text, "Id": id})
	})
	app.Get("/paste-ui/:id", func(c *fiber.Ctx) error {
		sid := c.Params("id", "")
		id, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			log.Printf("GET /paste-ui/:id (%s) error: %s", sid, err.Error())
			return c.Render("views/home", fiber.Map{"Error": "Failed to get paste."})
		}
		var title, text string
		row := db.QueryRow(getPasteSql, id)
		err = row.Scan(&title, &text)
		if err != nil {
			log.Printf("GET /paste-ui/:id (%s) error: %s", sid, err.Error())
			return c.Render("views/home", fiber.Map{"Error": "Failed to get paste."})
		}
		return c.Render("views/paste", fiber.Map{"Title": title, "Text": text, "Id": id})
	})

	app.Listen(":3001")
}
