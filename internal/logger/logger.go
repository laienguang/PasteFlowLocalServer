package logger

import (
	"log"
	"os"
)

// Setup sets up logging to stdout only.
func Setup() {
	// Go's default logging includes date and time
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)
}
