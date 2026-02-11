package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

var (
	app *fiber.App
)

type TestCase struct {
	description string

	// Test input
	query   string
	body    string
	headers map[string][]string

	// Expected output
	expectedCode        int
	expectedBody        string
	expectedContentType string
	validateFunc        func(t *testing.T, response map[string]interface{})
}

func init() {
	_ = os.Setenv("ENVIRONMENT", "test")

	// Setup the app as it is done in the main function
	app = Setup()
}

func TestMain(m *testing.M) {
	code := m.Run()

	os.Exit(code)
}

func runTestCases(t *testing.T, tests []TestCase) {
	for _, test := range tests {
		description := test.description
		if description == "" {
			description = test.query
		}

		t.Run(description, func(t *testing.T) {
			query := strings.Split(test.query, " ")

			u, err := url.Parse(query[1])
			if err != nil {
				assert.Fail(t, err.Error())
			}

			req, err := http.NewRequest(
				query[0],
				query[1],
				strings.NewReader(test.body),
			)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			if test.headers != nil {
				req.Header = test.headers
			}
			req.URL.RawQuery = u.Query().Encode()

			res, err := app.Test(req, -1)
			assert.Nil(t, err)

			assert.Equal(t, test.expectedCode, res.StatusCode)

			body, err := io.ReadAll(res.Body)

			assert.Nil(t, err)

			if test.validateFunc != nil {
				var bodyMap map[string]interface{}
				err = json.Unmarshal(body, &bodyMap)
				assert.Nil(t, err)

				test.validateFunc(t, bodyMap)
				if t.Failed() {
					log.Printf("\nAPI response:\n%s\n", body)
				}
			} else {
				assert.Equal(t, test.expectedBody, string(body))
			}

			assert.Equal(t, test.expectedContentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestApi(t *testing.T) {
	tests := []TestCase{
		{
			description: "non existing route",
			query:       "GET /v1/i-dont-exist",

			expectedCode:        404,
			expectedBody:        `{"title":"Not Found","detail":"Cannot GET /v1/i-dont-exist","status":404}`,
			expectedContentType: "application/problem+json",
		},
	}

	runTestCases(t, tests)
}

func TestStatusEndpoints(t *testing.T) {
	tests := []TestCase{
		{
			query:               "GET /v1/status",
			expectedCode:        204,
			expectedBody:        "",
			expectedContentType: "",
			// TODO: test cache headers
		},
	}

	runTestCases(t, tests)
}
