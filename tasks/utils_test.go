package tasks

import (
	"fmt"
	"testing"
	"time"
)

func AssignTaskAndQueue(t *testing.T, task *TaskObject) (*TaskObject, int) {
	t.Helper()
	task.TaskMu.RLock()
	taskQueue := task
	taskQueueLength := len(taskQueue.QueuedProcesses)

	if taskQueueLength != len(task.QueuedProcesses) {
		t.Errorf("Length for task %s is not equal to length of queue %d != %d", task.DisplayName, taskQueueLength, len(task.QueuedProcesses))
	}
	task.TaskMu.RUnlock()

	return taskQueue, taskQueueLength

}

func MockAddingRequests(t *testing.T, count int, task *TaskObject) {
	t.Helper()
	for i := 0; i < count; i++ {
		task.TaskMu.Lock()
		task.QueuedProcesses = append(task.QueuedProcesses, &QueuedProcess{
			Prompt:      fmt.Sprintf("test%d for %s", i, task.DisplayName),
			Done:        make(chan ResponseChan),
			TimeStarted: time.Now(),
		})
		task.TaskMu.Unlock()
	}

}
