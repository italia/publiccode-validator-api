# publiccode.yml RESTful validator API

[![Join the #publiccode channel](https://img.shields.io/badge/Slack%20channel-%23publiccode-blue.svg?logo=slack)](https://developersitalia.slack.com/messages/CAM3F785T)
[![Get invited](https://slack.developers.italia.it/badge.svg)](https://slack.developers.italia.it/)

A RESTful API for validating [publiccode.yml](https://github.com/italia/publiccode.yml)
files, using the [publiccode-parser-go](https://github.com/italia/publiccode-parser-go)
library.

## Usage

```console
go run main.go
```

### Validate a publiccode.yml file

```console
curl -X QUERY "http://localhost:3000/v1/validate" -H "Content-Type: application/yaml" --data-binary "@./publiccode.yml"
```

or

```console
curl -X QUERY "http://localhost:3000/v1/validate" -H "Content-Type: application/yaml" \
--data-binary $'publiccodeYmlVersion: "0.5"\ndevelopmentStatus: stable\n [... rest of the data ...]'
```

### Example response (valid publiccode.yml)

```json
{
  "valid": true,
  "results": []
}
```

### Example response (with errors / warnings)

```json
{
  "valid": false,
  "results": [
    {
      "type": "error",
      "key": "legal.license",
      "description": "license must be a valid license (see https://spdx.org/licenses)",
      "line": 12,
      "column": 5
    },
    {
      "type": "warning",
      "key": "publiccodeYmlVersion",
      "description": "v0.2 is not the latest version, use '0'. Parsing this file as v0.5",
      "line": 1,
      "column": 1
    }
  ]
}
```

### Example error response

All error responses are returned using the `application/problem+json` media type,
in accordance with [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457.html).

```http
HTTP/1.1 400 Bad Request
Content-Type: application/problem+json
```

```json
{
  "title": "empty body",
  "detail": "need a body to validate",
  "status": 400
}
```

## Contributing

Contributing is always appreciated, see [CONTRIBUTING.md](CONTRIBUTING.md).
Feel free to open issues, fork or submit a Pull Request.

## Maintainers

This software is maintained by community maintainers.

## License

© 2018-2020 Team per la Trasformazione Digitale - Presidenza del Consiglio dei Minstri

Licensed under AGPL-3.0.
The version control system provides attribution for specific lines of code.
