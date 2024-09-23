package tasks

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestUpdateOverloaded(t *testing.T) {
	TestNewTaskManager(t)

	singleQueueOverload(t)
	taskManagerInstance.AllTasks.Tasks["test1"].TaskMu.Lock()
	for i := 0; i < 4; i++ {
		taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses = append(taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses, &QueuedProcess{
			Prompt:      fmt.Sprintf("test%d in test1", i),
			Done:        make(chan ResponseChan),
			TimeStarted: time.Now(),
		})
	}
	taskManagerInstance.AllTasks.Tasks["test1"].TaskMu.Unlock()
	taskManagerInstance.AllTasks.Tasks["test2"].TaskMu.Lock()
	for i := 0; i < 4; i++ {
		taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses = append(taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses, &QueuedProcess{
			Prompt:      fmt.Sprintf("test%d in test2", i),
			Done:        make(chan ResponseChan),
			TimeStarted: time.Now(),
		})
	}
	taskManagerInstance.AllTasks.Tasks["test2"].TaskMu.Unlock()

	readyChanOne := make(chan bool)
	readyChanTwo := make(chan bool)

	pendingChanOne := make(chan bool)
	pendingChanTwo := make(chan bool)

	queueOne := taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses[0]
	taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses = taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses[1:]
	taskManagerInstance.AllTasks.Tasks["test1"].UpdateOverloaded(queueOne, readyChanOne, pendingChanOne)
	queueTwo := taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses[0]
	taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses = taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses[1:]
	taskManagerInstance.AllTasks.Tasks["test2"].UpdateOverloaded(queueTwo, readyChanTwo, pendingChanTwo)
	taskManagerInstance.PendingTasks.PendingMu.RLock()
	if len(taskManagerInstance.PendingTasks.PendingQueue) != 8 {
		t.Error("Pending tasks not adding up to 8", len(taskManagerInstance.PendingTasks.PendingQueue))
	}
	taskManagerInstance.PendingTasks.PendingMu.RUnlock()

	var wg sync.WaitGroup
	wg.Add(4) // We expect 4 signals

	go func() {
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

	go func() {
		defer wg.Done()
		select {
		case isReady := <-readyChanTwo:
			if !isReady {
				t.Error("Task two did not become ready after overload")
			}
		case <-time.After(10 * time.Second):
			t.Error("Timeout waiting for task two to become ready")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case isReady := <-pendingChanOne:
			if !isReady {
				t.Error("Pending one did not become ready after overload")
			}

		case <-time.After(10 * time.Second):
			t.Error("Timeout waiting for pending one to become ready")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case isReady := <-pendingChanTwo:
			if !isReady {
				t.Error("Pending two did not become ready after overload")
			}
		case <-time.After(10 * time.Second):
			t.Error("Timeout waiting for pending two to become ready")
		}
	}()

	// Use a timeout for the entire wait
	if waitTimeout(&wg, 15*time.Second) {
		t.Error("Timeout waiting for all channels to receive signals")
	}

	taskManagerInstance.AllTasks.Tasks["test1"].TaskMu.RLock()
	taskManagerInstance.AllTasks.Tasks["test2"].TaskMu.RLock()
	if len(taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses)+len(taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses) != 8 {
		t.Error("Tasks not adding up to 8", len(taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses), len(taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses))
	}

	if len(taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses) == 0 {
		t.Error("Test1 queue empty after overload", len(taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses))
	}

	if len(taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses) == 0 {
		t.Error("Test2 queue empty after overload", len(taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses))
	}
	taskManagerInstance.AllTasks.Tasks["test1"].TaskMu.RUnlock()
	taskManagerInstance.AllTasks.Tasks["test2"].TaskMu.RUnlock()
}

// Helper function
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func singleQueueOverload(t *testing.T) {
	taskQueueOne, _ := assignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test1"])
	taskQueueTwo, _ := assignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test2"])
	totalRequests := 8

	mockAddingRequests(t, (totalRequests / 2), taskQueueOne)
	mockAddingRequests(t, (totalRequests / 2), taskQueueTwo)
	readyChan := make(chan bool)

	taskQueueOne.TaskMu.Lock()
	queue := taskQueueOne.QueuedProcesses[0]
	taskQueueOne.QueuedProcesses = taskQueueOne.QueuedProcesses[1:]
	taskQueueOne.TaskMu.Unlock()

	taskQueueOne.UpdateOverloaded(queue, readyChan, nil)

	_, secondQueuedProcesses := assignTaskAndQueue(t, taskQueueTwo)
	if secondQueuedProcesses != totalRequests {
		testLoggingHelper(t, fmt.Sprintf("All requests not moved to Queue2. Total:: %d", totalRequests), true)
	}

	select {
	case isReady := <-readyChan:
		if !isReady {
			t.Error("Queue 1 did not become ready after overload")
		}
	case <-time.After(10 * time.Second):
		t.Error("Timeout waiting for queue 1 to become ready")
	}

	_, firstQueuedProcesses := assignTaskAndQueue(t, taskQueueOne)
	_, secondQueuedProcesses = assignTaskAndQueue(t, taskQueueTwo)

	if secondQueuedProcesses == totalRequests || firstQueuedProcesses == totalRequests {
		testLoggingHelper(t, "Tasks not distributing correctly among queues", true)
	}

	if firstQueuedProcesses+secondQueuedProcesses != totalRequests {

		testLoggingHelper(t, fmt.Sprintf("Tasks not adding upto total (%d)", totalRequests), true)
	}
	resetTaskQueue(t, taskQueueOne)
	resetTaskQueue(t, taskQueueTwo)
}
