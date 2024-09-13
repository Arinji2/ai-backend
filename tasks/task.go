package tasks

import (
	"fmt"
	"sync"
	"time"

	custom_log "github.com/Arinji2/ai-backend/logger"
)

func (task *TaskObject) ProcessTasks() {
	defer recoverPanic()

	if task.IsOverloaded {
		return
	}

	defer func() {

		taskManagerInstance.CheckPendingTasks(task)
	}()

	for len(task.QueuedProcesses) > 0 {

		task.TaskMu.Lock()
		queue := task.QueuedProcesses[0]
		task.QueuedProcesses = task.QueuedProcesses[1:]
		task.TaskMu.Unlock()

		response, err := GetPromptResponse(task, queue.Prompt)

		if err != nil {

			if err.Error() == "googleapi: Error 429: Resource has been exhausted (e.g. check quota)." {

				task.TaskMu.Lock()
				task.QueuedProcesses = append(task.QueuedProcesses, queue)
				task.TaskMu.Unlock()
				task.UpdateOverloaded()
				continue
			}
			queue.Retries++
			if queue.Retries > MaxRetries {
				queue.ErrorWithProcess(err)
				continue
			}
			task.TaskMu.Lock()
			task.QueuedProcesses = append(task.QueuedProcesses, queue)
			task.TaskMu.Unlock()
			continue
		}

		fmt.Println("Fullfilled via ", task.ApiKey)

		select {
		case queue.Done <- ResponseChan{Response: response}:

		default:

		}
	}
}

func (process *QueuedProcess) ErrorWithProcess(err error) {
	select {
	case process.Done <- ResponseChan{Response: "Error Processing Task: " + err.Error()}:

	default:

	}
}

func (task *TaskObject) UpdateOverloaded() {
	custom_log.Logger.Warn(task.ApiKey + "IS OVERLOADED")
	task.TaskMu.Lock()
	task.IsOverloaded = true

	task.TaskMu.Unlock()

	task.MoveQueueOut()

	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for range ticker.C {
			defer ticker.Stop()
			isReady := GetModelStatus(task)

			if isReady {
				task.TaskMu.Lock()
				task.IsOverloaded = false

				task.TaskMu.Unlock()
				custom_log.Logger.Warn(task.ApiKey + "IS READY")

				taskManagerInstance.TaskQueueUnloaded(task) // Ensure this is only called after it's ready
				break
			}
		}
	}()
}

func (taskQueue *TaskObject) MoveQueueOut() {
	taskQueue.TaskMu.Lock()
	defer taskQueue.TaskMu.Unlock()

	// Create a copy of the queue to iterate over
	queueCopy := make([]*QueuedProcess, len(taskQueue.QueuedProcesses))
	copy(queueCopy, taskQueue.QueuedProcesses)

	var wg sync.WaitGroup
	for _, task := range queueCopy {
		wg.Add(1)
		go func(t *QueuedProcess) {
			defer wg.Done()
			taskManagerInstance.MoveAddedRequest(t.Prompt, t.Done)
		}(task)
	}

	// Clear the queue immediately
	taskQueue.QueuedProcesses = []*QueuedProcess{}

	// Wait for all tasks to be moved
	wg.Wait()

	// Double-check that the queue is still empty
	if len(taskQueue.QueuedProcesses) > 0 {
		custom_log.Logger.Warn("Queue was not empty after clearing. Current length:", len(taskQueue.QueuedProcesses))
	}
}
func recoverPanic() {
	if r := recover(); r != nil {

		// Optionally restart the goroutine or perform other recovery actions
	}
}
