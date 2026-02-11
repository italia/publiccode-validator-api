package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type Status struct{}

func NewStatus() *Status {
	return &Status{}
}

// GetStatus gets status of the API.
func (s *Status) GetStatus(ctx *fiber.Ctx) error {
	ctx.Append("Cache-Control", "no-cache")

	return ctx.SendStatus(fiber.StatusNoContent)
}
