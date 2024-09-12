package tasks

import (
	"time"

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
	go taskManager.ProcessTasks()
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

	return done
}

func (tm *TaskManager) ProcessTasks() {
	ticker := time.NewTicker(time.Second * 1)
	custom_log.Logger.Info("Starting Task Processor")
	for range ticker.C {
		go func() {
			tm.AllTasks.TasksMu.RLock()
			for _, task := range tm.AllTasks.Tasks {
				task.ProcessTasks()
			}
			tm.AllTasks.TasksMu.RUnlock()
		}()
	}
}
