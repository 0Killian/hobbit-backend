package main

import (
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"server/models"
)

// Configuration struct for Keycloak settings
type Config struct {
	ClientID     string
	ClientSecret string
	KeycloakURL  string
	Realm        string
	DatabaseURL  string
}

var (
	config    *Config
	publicKey *rsa.PublicKey
	db        *sql.DB
)

func init() {
	config = &Config{
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		KeycloakURL:  os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
	}

	publicKeyPath := os.Getenv("KEYCLOAK_PUBLIC_KEY_PATH")
	/*if config.ClientID == "" {
		log.Fatal("KEYCLOAK_CLIENT_ID is not set")
	}

	if config.ClientSecret == "" {
		log.Fatal("KEYCLOAK_CLIENT_SECRET is not set")
	}

	if config.KeycloakURL == "" {
		log.Fatal("KEYCLOAK_URL is not set")
	}

	if config.Realm == "" {
		log.Fatal("KEYCLOAK_REALM is not set")
	}*/
	if publicKeyPath == "" {
		log.Fatal("KEYCLOAK_PUBLIC_KEY_PATH is not set")
	}

	if config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	db, err = sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/tasks", corsMiddleware(authMiddleware(handleGetTasks))).Methods("GET")
	r.HandleFunc("/api/v1/tasks", corsMiddleware(authMiddleware(handleCreateTask))).Methods("POST")
	r.HandleFunc("/api/v1/tasks/{uuid}", corsMiddleware(authMiddleware(handleGetTask))).Methods("GET")
	r.HandleFunc("/api/v1/tasks/{uuid}", corsMiddleware(authMiddleware(handleUpdateTask))).Methods("PUT")
	r.HandleFunc("/api/v1/tasks/{uuid}/complete", corsMiddleware(authMiddleware(handleCompleteTask))).Methods("PUT")

	log.Printf("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func corsMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		if !strings.HasPrefix(bearer, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		accessToken := bearer[7:]

		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		context.Set(r, "user", token.Claims)

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sub, err := token.Claims.(jwt.MapClaims).GetSubject()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = models.FetchOneUserByCloudIamSub(tx, sub)
		if err != nil {
			if err == sql.ErrNoRows {
				user := models.User{
					UserID:      uuid.New().String(),
					CloudIamSub: token.Claims.(jwt.MapClaims)["sub"].(string),
					Rank:        0,
				}

				if err := models.CreateUser(tx, user); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleGetTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	userID := context.Get(r, "user").(jwt.MapClaims)["sub"].(string)
	user, err := models.FetchOneUserByCloudIamSub(tx, userID)

	limit := 50
	offset := 0
	if page := query.Get("page"); page != "" {
		p, err := strconv.Atoi(page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		offset = (p - 1) * limit
	}

	filter := models.TaskFilter{}

	if name := query.Get("name"); name != "" {
		filter.Name = &name
	}

	if description := query.Get("description"); description != "" {
		filter.Description = &description
	}

	if categories := query.Get("categories"); categories != "" {
		filter.Categories = strings.Split(categories, ",")
	}

	filter.UserID = &user.UserID

	if completed := query.Get("completed"); completed != "" {
		if completedBool, err := strconv.ParseBool(completed); err == nil {
			filter.Completed = &completedBool
		}
	}

	if completionTimeMin := query.Get("completionTimeMin"); completionTimeMin != "" {
		time, err := time.Parse("2006-01-02", completionTimeMin)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		filter.CompletionTimeMin = &time
	}

	if completionTimeMax := query.Get("completionTimeMax"); completionTimeMax != "" {
		time, err := time.Parse("2006-01-02", completionTimeMax)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		filter.CompletionTimeMax = &time
	}

	sortBy := (*models.TaskSortBy)(nil)
	switch sort := query.Get("sort"); sort {
	case "name":
		sort := models.TaskSortByName
		sortBy = &sort
	case "completion_time":
		sort := models.TaskSortByCompletionTime
		sortBy = &sort
	}

	tasks, count, err := models.FetchAllTasks(tx, filter, sortBy, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"tasks": %s, "current_page": %d, max_page": %d}`, jsonData, offset/limit+1, (count-1)/limit+1)))
}

func handleGetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	task, err := models.FetchOneTask(tx, uuid)
	if err == nil {
		jsonData, err := json.Marshal(task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	if err == sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type CreateTaskPayload struct {
	Quantity    int    `json:"quantity"`
	Unit        string `json:"unit"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Frequency   string `json:"frequency"`
}

func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	body := r.Body

	var payload CreateTaskPayload
	err := json.NewDecoder(body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	unit, err := models.UnitFromString(payload.Unit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	userID, err := context.Get(r, "user").(jwt.MapClaims).GetSubject()
	user, err := models.FetchOneUserByCloudIamSub(tx, userID)

	task := models.Task{
		TaskID:           uuid.New().String(),
		Quantity:         payload.Quantity,
		Unit:             unit,
		Name:             payload.Name,
		Description:      payload.Description,
		Frequency:        payload.Frequency,
		ExperienceGained: 100,
		IsPublic:         false,
		UserID:           &user.UserID,
	}

	err = models.CreateTask(tx, task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	body := r.Body

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	userID, err := context.Get(r, "user").(jwt.MapClaims).GetSubject()
	user, err := models.FetchOneUserByCloudIamSub(tx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payload CreateTaskPayload
	err = json.NewDecoder(body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err := models.FetchOneTask(tx, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if task.UserID != nil && *task.UserID != user.UserID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	unit, err := models.UnitFromString(payload.Unit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task = models.Task{
		TaskID:           uuid,
		Quantity:         payload.Quantity,
		Unit:             unit,
		Name:             payload.Name,
		Description:      payload.Description,
		Frequency:        payload.Frequency,
		ExperienceGained: 100,
		IsPublic:         false,
		UserID:           task.UserID,
	}

	err = models.UpdateTask(tx, task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleCompleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	fmt.Println("Fetching user")
	userID, err := context.Get(r, "user").(jwt.MapClaims).GetSubject()
	user, err := models.FetchOneUserByCloudIamSub(tx, userID)

	fmt.Println("Fetching task")
	task, err := models.FetchOneTask(tx, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if task.UserID != nil && *task.UserID != user.UserID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = models.CompleteTask(tx, user.UserID, task.TaskID, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
