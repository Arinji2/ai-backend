package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/Arinji2/ai-backend/completions"
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
		if os.Getenv("ENVIRONMENT") == "PRODUCTION" {
			fmt.Println("Production Environment")
		} else {
			fmt.Println("Environment File Found")
		}
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

			for _, tasks := range taskManager.AllTasks.Tasks {
				if len(tasks.QueuedProcesses) == 0 {
					continue
				}
				fmt.Println("Tasks In Queue: ", tasks.DisplayName, len(tasks.QueuedProcesses))
			}

			if len(taskManager.PendingTasks.PendingQueue) > 0 {
				fmt.Println("Pending Tasks: ", len(taskManager.PendingTasks.PendingQueue))
			}

		}
	}()

	go func() {
		//The task manager breaks apart after 2+ hours, we restart it as a temp fix
		ticker := time.NewTicker(time.Hour * 2)
		for range ticker.C {
			os.Exit(0)
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
