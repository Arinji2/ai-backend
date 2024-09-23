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
		taskManagerInstance.AllTasks.Tasks["test1"].TaskMu.RLock()
		taskManagerInstance.AllTasks.Tasks["test2"].TaskMu.RLock()
		t.Errorf("%s. Queue Lengths: First:: %d, Second:: %d", message, len(taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses), len(taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses))
		taskManagerInstance.AllTasks.Tasks["test1"].TaskMu.RUnlock()
		taskManagerInstance.AllTasks.Tasks["test2"].TaskMu.RUnlock()
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

func overloadChecker(t *testing.T, wg *sync.WaitGroup, readyChanOne chan bool, taskName string) {
	func() {
		defer func() {
			if wg != nil {
				wg.Done()
			}
		}()
		select {
		case isReady := <-readyChanOne:
			if !isReady {
				t.Errorf("%s did not become ready after overload", taskName)
			}
		case <-time.After(10 * time.Second):
			t.Errorf("Timeout waiting for %s to become ready", taskName)
		}
	}()
}
