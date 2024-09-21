package tasks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	custom_log "github.com/Arinji2/ai-backend/logger"
)

var (
	taskManagerInstance *TaskManager
	once                sync.Once
)

func SetupTasks() (*TaskObjects, *PendingTaskObjects) {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	keyDir := ""

	if os.Getenv("ENVIRONMENT") == "PRODUCTION" {
		keyDir = "/keys.json"
	} else {
		keyDir = fmt.Sprintf("%s/keys.json", dir)
	}

	jsonFile, err := os.Open(keyDir)
	if err != nil {
		panic(err)
	}

	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
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
			ApiKey:          trimmedKey,
			DisplayName:     keyData.Name,
			QueuedProcesses: make([]*QueuedProcess, 0),
		}
		tasks.Tasks[trimmedKey] = taskObject
		custom_log.Logger.Info("Loaded Task Key:", keyData.Name)
	}

	pendingTasks := &PendingTaskObjects{
		PendingMu:    sync.RWMutex{},
		PendingQueue: make([]*QueuedProcess, 0),
	}
	return tasks, pendingTasks
}
