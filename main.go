package main

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"
)

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
	app := fiber.New()
	db.Exec(createPasteTableSql)
	db.Exec(insertPasteTableSql, -1, "seed", "seed")

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
	app.Post("/paste", func(c *fiber.Ctx) error {
		p := new(Paste)
		if err := c.BodyParser(p); err != nil {
			log.Printf("POST /paste error: %s", err.Error())
			return err
		}
		if p.Title == "" {
			log.Printf("POST /paste error: title is empty")
			return c.Status(fiber.StatusBadRequest).JSON("title is empty")
		}
		if p.Text == "" {
			log.Printf("POST /paste error: text is empty")
			return c.Status(fiber.StatusBadRequest).JSON("text is empty")
		}
		row := db.QueryRow(getMaxPasteIdSql)
		var curId int
		err := row.Scan(&curId)
		if err != nil {
			log.Printf("POST /paste error: %s", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON("failed to get id")
		}
		// Increment for new paste
		curId++
		_, err = db.Exec(insertPasteTableSql, curId, p.Title, p.Text)
		if err != nil {
			log.Printf("POST /paste error: %s", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON("failed to insert paste")
		}
		return c.JSON(curId)
	})

	app.Listen(":3001")
}
