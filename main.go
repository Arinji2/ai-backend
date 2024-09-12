package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Arinji2/ai-backend/completions"
	custom_log "github.com/Arinji2/ai-backend/logger"
	"github.com/Arinji2/ai-backend/tasks"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/joho/godotenv"
)

func main() {
	r := chi.NewRouter()
	r.Use(SkipLoggingMiddleware)

	err := godotenv.Load()
	if err != nil {
		isProduction := os.Getenv("ENVIRONMENT") == "PRODUCTION"
		if !isProduction {
			log.Fatal("Error loading .env file")
		} else {
			custom_log.Logger.Warn("Using Production Environment")
		}
	} else {
		custom_log.Logger.Warn("Using Development Environment")
	}

	r.Get("/", healthHandler)
	r.Get("/health", healthCheckHandler)
	r.Post("/completions", completions.CompletionsHandler)

	taskManager := tasks.GetTaskManager()
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for range ticker.C {

			for key, tasks := range taskManager.AllTasks.Tasks {
				if len(tasks.QueuedProcesses) == 0 {
					continue
				}
				custom_log.Logger.Debug("Tasks In Queue: ", key, len(tasks.QueuedProcesses))
			}

		}
	}()

	http.ListenAndServe(":8080", r)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Vibeify Backend: Request Received")
	w.Write([]byte("Vibeify Backend: Request Received"))
	render.Status(r, http.StatusOK)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Vibeify Backend: Health Check"))
	render.Status(r, http.StatusOK)
}

func SkipLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		middleware.Logger(next).ServeHTTP(w, r)
	})
}
