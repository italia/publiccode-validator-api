package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
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
	validateFunc        func(t *testing.T, response map[string]any)
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

func loadTestdata(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("cannot read testdata/%s: %v", name, err)
	}
	return string(b)
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

			res, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
			assert.Nil(t, err)

			assert.Equal(t, test.expectedCode, res.StatusCode)

			body, err := io.ReadAll(res.Body)

			assert.Nil(t, err)

			if test.validateFunc != nil {
				var bodyMap map[string]any
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
			expectedBody:        `{"title":"Not Found","detail":"Not Found","status":404}`,
			expectedContentType: "application/problem+json",
		},
	}

	runTestCases(t, tests)
}

func TestValidateEndpoint(t *testing.T) {
	validYml := loadTestdata(t, "valid.publiccode.yml")
	invalidYml := loadTestdata(t, "invalid.publiccode.yml")
	sourceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid.publiccode.yml":
			_, _ = w.Write([]byte(validYml))
		case "/invalid.publiccode.yml":
			_, _ = w.Write([]byte(invalidYml))
		case "/empty.publiccode.yml":
			// Return 200 with an empty response body
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer sourceServer.Close()

	validYmlURL := url.QueryEscape(sourceServer.URL + "/valid.publiccode.yml")
	emptyYmlURL := url.QueryEscape(sourceServer.URL + "/empty.publiccode.yml")
	notFoundYmlURL := url.QueryEscape(sourceServer.URL + "/missing.publiccode.yml")

	tests := []TestCase{
		{
			description:         "validate: empty body",
			query:               "QUERY /v1/validate",
			body:                "",
			expectedCode:        400,
			expectedBody:        `{"title":"empty body","detail":"need a body to validate","status":400}`,
			expectedContentType: "application/problem+json",
		},
		{
			description:         "validate: valid file",
			query:               "QUERY /v1/validate",
			body:                validYml,
			expectedCode:        200,
			expectedContentType: "application/json; charset=utf-8",
			validateFunc: func(t *testing.T, response map[string]any) {
				assert.Equal(t, true, response["valid"])
				results, ok := response["results"].([]any)
				assert.True(t, ok)
				assert.Len(t, results, 0)
				assert.NotNil(t, response["normalized"])
			},
		},
		{
			description:         "validate: invalid file",
			query:               "QUERY /v1/validate",
			body:                invalidYml,
			expectedCode:        200,
			expectedContentType: "application/json; charset=utf-8",
			validateFunc: func(t *testing.T, response map[string]any) {
				assert.Equal(t, false, response["valid"])
				results, ok := response["results"].([]any)
				assert.True(t, ok)
				assert.NotEmpty(t, results)
				assert.Nil(t, response["normalized"])
			},
		},
		{
			description:         "validate: valid file from URL query parameter",
			query:               "QUERY /v1/validate?url=" + validYmlURL,
			expectedCode:        200,
			expectedContentType: "application/json; charset=utf-8",
			validateFunc: func(t *testing.T, response map[string]any) {
				assert.Equal(t, true, response["valid"])
				results, ok := response["results"].([]any)
				assert.True(t, ok)
				assert.Len(t, results, 0)
				assert.NotNil(t, response["normalized"])
			},
		},
		{
			description:         "validate: invalid URL query parameter",
			query:               "QUERY /v1/validate?url=not-a-url",
			expectedCode:        400,
			expectedBody:        `{"title":"invalid url","detail":"query parameter 'url' must be a valid http(s) URL","status":400}`,
			expectedContentType: "application/problem+json",
		},
		{
			description:         "validate: URL query parameter returns empty body",
			query:               "QUERY /v1/validate?url=" + emptyYmlURL,
			expectedCode:        400,
			expectedBody:        `{"title":"empty body","detail":"the URL response is empty","status":400}`,
			expectedContentType: "application/problem+json",
		},
		{
			description:         "validate: URL query parameter returns non-2xx",
			query:               "QUERY /v1/validate?url=" + notFoundYmlURL,
			expectedCode:        400,
			expectedBody:        `{"title":"url fetch failed","detail":"failed to fetch URL: HTTP 404","status":400}`,
			expectedContentType: "application/problem+json",
		},
	}

	runTestCases(t, tests)
}

func TestStatusEndpoints(t *testing.T) {
	tests := []TestCase{
		{
			query:               "GET /v1/status",
			expectedCode:        200,
			expectedBody:        `{"v":"dev","commit":"-"}`,
			expectedContentType: "application/json; charset=utf-8",
			// TODO: test cache headers
		},
	}

	runTestCases(t, tests)
}
