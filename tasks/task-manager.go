package tasks

import (
	custom_log "github.com/Arinji2/ai-backend/logger"
)

func GetTaskManager() *TaskManager {
	once.Do(func() {
		taskManagerInstance = NewTaskManager()
	})
	return taskManagerInstance
}

func NewTaskManager() *TaskManager {
	tasks := SetupTasks()
	taskManager := &TaskManager{AllTasks: tasks}

	return taskManager
}

func (tm *TaskManager) AddRequest(prompt string) chan struct{} {
	tm.AllTasks.TasksMu.RLock()
	defer tm.AllTasks.TasksMu.RUnlock()

	var leastBusyTask *TaskObject
	minQueueLength := -1
	done := make(chan struct{})

	for _, task := range tm.AllTasks.Tasks {
		task.TaskMu.Lock()
		if !task.IsProcessing && len(task.QueuedProcesses) == 0 {
			task.QueuedProcesses = append(task.QueuedProcesses, &QueuedProcess{Prompt: prompt, Done: done})
			task.TaskMu.Unlock()
			taskManagerInstance.PingProcessor(task.ApiKey)
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
		leastBusyTask.QueuedProcesses = append(leastBusyTask.QueuedProcesses, &QueuedProcess{Prompt: prompt, Done: done})
		leastBusyTask.TaskMu.Unlock()
	}

	taskManagerInstance.PingProcessor(leastBusyTask.ApiKey)

	return done
}

func (tm *TaskManager) PingProcessor(key string) {
	custom_log.Logger.Debug("Pinging Processor: ", key)
	tm.AllTasks.TasksMu.RLock()
	defer tm.AllTasks.TasksMu.RUnlock()
	task := tm.AllTasks.Tasks[key]
	task.ProcessTasks()
}
