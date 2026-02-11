package common

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

func InternalServerError(title string) ProblemJSONError {
	return Error(fiber.StatusInternalServerError, title, fiber.ErrInternalServerError.Message)
}

func Error(status int, title string, detail string) ProblemJSONError {
	return ProblemJSONError{Title: title, Detail: detail, Status: status}
}

func CustomErrorHandler(ctx *fiber.Ctx, err error) error {
	var problemJSON *ProblemJSONError

	// Retrieve the custom status code if it's a fiber.*Error
	var e *fiber.Error
	if errors.Is(err, e) {
		problemJSON = &ProblemJSONError{Status: e.Code, Title: e.Message}
	}

	if problemJSON == nil {
		//nolint:errorlint
		switch e := err.(type) {
		case ProblemJSONError:
			problemJSON = &e
		default:
			problemJSON = &ProblemJSONError{Status: fiber.StatusNotFound, Title: fiber.ErrNotFound.Message, Detail: e.Error()}
		}
	}

	ctx.Status(problemJSON.Status)

	return ctx.JSON(problemJSON, "application/problem+json")
}
