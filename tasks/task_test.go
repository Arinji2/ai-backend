package tasks

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestUpdateOverloaded(t *testing.T) {
	TestNewTaskManager(t)
	for i := 0; i < 4; i++ {
		testingTaskManager.AllTasks.Tasks["test1"].QueuedProcesses = append(testingTaskManager.AllTasks.Tasks["test1"].QueuedProcesses, &QueuedProcess{
			Prompt:      fmt.Sprintf("test%d in test1", i),
			Done:        make(chan ResponseChan),
			TimeStarted: time.Now(),
		})
	}

	for i := 0; i < 4; i++ {
		testingTaskManager.AllTasks.Tasks["test2"].QueuedProcesses = append(testingTaskManager.AllTasks.Tasks["test2"].QueuedProcesses, &QueuedProcess{
			Prompt:      fmt.Sprintf("test%d in test2", i),
			Done:        make(chan ResponseChan),
			TimeStarted: time.Now(),
		})
	}

	queue := testingTaskManager.AllTasks.Tasks["test1"].QueuedProcesses[0]
	testingTaskManager.AllTasks.Tasks["test1"].UpdateOverloaded(queue, true)

	secondQueuedProcesses := len(testingTaskManager.AllTasks.Tasks["test2"].QueuedProcesses)
	if secondQueuedProcesses != 8 {
		t.Error("Tasks not adding up to 8", secondQueuedProcesses)
	}
	fmt.Println(secondQueuedProcesses)

	tries := make(chan struct{}, 7)

	var wg sync.WaitGroup
	wg.Add(2)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			if testingTaskManager.AllTasks.Tasks["test1"].IsOverloaded {
				tries <- struct{}{}
				wg.Done()
			} else {
				wg.Done()
			}

			if len(tries) >= 2 {
				break
			}
		}
	}()

	wg.Wait()
	close(tries)

	if len(tries) > 2 {
		t.Error("Task took more than 2 tries", len(tries))
	}
	firstQueuedProcesses := len(testingTaskManager.AllTasks.Tasks["test1"].QueuedProcesses)
	secondQueuedProcesses = len(testingTaskManager.AllTasks.Tasks["test2"].QueuedProcesses)

	fmt.Println(firstQueuedProcesses, secondQueuedProcesses)

	if secondQueuedProcesses == 8 {
		t.Error("Tasks not distributing correctly")
	}

	if firstQueuedProcesses+secondQueuedProcesses != 8 {
		t.Error("Tasks not adding up to 8", firstQueuedProcesses, secondQueuedProcesses)
	}
}
