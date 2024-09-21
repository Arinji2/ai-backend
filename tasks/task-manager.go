package tasks

import (
	"time"
)

func GetTaskManager() *TaskManager {
	once.Do(func() {
		taskManagerInstance = NewTaskManager(nil)
	})
	return taskManagerInstance
}

func NewTaskManager(optionalKeys []JsonKeys) *TaskManager {
	tasks, pendingTasks := SetupTasks(optionalKeys)
	taskManager := &TaskManager{AllTasks: tasks, PendingTasks: pendingTasks}

	return taskManager
}

func (tm *TaskManager) AddRequest(prompt string, testingMode bool) chan ResponseChan {
	return tm.addRequestInternal(prompt, nil, time.Now(), testingMode)
}

func (tm *TaskManager) MoveAddedRequest(prompt string, done chan ResponseChan, testingMode bool) {

	tm.addRequestInternal(prompt, done, time.Time{}, testingMode)

}

func (tm *TaskManager) addRequestInternal(prompt string, done chan ResponseChan, initialTime time.Time, testingMode bool) chan ResponseChan {

	tm.AllTasks.TasksMu.RLock()
	defer func() {
		tm.AllTasks.TasksMu.RUnlock()
	}()

	var leastBusyTask *TaskObject
	minQueueLength := -1
	taskAdded := false

	if done == nil {
		done = make(chan ResponseChan)
	}

	if initialTime.IsZero() {
		initialTime = time.Now()
	}

	for _, task := range tm.AllTasks.Tasks {
		if task.IsOverloaded {
			continue
		}
		task.TaskMu.Lock()

		if len(task.QueuedProcesses) == 0 && !task.IsOverloaded {

			task.QueuedProcesses = append(task.QueuedProcesses, &QueuedProcess{Prompt: prompt, Done: done, TimeStarted: initialTime})
			task.TaskMu.Unlock()
			if !testingMode {
				go taskManagerInstance.PingProcessor(task.ApiKey)
			}
			taskAdded = true
			return done
		}
		if minQueueLength == -1 || len(task.QueuedProcesses) < minQueueLength {
			leastBusyTask = task
			minQueueLength = len(task.QueuedProcesses)
		}
		task.TaskMu.Unlock()
	}

	if leastBusyTask != nil {

		leastBusyTask.TaskMu.Lock()
		leastBusyTask.QueuedProcesses = append(leastBusyTask.QueuedProcesses, &QueuedProcess{Prompt: prompt, Done: done, TimeStarted: initialTime})
		leastBusyTask.TaskMu.Unlock()
		if !testingMode {
			go taskManagerInstance.PingProcessor(leastBusyTask.ApiKey)
		}
		taskAdded = true
	}

	if !taskAdded {

		taskManagerInstance.PendingTasks.PendingMu.Lock()

		taskManagerInstance.PendingTasks.PendingQueue = append(taskManagerInstance.PendingTasks.PendingQueue, &QueuedProcess{
			Prompt:      prompt,
			Done:        done,
			TimeStarted: initialTime,
		})
		taskManagerInstance.PendingTasks.PendingMu.Unlock()

	}

	return done
}

func (tm *TaskManager) CheckPendingTasks(task *TaskObject) {

	tm.PendingTasks.PendingMu.Lock()
	defer tm.PendingTasks.PendingMu.Unlock()

	if len(tm.PendingTasks.PendingQueue) == 0 {
		return
	}

	pendingTasks := make([]*QueuedProcess, len(tm.PendingTasks.PendingQueue))
	copy(pendingTasks, tm.PendingTasks.PendingQueue)
	tm.PendingTasks.PendingQueue = nil

	tm.PendingTasks.PendingMu.Unlock()

	for _, pendingTask := range pendingTasks {
		tm.addRequestInternal(pendingTask.Prompt, pendingTask.Done, pendingTask.TimeStarted, false)
	}

	tm.PendingTasks.PendingMu.Lock()
}
func (tm *TaskManager) RemoveRequest(prompt string, task *TaskObject) {
	task.TaskMu.Lock()
	defer task.TaskMu.Unlock()

	for i, queuedProcess := range task.QueuedProcesses {
		if queuedProcess.Prompt == prompt {
			task.QueuedProcesses = append(task.QueuedProcesses[:i], task.QueuedProcesses[i+1:]...)

			return
		}
	}

}

func (tm *TaskManager) TaskQueueUnloaded(task *TaskObject, testingMode bool) {
	tm.AllTasks.TasksMu.RLock()
	defer tm.AllTasks.TasksMu.RUnlock()

	var largestQueue *TaskObject
	var largestQueueLength int = -1
	for _, t := range tm.AllTasks.Tasks {
		t.TaskMu.Lock()

		if len(t.QueuedProcesses) > largestQueueLength {
			largestQueue = t
			largestQueueLength = len(t.QueuedProcesses)
		}
		t.TaskMu.Unlock()

	}

	if largestQueue != nil && largestQueueLength > 0 {

		largestQueue.TaskMu.Lock()

		numToMove := largestQueueLength / 2
		tasksToMove := make([]*QueuedProcess, numToMove)

		for i := 0; i < numToMove; i++ {
			tasksToMove[i] = largestQueue.QueuedProcesses[i]
		}

		largestQueue.QueuedProcesses = largestQueue.QueuedProcesses[numToMove:]

		largestQueue.TaskMu.Unlock()

		for _, task := range tasksToMove {
			tm.MoveAddedRequest(task.Prompt, task.Done, testingMode)
		}

	}
	for _, taskQueue := range tm.AllTasks.Tasks {
		if len(taskQueue.QueuedProcesses) > 0 && taskQueue != largestQueue {
			taskQueue.TaskMu.Lock()

			var taskToMove *QueuedProcess
			shouldMove := false

			if len(taskQueue.QueuedProcesses) > TaskDistributionThreshold {
				taskToMove = taskQueue.QueuedProcesses[0]
				taskQueue.QueuedProcesses = taskQueue.QueuedProcesses[1:]
				shouldMove = true
			}

			taskQueue.TaskMu.Unlock()

			if shouldMove {
				tm.MoveAddedRequest(taskToMove.Prompt, taskToMove.Done, testingMode)
			}
		}
	}
	task.TaskMu.Lock()
	queueEmpty := len(task.QueuedProcesses) == 0
	task.TaskMu.Unlock()

	if queueEmpty {
		go taskManagerInstance.CheckPendingTasks(task)
	}

}

func (tm *TaskManager) PingProcessor(key string) {

	task := tm.AllTasks.Tasks[key]
	if task.IsProcessing {
		return
	}
	go task.ProcessTasks()
}
