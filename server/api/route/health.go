package route

import "github.com/gofiber/fiber/v2"

func Health(c *fiber.Ctx) error {
	c.SendString("OK")
	return nil
}