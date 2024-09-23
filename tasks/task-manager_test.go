package tasks

import (
	"fmt"
	"testing"
	"time"
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

	for i := 0; i < totalRequests; i++ {
		taskManagerInstance.AddRequest(fmt.Sprintf("test%d", i))
	}

	_, firstQueuedProcesses = AssignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test1"])
	_, secondQueuedProcesses = AssignTaskAndQueue(t, taskManagerInstance.AllTasks.Tasks["test2"])

	if (firstQueuedProcesses + secondQueuedProcesses) != totalRequests {
		t.Error("Task not added correctly (10)", firstQueuedProcesses, secondQueuedProcesses)
	}
	if firstQueuedProcesses == totalRequests || secondQueuedProcesses == totalRequests {
		t.Error("Tasks not distributed equally (10)", firstQueuedProcesses, secondQueuedProcesses)
	}

	for i := 0; i < totalRequests; i++ {
		taskManagerInstance.RemoveRequest(fmt.Sprintf("test%d", i), taskQueueOne)
	}

}

func TestTaskQueueUnloaded(t *testing.T) {
	TestNewTaskManager(t)
	for i := 0; i < 10; i++ {
		taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses = append(taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses, &QueuedProcess{
			Prompt:      fmt.Sprintf("test%d", i),
			Done:        make(chan ResponseChan),
			TimeStarted: time.Now(),
		})

	}
	taskManagerInstance.TaskQueueUnloaded(taskManagerInstance.AllTasks.Tasks["test2"])

	firstQueuedProcesses := len(taskManagerInstance.AllTasks.Tasks["test1"].QueuedProcesses)
	secondQueuedProcesses := len(taskManagerInstance.AllTasks.Tasks["test2"].QueuedProcesses)

	if (firstQueuedProcesses) == 10 {
		t.Error("All tasks stuck in 1st queue", firstQueuedProcesses, secondQueuedProcesses)
	}

	if (firstQueuedProcesses + secondQueuedProcesses) != 10 {
		t.Error("Tasks not adding upto 10", firstQueuedProcesses, secondQueuedProcesses)
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
