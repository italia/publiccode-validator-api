package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	publiccodeParser "github.com/italia/publiccode-parser-go/v5"
	"github.com/italia/publiccode-validator-api/internal/common"
)

type PubliccodeymlValidatorHandler struct {
	parser               *publiccodeParser.Parser
	parserExternalChecks *publiccodeParser.Parser
	httpClient           *http.Client
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

	return &PubliccodeymlValidatorHandler{
		parser:               parser,
		parserExternalChecks: parserExternalChecks,
		httpClient:           &http.Client{Timeout: 10 * time.Second},
	}
}

func (vh *PubliccodeymlValidatorHandler) Query(ctx fiber.Ctx) error {
	var normalized *string

	valid := true
	parser := vh.parser

	if checks := fiber.Query[bool](ctx, "external-checks", false); checks {
		parser = vh.parserExternalChecks
	}

	input := ctx.Body()
	if len(input) == 0 {
		rawURL := strings.TrimSpace(fiber.Query[string](ctx, "url", ""))
		if rawURL == "" {
			return common.Error(fiber.StatusBadRequest, "empty body", "need a body to validate")
		}

		content, err := vh.fetchURL(rawURL)
		if err != nil {
			return err
		}

		input = content
	}

	results := make(publiccodeParser.ValidationResults, 0)

	reader := bytes.NewReader(input)

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

func (vh *PubliccodeymlValidatorHandler) fetchURL(rawURL string) ([]byte, error) {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil || parsedURL.Host == "" {
		return nil, common.Error(fiber.StatusBadRequest, "invalid url", "query parameter 'url' must be a valid http(s) URL")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, common.Error(fiber.StatusBadRequest, "invalid url", "query parameter 'url' must use http or https")
	}

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, common.Error(fiber.StatusBadRequest, "invalid url", "query parameter 'url' is invalid")
	}

	resp, err := vh.httpClient.Do(req)
	if err != nil {
		return nil, common.Error(fiber.StatusBadRequest, "url fetch failed", fmt.Sprintf("failed to fetch URL: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, common.Error(
			fiber.StatusBadRequest,
			"url fetch failed",
			fmt.Sprintf("failed to fetch URL: HTTP %d", resp.StatusCode),
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, common.Error(fiber.StatusBadRequest, "url fetch failed", "failed to read response body")
	}

	if len(body) == 0 {
		return nil, common.Error(fiber.StatusBadRequest, "empty body", "the URL response is empty")
	}

	return body, nil
}
