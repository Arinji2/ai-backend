package tasks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var (
	taskManagerInstance *TaskManager
	once                sync.Once
)

func SetupTasks(optionalKeys []JsonKeys) (*TaskObjects, *PendingTaskObjects) {
	var keys []JsonKeys
	if optionalKeys == nil {
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

		err = json.Unmarshal(byteValue, &keys)
		if err != nil {
			panic(err)
		}
	} else {
		keys = optionalKeys
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
		if optionalKeys == nil {
			fmt.Println("Loaded Task Key:", keyData.Name)
		}
	}

	pendingTasks := &PendingTaskObjects{
		PendingMu:    sync.RWMutex{},
		PendingQueue: make([]*QueuedProcess, 0),
	}
	return tasks, pendingTasks
}
