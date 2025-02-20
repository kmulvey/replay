package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	app.Get("/one", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond)
		return c.SendString("one")
	})

	app.Get("/two", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 2)
		return c.SendString("two")
	})

	app.Get("/three", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 3)
		return c.SendString("three")
	})

	app.Get("/four", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 4)
		return c.SendString("four")
	})

	app.Get("/five", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 5)
		return c.SendString("five")
	})

	app.Get("/six", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 6)
		return c.SendString("six")
	})

	app.Get("/seven", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 7)
		return c.SendString("seven")
	})

	app.Get("/eight", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 8)
		return c.SendString("eight")
	})

	app.Get("/nine", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 9)
		return c.SendString("nine")
	})

	log.Fatal(app.Listen(":7000", fiber.ListenConfig{
		EnablePrefork: true,
	}))
}
