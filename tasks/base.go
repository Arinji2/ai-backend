package tasks

import (
	"os"
	"strings"
	"sync"
)

type QueuedProcess struct {
	Prompt  string
	Done    chan ResponseChan
	Retries int
}

type TaskObject struct {
	ApiKey          string
	QueuedProcesses []*QueuedProcess
	TaskMu          sync.Mutex
	IsOverloaded    bool
}

type TaskObjects struct {
	Tasks   map[string]*TaskObject
	TasksMu sync.RWMutex
}

type PendingTaskObjects struct {
	PendingTasks []*TaskObject
	PendingMu    sync.RWMutex
}

type TaskManager struct {
	AllTasks     *TaskObjects
	PendingTasks *PendingTaskObjects
}

type ResponseChan struct {
	Response string
}

const TaskDistributionThreshold = 3 //The amount of tasks to distribute among moderately filled task queues
const MaxRetries = 3

var (
	taskManagerInstance *TaskManager
	once                sync.Once
)

func SetupTasks() (*TaskObjects, *PendingTaskObjects) {
	apiKeys := os.Getenv("API_KEY")
	allKeys := strings.Split(apiKeys, ",")
	tasks := &TaskObjects{
		Tasks: make(map[string]*TaskObject),
	}
	for _, key := range allKeys {
		trimmedKey := strings.TrimSpace(key)
		taskObject := &TaskObject{
			ApiKey: trimmedKey,

			QueuedProcesses: make([]*QueuedProcess, 0),
		}
		tasks.Tasks[trimmedKey] = taskObject
	}

	pendingTasks := &PendingTaskObjects{
		PendingTasks: make([]*TaskObject, 0),
	}
	return tasks, pendingTasks
}
