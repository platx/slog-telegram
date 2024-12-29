# Telegram handler for slog
[![Release](https://img.shields.io/github/release/platx/slog-telegram.svg?style=flat-square)](https://github.com/platx/slog-telegram/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](LICENSE)
[![codecov](https://codecov.io/github/platx/slog-telegram/graph/badge.svg?token=LYZMRRHG3P)](https://codecov.io/github/platx/slog-telegram)
[![godoc](https://godoc.org/github.com/platx/slog-telegram?status.svg)](https://godoc.org/github.com/platx/slog-telegram)
[![Go Report Card](https://goreportcard.com/badge/github.com/platx/slog-telegram?style=flat-square)](https://goreportcard.com/report/github.com/platx/slog-telegram)

This is a handler for the [slog](https://pkg.go.dev/log/slog) package that sends log messages to a [Telegram](https://telegram.org/) chat.

## Installation
```bash
go get github.com/platx/slog-telegram
```
Minimum Go version: 1.23

## Usage
* [basic usage](./docs/base.md)
* [batch processing](./docs/batch.md)
* [custom formatter](./docs/format.md)

## License
MIT licensed. See the [LICENSE](./LICENSE) file for details.
