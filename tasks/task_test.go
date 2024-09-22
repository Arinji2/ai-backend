package tasks

import (
	"fmt"
	"testing"
	"time"
)

//TODO CHANGE FROM TESTINGMODE BOOL

func TestUpdateOverloaded(t *testing.T) {
	TestNewTaskManager(t)

	for i := 0; i < 4; i++ {
		taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses = append(taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses, &QueuedProcess{
			Prompt:      fmt.Sprintf("test%d in test1", i),
			Done:        make(chan ResponseChan),
			TimeStarted: time.Now(),
		})
	}

	for i := 0; i < 4; i++ {
		taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses = append(taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses, &QueuedProcess{
			Prompt:      fmt.Sprintf("test%d in test2", i),
			Done:        make(chan ResponseChan),
			TimeStarted: time.Now(),
		})
	}

	readyChan := make(chan bool)

	queue := taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses[0]
	taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses = taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses[1:]
	taskManagerInstance.AllTasks.Tasks["test1"].UpdateOverloaded(queue, true, readyChan)

	secondQueuedProcesses := len(taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses)
	if secondQueuedProcesses != 8 {
		t.Error("Tasks not adding up to 8", secondQueuedProcesses)
	}
	fmt.Println(secondQueuedProcesses)

	select {
	case isReady := <-readyChan:
		if !isReady {
			t.Error("Task did not become ready after overload")
		}
	case <-time.After(10 * time.Second):
		t.Error("Timeout waiting for task to become ready")
	}

	firstQueuedProcesses := len(taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses)
	secondQueuedProcesses = len(taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses)

	fmt.Println(firstQueuedProcesses, secondQueuedProcesses)

	if secondQueuedProcesses == 8 {
		t.Error("Tasks not distributing correctly")
	}

	if firstQueuedProcesses+secondQueuedProcesses != 8 {
		t.Error("Tasks not adding up to 8", firstQueuedProcesses, secondQueuedProcesses)
	}
}
