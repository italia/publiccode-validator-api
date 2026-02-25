package handlers

import (
	"github.com/gofiber/fiber/v3"
)

type Status struct {
	Version string `json:"v"`
	Commit  string `json:"commit"`
}

func NewStatus(version string, commit string) *Status {
	return &Status{Version: version, Commit: commit}
}

// GetStatus gets status of the API.
func (s *Status) GetStatus(ctx fiber.Ctx) error {
	ctx.Append("Cache-Control", "no-cache")

	//nolint:wrapcheck
	return ctx.Status(fiber.StatusOK).JSON(s)
}
