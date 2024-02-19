package handlers

import (
	"fmt"
	"log"
	"pb/db"
	"pb/models"
	"pb/sql"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetPaste(c *fiber.Ctx) error {
	sid := c.Params("id", "")
	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		log.Printf("GET /paste/:id (%s) error: %s", sid, err.Error())
		return c.Status(fiber.StatusBadRequest).JSON("invalid id " + sid)
	}
	var title, text string
	row := db.DB.QueryRow(sql.GetPaste, id)
	err = row.Scan(&title, &text)
	if err != nil {
		log.Printf("GET /paste/:id (%s) error: %s", sid, err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON("failed to find paste with id " + sid)
	}
	return c.JSON(models.Paste{Title: title, Text: text})
}

func pasteHandler(c *fiber.Ctx) (int, error) {
	p := new(models.Paste)
	if err := c.BodyParser(p); err != nil {
		log.Printf("POST /paste error: %s", err.Error())
		return -1, err
	}
	err := p.Validate()
	if err != nil {
		return -1, c.Status(fiber.StatusBadRequest).JSON("POST /paste error: " + err.Error())
	}
	row := db.DB.QueryRow(sql.GetMaxPasteId)
	var curId int
	err = row.Scan(&curId)
	if err != nil {
		log.Printf("POST /paste error: %s", err.Error())
		return -1, c.Status(fiber.StatusInternalServerError).JSON("failed to get id")
	}
	// Increment for new paste
	curId++
	_, err = db.DB.Exec(sql.InsertPasteTable, curId, p.Title, p.Text)
	if err != nil {
		log.Printf("POST /paste error: %s", err.Error())
		return -1, c.Status(fiber.StatusInternalServerError).JSON("failed to insert paste")
	}
	return curId, c.JSON(curId)
}

func PostPaste(c *fiber.Ctx) error {
	_, err := pasteHandler(c)
	return err
}

func PostPasteUI(c *fiber.Ctx) error {
	id, _ := pasteHandler(c)
	if id == -1 {
		// Return the error on the ui
		return c.Render("views/home", fiber.Map{"Error": "Some error happened."})
	}
	var title, text string
	row := db.DB.QueryRow(sql.GetPaste, id)
	err := row.Scan(&title, &text)
	if err != nil {
		return c.Render("views/home", fiber.Map{"Error": "Failed to get paste."})
	}
	height := fmt.Sprintf("%dem", strings.Count(text, "\n"))
	return c.Render("views/paste", fiber.Map{"Title": title, "Text": text, "Id": id, "Height": height})

}

func GetPasteUI(c *fiber.Ctx) error {
	sid := c.Params("id", "")
	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		log.Printf("GET /paste-ui/:id (%s) error: %s", sid, err.Error())
		return c.Render("views/home", fiber.Map{"Error": "Failed to get paste."})
	}
	var title, text string
	row := db.DB.QueryRow(sql.GetPaste, id)
	err = row.Scan(&title, &text)
	if err != nil {
		log.Printf("GET /paste-ui/:id (%s) error: %s", sid, err.Error())
		return c.Render("views/home", fiber.Map{"Error": "Failed to get paste."})
	}
	height := fmt.Sprintf("%dem", strings.Count(text, "\n"))
	return c.Render("views/paste", fiber.Map{"Title": title, "Text": text, "Id": id, "Height": height})
}
