package handlers

import (
	"bytes"
	"errors"

	"github.com/gofiber/fiber/v2"
	publiccodeParser "github.com/italia/publiccode-parser-go/v5"
)

type PubliccodeymlValidatorHandler struct {
	parser               *publiccodeParser.Parser
	parserExternalChecks *publiccodeParser.Parser
}

func NewPubliccodeymlValidatorHandler() *PubliccodeymlValidatorHandler {
	parser, err := publiccodeParser.NewParser(publiccodeParser.ParserConfig{DisableExternalChecks: true})
	if err != nil {
		panic("can't create a publiccode.yml parser: " + err.Error())
	}

	parserExternalChecks, err := publiccodeParser.NewDefaultParser()
	if err != nil {
		panic("can't create a publiccode.yml parser: " + err.Error())
	}

	return &PubliccodeymlValidatorHandler{parser: parser, parserExternalChecks: parserExternalChecks}
}

func (vh *PubliccodeymlValidatorHandler) Query(ctx *fiber.Ctx) error {
	valid := true
	parser := vh.parser

	if checks := ctx.QueryBool("external-checks", false); checks {
		parser = vh.parserExternalChecks
	}

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

	results := make(publiccodeParser.ValidationResults, 0)

	reader := bytes.NewReader(ctx.Body())

	_, err := parser.ParseStream(reader)
	if err != nil {
		var validationResults publiccodeParser.ValidationResults
		if errors.As(err, &validationResults) {
			var validationError publiccodeParser.ValidationError
			for _, res := range validationResults {
				if errors.As(res, &validationError) {
					valid = false
				}

				results = append(results, res)
			}
		}
	}

	return ctx.JSON(fiber.Map{"valid": valid, "results": results})
}
