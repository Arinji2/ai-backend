package completions

import (
	"net/http"

	"github.com/Arinji2/ai-backend/input"
	"github.com/Arinji2/ai-backend/tasks"
)

func CompletionsHandler(w http.ResponseWriter, r *http.Request) {

	prompt, err := input.CompletionsBodyInput(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	taskManager := tasks.GetTaskManager()
	taskManager.AddRequest(prompt)

}
