Telegram handler for slog
=========================
This is a handler for the [slog](https://pkg.go.dev/log/slog) package that sends log messages to a [Telegram](https://telegram.org/) chat.

# Installation
```bash
go get github.com/platx/slog-telegram
```

Minimum Go version: 1.22

# Usage
## Basic usage
```go
package main

import (
	"log/slog"

	slogtelegram "github.com/platx/slog-telegram"
)

func main() {
	handler := slogtelegram.DefaultHandler(
		"YOUR_BOT_TOKEN", // Bot token (https://t.me/botfather)
		1234567890, // Chat ID (https://t.me/get_id_bot)
	)

	logger := slog.New(handler)

	logger.Info("Hello, World!", slog.Any("key1", "val1"))
}
```

## Usage examples
* [batch processing](./examples/batch/main.go)
* [custom formatter](./examples/format/main.go)

# License
MIT licensed. See the [LICENSE](./LICENSE) file for details.
