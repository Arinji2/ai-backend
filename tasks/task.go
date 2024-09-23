package tasks

import (
	"fmt"
	"math"
	"sync"
	"time"

	custom_log "github.com/Arinji2/ai-backend/logger"
)

func (task *TaskObject) ProcessTasks() {
	defer recoverPanic(task)

	if task.IsOverloaded {
		return
	}
	defer func() {
		task.TaskMu.Lock()
		task.IsProcessing = false
		task.TaskMu.Unlock()

		taskManagerInstance.CheckPendingTasks(task, nil)
	}()

	for len(task.QueuedProcesses) > 0 {

		task.TaskMu.Lock()
		task.IsProcessing = true
		queue := task.QueuedProcesses[0]
		task.QueuedProcesses = task.QueuedProcesses[1:]
		task.TaskMu.Unlock()
		response, err := GetPromptResponse(task, queue.Prompt)

		if err != nil {

			if err.Error() == "googleapi: Error 429: Resource has been exhausted (e.g. check quota)." {

				task.UpdateOverloaded(queue, nil, nil)
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

		loggableTime := int(math.Round(time.Since(queue.TimeStarted).Seconds()))

		custom_log.Logger.Debug(fmt.Sprintf("Fulfilled By %s In %d Seconds", task.DisplayName, loggableTime))

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

func (task *TaskObject) UpdateOverloaded(queue *QueuedProcess, readyChan chan bool, pendingChan chan bool) {
	if task.IsOverloaded {
		return
	}
	custom_log.Logger.Warn(fmt.Sprintf("%s is overloaded", task.DisplayName))
	timeForOverloaded := time.Now()

	task.TaskMu.Lock()
	task.IsOverloaded = true
	task.QueuedProcesses = append(task.QueuedProcesses, queue)

	task.TaskMu.Unlock()

	task.MoveQueueOut()

	go func() {
		var ticker *time.Ticker
		if !taskManagerInstance.IsTesting {
			ticker = time.NewTicker(time.Second * 5)
		} else {
			ticker = time.NewTicker(time.Second * 1)
		}
		defer ticker.Stop()

		for range ticker.C {
			var isReady bool
			if taskManagerInstance.IsTesting {
				isReady = true
			} else {
				isReady = GetModelStatus(task)
			}

			if isReady {
				task.TaskMu.Lock()
				task.IsOverloaded = false
				task.TaskMu.Unlock()

				readyTime := int(math.Round(time.Since(timeForOverloaded).Seconds()))
				custom_log.Logger.Warn(fmt.Sprintf("%s is ready in %d seconds", task.DisplayName, readyTime))

				taskManagerInstance.TaskQueueUnloaded(task)

				taskManagerInstance.CheckPendingTasks(task, pendingChan)

				if readyChan != nil {
					readyChan <- true
				}

				if pendingChan != nil {
					pendingChan <- true
				}

				break
			}
		}
	}()
}

func (taskQueue *TaskObject) MoveQueueOut() {
	taskQueue.TaskMu.Lock()

	queueCopy := make([]*QueuedProcess, len(taskQueue.QueuedProcesses))

	copy(queueCopy, taskQueue.QueuedProcesses)
	taskQueue.TaskMu.Unlock()

	var wg sync.WaitGroup
	for _, task := range queueCopy {
		wg.Add(1)
		go func(t *QueuedProcess) {
			defer wg.Done()
			taskManagerInstance.MoveAddedRequest(t.Prompt, t.Done)
		}(task)
	}
	taskQueue.QueuedProcesses = []*QueuedProcess{}
	wg.Wait()
	taskQueue.TaskMu.Lock()
	if len(taskQueue.QueuedProcesses) > 0 {
		custom_log.Logger.Warn("Queue was not empty after clearing. Current length:", len(taskQueue.QueuedProcesses))
	}
	taskQueue.TaskMu.Unlock()
}
func recoverPanic(task *TaskObject) {
	r := recover()
	if r == nil {
		return
	}
	custom_log.Logger.Error(fmt.Sprintf("Recovered from panic in %s for %s", task.DisplayName, r))
}
