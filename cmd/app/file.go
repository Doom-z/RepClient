package app

import (
	"bufio"
	"os"

	"github.com/Doom-z/RepClient/pkg/logger"
)

func StreamFile(file string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		f, err := os.Open(file)
		if err != nil {
			logger.Fatalf("failed to open file %s: %v", file, err)
			return
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				out <- line
			}
		}

		if err := scanner.Err(); err != nil {
			logger.Fatalf("error reading file %s: %v", file, err)
		}
	}()

	return out
}
