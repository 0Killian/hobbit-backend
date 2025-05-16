package main

import (
	"log"
	"net/http"
	"server/controllers/auth"
	"server/controllers/tasks"
	"server/middlewares"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	r := mux.NewRouter()

	r.Use(middlewares.Cors)

	r.HandleFunc("/api/v1/auth/me", middlewares.Auth(authController.HandleGet)).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/v1/tasks", middlewares.Auth(taskController.HandleGetTasks)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/tasks/{uuid}", middlewares.Auth(taskController.HandleGetTask)).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/v1/tasks", middlewares.Auth(taskController.HandleCreateTask)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/tasks/{uuid}", middlewares.Auth(taskController.HandleUpdateTask)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/v1/tasks/{uuid}/complete", middlewares.Auth(taskController.HandleCompleteTask)).Methods("PUT", "OPTIONS")

	log.Printf("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
