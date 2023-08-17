package main

import "github.com/gofiber/fiber/v2"

func KeepAlive(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	return c.Next()
}
