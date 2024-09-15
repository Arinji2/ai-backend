package completions

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Arinji2/ai-backend/input"
	custom_log "github.com/Arinji2/ai-backend/logger"
	"github.com/Arinji2/ai-backend/tasks"
)

func writeResponse(w http.ResponseWriter, response string) {
	responseStruct := struct {
		Message string `json:"message"`
	}{
		Message: response,
	}
	jsonResponse, err := json.Marshal(responseStruct)
	if err != nil {
		http.Error(w, "Unable to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
func CompletionsHandler(w http.ResponseWriter, r *http.Request) {

	prompt, err := input.CompletionsBodyInput(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	taskManager := tasks.GetTaskManager()
	custom_log.Logger.Debug("NEW REQUEST RECEIVED")

	response := <-taskManager.AddRequest(prompt)
	custom_log.Logger.Debug("SENT BACK RESPONSE")

	if strings.Contains(response.Response, "Error Processing Task") {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(response.Response))
		return
	}
	w.WriteHeader(http.StatusOK)
	writeResponse(w, response.Response)

}
