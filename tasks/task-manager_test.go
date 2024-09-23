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
	taskQueueOne, firstQueuedProcesses := AssignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test1"])
	taskQueueTwo, secondQueuedProcesses := AssignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test2"])

	if (firstQueuedProcesses + secondQueuedProcesses) != 1 {
		t.Error("Task not added correctly (1)", firstQueuedProcesses, secondQueuedProcesses)
	}

	taskManagerInstance.RemoveRequest("test", taskQueueOne)
	taskManagerInstance.RemoveRequest("test", taskQueueTwo)
	totalRequests := 10

	MockAddingRequests(t, (totalRequests / 2), taskQueueOne)
	MockAddingRequests(t, (totalRequests / 2), taskQueueTwo)

	_, firstQueuedProcesses = AssignTaskAndQueue(t, taskQueueOne)
	_, secondQueuedProcesses = AssignTaskAndQueue(t, taskQueueTwo)

	if (firstQueuedProcesses + secondQueuedProcesses) != totalRequests {
		t.Errorf("Task not added correctly. Total: (%d). First: (%d). Second: (%d)", totalRequests, firstQueuedProcesses, secondQueuedProcesses)
	}
	if firstQueuedProcesses == totalRequests || secondQueuedProcesses == totalRequests {
		t.Errorf("Tasks not distributed equally. Total: (%d). First: (%d). Second: (%d)", totalRequests, firstQueuedProcesses, secondQueuedProcesses)
	}

	for i := 0; i < totalRequests; i++ {
		taskManagerInstance.RemoveRequest(fmt.Sprintf("test%d", i), taskQueueOne)
	}

}

func TestTaskQueueUnloaded(t *testing.T) {
	TestNewTaskManager(t)

	taskQueueOne, _ := AssignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test1"])
	taskQueueTwo, _ := AssignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test2"])

	totalRequests := 10

	MockAddingRequests(t, totalRequests, taskQueueTwo)
	taskManagerInstance.TaskQueueUnloaded(taskQueueTwo)

	_, firstQueuedProcesses := AssignTaskAndQueue(t, taskQueueOne)
	_, secondQueuedProcesses := AssignTaskAndQueue(t, taskQueueTwo)

	if (firstQueuedProcesses) == totalRequests {
		t.Error("All tasks stuck in 1st queue", firstQueuedProcesses, secondQueuedProcesses)
	}

	if (firstQueuedProcesses + secondQueuedProcesses) != totalRequests {
		t.Errorf("Tasks not adding upto %d. First: %d, Second: %d", totalRequests, firstQueuedProcesses, secondQueuedProcesses)
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
