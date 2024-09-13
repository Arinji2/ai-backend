package tasks

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	custom_log "github.com/Arinji2/ai-backend/logger"
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

type JsonKeys struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

const TaskDistributionThreshold = 3 //The amount of tasks to distribute among moderately filled task queues
const MaxRetries = 3                //The amount of times to retry a failed task

var (
	taskManagerInstance *TaskManager
	once                sync.Once
)

func SetupTasks() (*TaskObjects, *PendingTaskObjects) {
	jsonFile, err := os.Open("./keys.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var keys []JsonKeys
	err = json.Unmarshal(byteValue, &keys)
	if err != nil {
		panic(err)
	}

	tasks := &TaskObjects{
		Tasks: make(map[string]*TaskObject),
	}
	for _, keyData := range keys {
		trimmedKey := strings.TrimSpace(keyData.Key)
		taskObject := &TaskObject{
			ApiKey: trimmedKey,

			QueuedProcesses: make([]*QueuedProcess, 0),
		}
		tasks.Tasks[trimmedKey] = taskObject
		custom_log.Logger.Info("Loaded Task Key:", keyData.Name)
	}

	pendingTasks := &PendingTaskObjects{
		PendingTasks: make([]*TaskObject, 0),
	}
	return tasks, pendingTasks
}
