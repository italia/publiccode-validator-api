package handlers

import (
	"bytes"
	"errors"

	"github.com/gofiber/fiber/v3"
	publiccodeParser "github.com/italia/publiccode-parser-go/v5"
	"github.com/italia/publiccode-validator-api/internal/common"
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

func (vh *PubliccodeymlValidatorHandler) Query(ctx fiber.Ctx) error {
	var normalized *string

	valid := true
	parser := vh.parser

	if checks := fiber.Query[bool](ctx, "external-checks", false); checks {
		parser = vh.parserExternalChecks
	}

	if len(ctx.Body()) == 0 {
		return common.Error(fiber.StatusBadRequest, "empty body", "need a body to validate")
	}

	results := make(publiccodeParser.ValidationResults, 0)

	reader := bytes.NewReader(ctx.Body())

	parsed, err := parser.ParseStream(reader)
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

	if valid && parsed != nil {
		yaml, err := parsed.ToYAML()
		if err == nil {
			s := string(yaml)
			normalized = &s
		}
	}

	//nolint:wrapcheck
	return ctx.JSON(fiber.Map{"valid": valid, "results": results, "normalized": normalized})
}
