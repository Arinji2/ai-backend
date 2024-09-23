package tasks

import (
	"fmt"
	"testing"
)

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

	taskManagerInstance = taskManager

}

func TestAddRequest(t *testing.T) {
	TestNewTaskManager(t)
	taskManagerInstance.AddRequest("test")
	taskQueueOne, firstQueuedProcesses := assignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test1"])
	taskQueueTwo, secondQueuedProcesses := assignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test2"])

	totalRequests := 1

	if (firstQueuedProcesses + secondQueuedProcesses) != totalRequests {
		testLoggingHelper(t, fmt.Sprintf("Tasks not adding upto (%d)", totalRequests), true)
	}

	taskManagerInstance.RemoveRequest("test", taskQueueOne)
	taskManagerInstance.RemoveRequest("test", taskQueueTwo)
	totalRequests = 10

	mockAddingRequests(t, (totalRequests / 2), taskQueueOne)
	mockAddingRequests(t, (totalRequests / 2), taskQueueTwo)

	_, firstQueuedProcesses = assignTaskAndQueue(t, taskQueueOne)
	_, secondQueuedProcesses = assignTaskAndQueue(t, taskQueueTwo)

	if (firstQueuedProcesses + secondQueuedProcesses) != totalRequests {
		testLoggingHelper(t, fmt.Sprintf("Tasks not adding upto (%d)", totalRequests), true)
	}
	if firstQueuedProcesses == totalRequests || secondQueuedProcesses == totalRequests {
		testLoggingHelper(t, "Tasks not distributed equally", true)
	}

	for i := 0; i < totalRequests; i++ {
		taskManagerInstance.RemoveRequest(fmt.Sprintf("test%d", i), taskQueueOne)
	}

}

func TestTaskQueueUnloaded(t *testing.T) {
	TestNewTaskManager(t)

	taskQueueOne, _ := assignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test1"])
	taskQueueTwo, _ := assignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test2"])

	totalRequests := 10

	mockAddingRequests(t, totalRequests, taskQueueTwo)
	taskManagerInstance.TaskQueueUnloaded(taskQueueTwo)

	_, firstQueuedProcesses := assignTaskAndQueue(t, taskQueueOne)
	_, secondQueuedProcesses := assignTaskAndQueue(t, taskQueueTwo)

	if (firstQueuedProcesses) == totalRequests {
		testLoggingHelper(t, fmt.Sprintf("All tasks stuck in %s queue", taskQueueOne.DisplayName), true)
	}

	if (firstQueuedProcesses + secondQueuedProcesses) != totalRequests {

		testLoggingHelper(t, fmt.Sprintf("Tasks not adding upto %d", totalRequests), true)
	}

}

func TestPingProcessor(t *testing.T) {
	TestNewTaskManager(t)
	taskManagerInstance.AllTasks.Tasks["test1"].IsProcessing = true
	taskManagerInstance.AllTasks.Tasks["test2"].IsProcessing = true

	for i := 0; i < 10; i++ {
		taskManagerInstance.AddRequest(fmt.Sprintf("test%d", i))
	}

	if !taskManagerInstance.PingProcessor("test1") {
		t.Error("PingProcessor is not able to handle processing tasks")
	}

	if !taskManagerInstance.PingProcessor("test2") {
		t.Error("PingProcessor is not able to handle processing tasks")
	}

}
