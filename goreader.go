// Demo code for the Flex primitive.
package main

import (
	"log/slog"
	"os"

	"github.com/taylorskalyo/goreader/views"
)

func main() {
	app := views.NewApplication()
	if err := app.Run(); err != nil {
		if err != views.ErrUsage {

			// Assume non-usage errors are fatal.
			slog.Error("Encountered fatal error.", "error", err)
		}

		os.Exit(1)
	}

	os.Exit(0)
}
