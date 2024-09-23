package tasks

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func assignTaskAndQueue(t *testing.T, task *TaskObject) (*TaskObject, int) {
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

func mockAddingRequests(t *testing.T, count int, task *TaskObject) {
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

func testLoggingHelper(t *testing.T, message string, showLengths bool) {
	if showLengths {
		t.Errorf("%s. Queue Lengths: First:: %d, Second:: %d", message, len(taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses), len(taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses))
	} else {
		t.Error(message)
	}
}

func resetTaskQueue(t *testing.T, task *TaskObject) {

	task.TaskMu.Lock()
	task.QueuedProcesses = []*QueuedProcess{}
	if len(task.QueuedProcesses) != 0 {
		t.Error("Queue not empty after reset", len(task.QueuedProcesses), task.DisplayName)
	}
	task.TaskMu.Unlock()
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}

func overloadChecker(wg *sync.WaitGroup, readyChanOne chan bool, t *testing.T) {
	func() {
		defer wg.Done()
		select {
		case isReady := <-readyChanOne:
			if !isReady {
				t.Error("Task one did not become ready after overload")
			}
		case <-time.After(10 * time.Second):
			t.Error("Timeout waiting for task one to become ready")
		}
	}()
}
