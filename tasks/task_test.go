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
	allQueueOverload(t)
}

func singleQueueOverload(t *testing.T) {
	t.Helper()
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

	overloadChecker(t, nil, readyChan, "Queue Test1")

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

func allQueueOverload(t *testing.T) {
	t.Helper()
	taskQueueOne, _ := assignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test1"])
	taskQueueTwo, _ := assignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test2"])

	totalRequests := 8

	mockAddingRequests(t, (totalRequests / 2), taskQueueOne)
	mockAddingRequests(t, (totalRequests / 2), taskQueueTwo)

	readyChanOne := make(chan bool)
	readyChanTwo := make(chan bool)

	pendingChanOne := make(chan bool)
	pendingChanTwo := make(chan bool)

	taskQueueOne.TaskMu.Lock()
	queueOne := taskQueueOne.QueuedProcesses[0]
	taskQueueOne.QueuedProcesses = taskQueueOne.QueuedProcesses[1:]
	taskQueueOne.TaskMu.Unlock()

	taskQueueOne.UpdateOverloaded(queueOne, readyChanOne, pendingChanOne)

	taskQueueTwo.TaskMu.Lock()
	queueTwo := taskQueueTwo.QueuedProcesses[0]
	taskQueueTwo.QueuedProcesses = taskQueueTwo.QueuedProcesses[1:]
	taskQueueTwo.TaskMu.Unlock()

	taskQueueTwo.UpdateOverloaded(queueTwo, readyChanTwo, pendingChanTwo)

	taskManagerInstance.PendingTasks.PendingMu.RLock()
	pendingQueue := len(taskManagerInstance.PendingTasks.PendingQueue)
	if pendingQueue != totalRequests {
		testLoggingHelper(t, fmt.Sprintf("Pending tasks not adding upto (%d). Pending:: %d", totalRequests, pendingQueue), true)
	}
	taskManagerInstance.PendingTasks.PendingMu.RUnlock()

	var wg sync.WaitGroup
	wg.Add(4)

	go overloadChecker(t, &wg, readyChanOne, "Queue Test1")
	go overloadChecker(t, &wg, readyChanTwo, "Queue Test2")
	go overloadChecker(t, &wg, pendingChanOne, "Pending Test1")
	go overloadChecker(t, &wg, pendingChanTwo, "Pending Test2")

	if waitTimeout(&wg, 15*time.Second) {
		t.Error("Timeout waiting for all channels to receive signals")
	}

	_, firstQueuedProcesses := assignTaskAndQueue(t, taskQueueOne)
	_, secondQueuedProcesses := assignTaskAndQueue(t, taskQueueTwo)

	if (firstQueuedProcesses + secondQueuedProcesses) != totalRequests {
		testLoggingHelper(t, fmt.Sprintf("Tasks not adding upto total (%d)", totalRequests), true)
	}

	if firstQueuedProcesses == 0 {
		testLoggingHelper(t, "Test1 queue empty after overload", true)
	}
	if secondQueuedProcesses == 0 {
		testLoggingHelper(t, "Test2 queue empty after overload", true)
	}

}
