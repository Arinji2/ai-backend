package completions

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Arinji2/ai-backend/input"
	custom_log "github.com/Arinji2/ai-backend/logger"
	"github.com/Arinji2/ai-backend/tasks"
)

func writeResponse(w http.ResponseWriter, response string) {
	// Define the struct with a message field
	responseStruct := struct {
		Message string `json:"message"`
	}{
		Message: response,
	}

	// Marshal the struct into JSON
	jsonResponse, err := json.Marshal(responseStruct)
	if err != nil {
		http.Error(w, "Unable to marshal response", http.StatusInternalServerError)
		return
	}

	// Set Content-Type to application/json and write the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
func CompletionsHandler(w http.ResponseWriter, r *http.Request) {

	prompt, err := input.CompletionsBodyInput(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	taskManager := tasks.GetTaskManager()
	custom_log.Logger.Debug("NEW REQUEST RECEIVED")

	response := <-taskManager.AddRequest(prompt)
	custom_log.Logger.Debug("SENT BACK RESPONSE")
	writeResponse(w, response.Response)

}
