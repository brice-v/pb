package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pb/db"
	"pb/handlers"
	"pb/sql"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
)

//go:embed views/*
var viewsStaticDir embed.FS

//go:embed public/*
var publicStaticDir embed.FS

type PB struct {
	app *fiber.App
}

func NewPB() *PB {
	engine := html.NewFileSystem(http.FS(viewsStaticDir), ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// use embedded public directory
	app.Use("/public", filesystem.New(filesystem.Config{
		Root:       http.FS(publicStaticDir),
		PathPrefix: "public",
	}))
	db.DB.Exec(sql.CreatePasteTable)
	db.DB.Exec(sql.InsertPasteTable, -1, "seed", "seed")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("views/home", "")
	})

	app.Get("/paste/:id", handlers.GetPaste)

	app.Post("/paste", handlers.PostPaste)
	app.Post("/paste-ui", handlers.PostPasteUI)
	app.Get("/paste-ui/:id", handlers.GetPasteUI)
	return &PB{app: app}
}

func (pb *PB) Run() {
	if pb.app == nil {
		log.Fatal("PB error: *fiber.App is nil")
	}
	go func() {
		if err := pb.app.Listen(":3001"); err != nil {
			log.Panic("error while listening: " + err.Error())
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	fmt.Println("gracefully shutting down...")
	if err := pb.app.Shutdown(); err != nil {
		log.Printf("FAILED to shutdown app, error: %s", err.Error())
	}

	fmt.Println("running cleanup tasks...")
	if err := db.DB.Close(); err != nil {
		log.Printf("FAILED to close DB, error: %s", err.Error())
	}
	fmt.Println("pb shutdown")
}

func main() {
	pb := NewPB()
	pb.Run()
}
