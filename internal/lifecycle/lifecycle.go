package lifecycle

import (
	"io"
	"log"
	"os"
)

// WatchParentProcess listens to Stdin, if closed (parent process exits), then auto exit.
func WatchParentProcess() {
	go func() {
		buf := make([]byte, 1)
		_, err := os.Stdin.Read(buf)
		if err == io.EOF || err != nil {
			log.Println("Parent process stdin closed or error, exiting...")
			os.Exit(0)
		}
	}()
}
