package tasks

import (
	"sync"
	"time"
)

type QueuedProcess struct {
	Prompt      string
	Done        chan ResponseChan
	Retries     int
	TimeStarted time.Time
}

type TaskObject struct {
	ApiKey          string
	QueuedProcesses []*QueuedProcess
	TaskMu          sync.Mutex
	IsOverloaded    bool
	DisplayName     string
	IsProcessing    bool
}

type TaskObjects struct {
	Tasks   map[string]*TaskObject
	TasksMu sync.RWMutex
}

type PendingTaskObjects struct {
	PendingMu    sync.RWMutex
	PendingQueue []*QueuedProcess
}

type TaskManager struct {
	AllTasks     *TaskObjects
	PendingTasks *PendingTaskObjects
	IsTesting    bool
}

type ResponseChan struct {
	Response string
}

type JsonKeys struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}
