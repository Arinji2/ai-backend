package tasks

import (
	"testing"
)

func TestSetupTasks(t *testing.T) {

	keys := []JsonKeys{
		{
			Name: "Testing Key 1",
			Key:  "test-key-1",
		},
		{
			Name: "Testing Key 2",
			Key:  "test-key-2",
		},
		{
			Name: "Testing Key 3",
			Key:  "test-key-3",
		},
	}

	tasks, pendingTasks := SetupTasks(keys)
	if len(tasks.Tasks) != len(keys) {
		t.Errorf("Expected %d tasks, got %d", len(keys), len(tasks.Tasks))
	}
	if len(pendingTasks.PendingTasks) != 0 {
		t.Errorf("Expected 0 pending tasks, got %d", len(pendingTasks.PendingTasks))
	}

	for _, key := range keys {
		if _, ok := tasks.Tasks[key.Key]; !ok {
			t.Errorf("Expected task with key %s to exist", key.Key)
		}
	}
}
