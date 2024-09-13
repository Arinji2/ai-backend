package tasks

import "sync"

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

type JsonKeys struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}
