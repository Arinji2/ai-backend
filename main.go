package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/Arinji2/ai-backend/completions"
	custom_log "github.com/Arinji2/ai-backend/logger"
	"github.com/Arinji2/ai-backend/tasks"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	} else {
		custom_log.Logger.Warn("Environment File Found")
	}
	r := chi.NewRouter()
	r.Use(SkipLoggingMiddleware)
	r.Use(CheckAccessKeyMiddleware)

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

			if len(taskManager.PendingTasks.PendingQueue) > 0 {
				custom_log.Logger.Debug("Pending Tasks: ", len(taskManager.PendingTasks.PendingQueue))
			}

		}
	}()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
		log.Println("Running Profiler on localhost:6060")
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

func CheckAccessKeyMiddleware(next http.Handler) http.Handler {
	isProduction := os.Getenv("ENVIRONMENT") == "PRODUCTION"
	if !isProduction {
		return next
	}
	accessKey := os.Getenv("ACCESS_KEY")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		inputAccessKey := r.Header.Get("Authorization")
		if inputAccessKey != accessKey {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}

		middleware.Logger(next).ServeHTTP(w, r)
	})
}
