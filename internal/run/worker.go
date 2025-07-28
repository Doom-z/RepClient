package run

import (
	"strings"
	"sync"

	"github.com/Doom-z/RepClient/pkg/fileutil"
	"github.com/Doom-z/RepClient/pkg/logger"
)

func startSaveWorker(wg *sync.WaitGroup, tasks <-chan SaveTask) {
	defer wg.Done()
	for task := range tasks {
		switch task.Format {
		case "txt":
			if s, ok := task.Data.(string); ok {
				fileutil.SaveData(s, task.Path, "append")
			}
		default:
			fileutil.SaveData(task.Data, task.Path, "append")
		}
	}
}

func (r *Run) runWorker(jobs <-chan string, wg *sync.WaitGroup, workerID int, handler func(param, target string)) {
	defer wg.Done()
	logger.WithGID().Tracef("Worker %d started", workerID)

	for line := range jobs {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		r.handleStreamInput(line, handler)
	}
}
