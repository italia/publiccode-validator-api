# publiccode.yml web validator for Go

[![Join the #publiccode channel](https://img.shields.io/badge/Slack%20channel-%23publiccode-blue.svg?logo=slack)](https://developersitalia.slack.com/messages/CAM3F785T)
[![Get invited](https://slack.developers.italia.it/badge.svg)](https://slack.developers.italia.it/)

A RESTful API for validating [publiccode.yml](https://github.com/italia/publiccode.yml)
files, using the [publiccode-parser-go](https://github.com/italia/publiccode-parser-go)
library.

## Usage

```bash
go run main.go

curl -X POST "http://localhost:3000/v1/validate"
  -H "Content-Type: application/x-yaml"
  --data-binary "@./publiccode.yml"
```

## Contributing

Contributing is always appreciated.
Feel free to open issues, fork or submit a Pull Request.
If you want to know more about how to add new fields, check out [CONTRIBUTING.md](CONTRIBUTING.md). In order to support other country-specific extensions in addition to Italy some refactoring might be needed.

## Maintainers

This software is maintained by community maintainers.

## License

© 2018-2020 Team per la Trasformazione Digitale - Presidenza del Consiglio dei Minstri

Licensed under AGPL-3.0.
The version control system provides attribution for specific lines of code.
