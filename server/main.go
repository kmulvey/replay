package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	app.Get("/slow", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 100)
		//fmt.Println("slow")
		return c.SendString("slow")
	})

	app.Get("/medium", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 50)
		//fmt.Println("medium")
		return c.SendString("medium")
	})

	app.Get("/fast", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond)
		///fmt.Println("fast")
		return c.SendString("fast")
	})

	log.Fatal(app.Listen(":7000", fiber.ListenConfig{
		EnablePrefork: true,
	}))
}
