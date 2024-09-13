package tasks

import (
	"encoding/json"
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

func SetupTasks(keys []JsonKeys) (*TaskObjects, *PendingTaskObjects) {
	if keys == nil {
		jsonFile, err := os.Open("./keys.json")
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
