package tasks

import (
	"fmt"
	"testing"
)

var testingTaskManager *TaskManager

func TestNewTaskManager(t *testing.T) {
	optionalKeys := []JsonKeys{
		{
			Name: "test",
			Key:  "test",
		},
		{
			Name: "test2",
			Key:  "test2",
		},
	}
	taskManager := NewTaskManager(optionalKeys)
	if taskManager.AllTasks.Tasks["test"].DisplayName != "test" {
		t.Error("Task Manager not initialized correctly")
	}
	if taskManager.AllTasks.Tasks["test2"].DisplayName != "test2" {
		t.Error("Task Manager not initialized correctly")
	}

	testingTaskManager = taskManager

}

func TestAddRequest(t *testing.T) {
	TestNewTaskManager(t)
	testingTaskManager.AddRequest("test", true)

	if len(testingTaskManager.AllTasks.Tasks["test"].QueuedProcesses) != 1 {
		t.Error("Task not added correctly")
	}

	testingTaskManager.RemoveRequest("test", testingTaskManager.AllTasks.Tasks["test"])
	totalRequests := 10

	for i := 0; i < totalRequests; i++ {
		testingTaskManager.AddRequest(fmt.Sprintf("test%d", i), true)
	}

	firstQueuedProcesses := len(testingTaskManager.AllTasks.Tasks["test"].QueuedProcesses)
	secondQueuedProcesses := len(testingTaskManager.AllTasks.Tasks["test2"].QueuedProcesses)

	if (firstQueuedProcesses + secondQueuedProcesses) != totalRequests {
		t.Error("Task not added correctly")
	}

	if firstQueuedProcesses == totalRequests || secondQueuedProcesses == totalRequests {
		t.Error("Tasks not distributed equally", firstQueuedProcesses, secondQueuedProcesses)
	}

	for i := 0; i < totalRequests; i++ {
		testingTaskManager.RemoveRequest(fmt.Sprintf("test%d", i), testingTaskManager.AllTasks.Tasks["test"])
	}

}
