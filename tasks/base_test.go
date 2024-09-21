package tasks

import "testing"

func TestSetupTasks(t *testing.T) {
	optionalKeys := []JsonKeys{
		{
			Name: "test",
			Key:  "test",
		},
		{
			Name: "test2",
			Key:  "test2",
		},
	}

	tasks, _ := SetupTasks(optionalKeys)
	if len(tasks.Tasks) != 2 {
		t.Error("Tasks not initialized correctly")
	}

}
