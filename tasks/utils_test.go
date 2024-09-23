package tasks

import "testing"

func AssignTaskAndQueue(t *testing.T, task *TaskObject) (*TaskObject, int) {
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
