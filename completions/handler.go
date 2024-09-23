package completions

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

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

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, time.Hour*1)
	defer cancel()

	r = r.WithContext(ctx)

	promptCh := make(chan struct {
		prompt string
		err    error
	}, 1)

	go func() {

		prompt, err := input.CompletionsBodyInput(r)
		promptCh <- struct {
			prompt string
			err    error
		}{prompt, err}
	}()

	select {
	case result := <-promptCh:
		if result.err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(result.err.Error()))
			return
		}

		prompt := result.prompt
		taskManager := tasks.GetTaskManager()
		custom_log.Logger.Debug("NEW REQUEST RECEIVED")

		taskCh := make(chan struct {
			response string
			err      error
		}, 1)

		go func() {
			response := <-taskManager.AddRequest(prompt)
			taskCh <- struct {
				response string
				err      error
			}{response.Response, nil}
		}()

		select {
		case taskResult := <-taskCh:

			if strings.Contains(taskResult.response, "Error Processing Task") {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(taskResult.response))
				return
			}
			w.WriteHeader(http.StatusOK)
			writeResponse(w, taskResult.response)

		case <-ctx.Done():
			custom_log.Logger.Warn("Task Processing Timed Out")
			w.WriteHeader(http.StatusGatewayTimeout)
			w.Write([]byte("Task processing timed out"))
		}

	case <-ctx.Done():

		custom_log.Logger.Warn("Request Processing Timed Out")
		w.WriteHeader(http.StatusGatewayTimeout)

		w.Write([]byte("Request timed out"))
	}
}
