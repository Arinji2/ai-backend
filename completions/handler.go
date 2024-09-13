package completions

import (
	"net/http"
	"strings"

	"github.com/Arinji2/ai-backend/input"
	custom_log "github.com/Arinji2/ai-backend/logger"
	"github.com/Arinji2/ai-backend/tasks"
)

func CompletionsHandler(w http.ResponseWriter, r *http.Request) {

	prompt, err := input.CompletionsBodyInput(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	promptParts := strings.Split(prompt, ":")
	taskManager := tasks.GetTaskManager()
	custom_log.Logger.Debug("NEW REQUEST RECEIVED", promptParts[1])

	response := <-taskManager.AddRequest(promptParts[0])
	custom_log.Logger.Debug("SENT BACK RESPONSE", promptParts[1])
	w.Write([]byte(response.Response))

}
