package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	publiccodeParser "github.com/italia/publiccode-parser-go/v5"
)

type PubliccodeymlValidatorHandler struct {
	parser *publiccodeParser.Parser
}

func NewPubliccodeymlValidatorHandler() *PubliccodeymlValidatorHandler {
	parser, err := publiccodeParser.NewDefaultParser()
	if err != nil {
		panic(fmt.Sprintf("can't create a publiccode.yml parser: %s", err.Error()))
	}

	return &PubliccodeymlValidatorHandler{parser: parser}
}

func (vh *PubliccodeymlValidatorHandler) Query(ctx *fiber.Ctx) error {
	valid := true

	// if all := ctx.QueryBool("all", false); !all {
	// 	stmt = stmt.Scopes(models.Active)
	// }

	// ct := c.Get("Content-Type")
	// if !strings.Contains(ct, "yaml") && ct != "application/octet-stream" {
	// 	return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
	// 		"error": "unsupported content-type",
	// 	})
	// }

	if len(ctx.Body()) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "empty body",
		})
	}

	// Body() ritorna []byte; lo wrappiamo in io.Reader
	reader := strings.NewReader(string(ctx.Body()))
	// alternativa più efficiente:
	// reader := bytes.NewReader(c.Body())

	// parsed, err = parser.Parse(repository.FileRawURL)
	_, err := vh.parser.ParseStream(reader)
	if err != nil {
		var validationResults publiccodeParser.ValidationResults
		if errors.As(err, &validationResults) {
			var validationError publiccodeParser.ValidationError
			for _, res := range validationResults {
				if errors.As(res, &validationError) {
					valid = false

					break
				}
			}
		}
	}

	return ctx.JSON(fiber.Map{"valid": valid})
}
