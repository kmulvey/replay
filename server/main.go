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
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/two", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 2)
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/three", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 3)
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/four", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 4)
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/five", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 5)
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/six", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 6)
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/seven", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 7)
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/eight", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 8)
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/nine", func(c fiber.Ctx) error {
		time.Sleep(time.Millisecond * 9)
		return c.SendStatus(fiber.StatusOK)
	})

	log.Fatal(app.Listen(":7000", fiber.ListenConfig{
		EnablePrefork: true,
	}))
}
