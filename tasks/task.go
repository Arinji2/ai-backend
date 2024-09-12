package tasks

import (
	"fmt"

	custom_log "github.com/Arinji2/ai-backend/logger"
)

func (task *TaskObject) ProcessTasks() {
	task.TaskMu.Lock()
	if task.IsProcessing {
		task.TaskMu.Unlock()
		return
	}

	task.IsProcessing = true
	task.TaskMu.Unlock()

	defer func() {
		task.TaskMu.Lock()
		task.IsProcessing = false
		task.TaskMu.Unlock()

	}()

	for len(task.QueuedProcesses) > 0 {

		custom_log.Logger.Debug("Processing Task")

		task.TaskMu.Lock()
		queue := task.QueuedProcesses[0]
		task.QueuedProcesses = task.QueuedProcesses[1:]
		task.TaskMu.Unlock()

		custom_log.Logger.Info(fmt.Sprintf("Prompt: %s", queue.Prompt))
		custom_log.Logger.Debug("Task Completed")
		close(queue.Done)

	}
}
