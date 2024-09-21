package tasks

import (
	"fmt"
	"testing"
	"time"
)

var testingTaskManager *TaskManager

func TestNewTaskManager(t *testing.T) {
	optionalKeys := []JsonKeys{
		{
			Name: "test1",
			Key:  "test1",
		},
		{
			Name: "test2",
			Key:  "test2",
		},
	}
	taskManager := NewTaskManager(optionalKeys)
	if taskManager.AllTasks.Tasks["test1"].DisplayName != "test1" {
		t.Error("Task Manager not initialized correctly")
	}
	if taskManager.AllTasks.Tasks["test2"].DisplayName != "test2" {
		t.Error("Task Manager not initialized correctly")
	}

	testingTaskManager = taskManager
	taskManagerInstance = taskManager
}

func TestAddRequest(t *testing.T) {
	TestNewTaskManager(t)
	testingTaskManager.AddRequest("test", true)

	firstQueuedProcesses := len(testingTaskManager.AllTasks.Tasks["test1"].QueuedProcesses)
	secondQueuedProcesses := len(testingTaskManager.AllTasks.Tasks["test2"].QueuedProcesses)

	if (firstQueuedProcesses + secondQueuedProcesses) != 1 {
		t.Error("Task not added correctly (1)", firstQueuedProcesses, secondQueuedProcesses)
	}

	testingTaskManager.RemoveRequest("test", testingTaskManager.AllTasks.Tasks["test1"])
	testingTaskManager.RemoveRequest("test", testingTaskManager.AllTasks.Tasks["test2"])
	totalRequests := 10

	for i := 0; i < totalRequests; i++ {
		testingTaskManager.AddRequest(fmt.Sprintf("test%d", i), true)
	}

	firstQueuedProcesses = len(testingTaskManager.AllTasks.Tasks["test1"].QueuedProcesses)
	secondQueuedProcesses = len(testingTaskManager.AllTasks.Tasks["test2"].QueuedProcesses)

	if (firstQueuedProcesses + secondQueuedProcesses) != totalRequests {
		t.Error("Task not added correctly (10)", firstQueuedProcesses, secondQueuedProcesses)
	}
	if firstQueuedProcesses == totalRequests || secondQueuedProcesses == totalRequests {
		t.Error("Tasks not distributed equally (10)", firstQueuedProcesses, secondQueuedProcesses)
	}

	for i := 0; i < totalRequests; i++ {
		testingTaskManager.RemoveRequest(fmt.Sprintf("test%d", i), testingTaskManager.AllTasks.Tasks["test1"])
	}

}

func TestTaskQueueUnloaded(t *testing.T) {
	TestNewTaskManager(t)
	for i := 0; i < 10; i++ {
		testingTaskManager.AllTasks.Tasks["test1"].QueuedProcesses = append(testingTaskManager.AllTasks.Tasks["test1"].QueuedProcesses, &QueuedProcess{
			Prompt:      fmt.Sprintf("test%d", i),
			Done:        make(chan ResponseChan),
			TimeStarted: time.Now(),
		})

	}
	testingTaskManager.TaskQueueUnloaded(testingTaskManager.AllTasks.Tasks["test2"], true)

	firstQueuedProcesses := len(testingTaskManager.AllTasks.Tasks["test1"].QueuedProcesses)
	secondQueuedProcesses := len(testingTaskManager.AllTasks.Tasks["test2"].QueuedProcesses)

	if (firstQueuedProcesses) == 10 {
		t.Error("All tasks stuck in 1st queue", firstQueuedProcesses, secondQueuedProcesses)
	}

	if (firstQueuedProcesses + secondQueuedProcesses) != 10 {
		t.Error("Tasks not adding upto 10", firstQueuedProcesses, secondQueuedProcesses)
	}

}

func TestPingProcessor(t *testing.T) {
	TestNewTaskManager(t)
	testingTaskManager.AllTasks.Tasks["test1"].IsProcessing = true
	testingTaskManager.AllTasks.Tasks["test2"].IsProcessing = true

	for i := 0; i < 10; i++ {
		testingTaskManager.AddRequest(fmt.Sprintf("test%d", i), true)
	}

	if !testingTaskManager.PingProcessor("test1", true) {
		t.Error("PingProcessor is not able to handle processing tasks")
	}

	if !testingTaskManager.PingProcessor("test2", true) {
		t.Error("PingProcessor is not able to handle processing tasks")
	}

}
