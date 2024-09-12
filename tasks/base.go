package tasks

import (
	"os"
	"strings"
	"sync"
)

type QueuedProcess struct {
	Prompt string
	Done   chan struct{}
}

type TaskObject struct {
	ApiKey          string
	IsProcessing    bool
	QueuedProcesses []*QueuedProcess
	TaskMu          sync.Mutex
}

type TaskObjects struct {
	Tasks   map[string]*TaskObject
	TasksMu sync.RWMutex
}

type TaskManager struct {
	AllTasks *TaskObjects
}

var (
	taskManagerInstance *TaskManager
	once                sync.Once
)

func SetupTasks() *TaskObjects {
	apiKeys := os.Getenv("API_KEY")
	allKeys := strings.Split(apiKeys, ",")
	tasks := &TaskObjects{
		Tasks: make(map[string]*TaskObject),
	}
	for _, key := range allKeys {
		trimmedKey := strings.TrimSpace(key)
		taskObject := &TaskObject{
			ApiKey:          trimmedKey,
			IsProcessing:    false,
			QueuedProcesses: make([]*QueuedProcess, 0),
		}
		tasks.Tasks[trimmedKey] = taskObject
	}
	return tasks
}
