package taskController

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"server/common"
	"server/models"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

func HandleGetTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	tx, err := common.Db.Begin()
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
	w.Write([]byte(fmt.Sprintf(`{"tasks": %s, "current_page": %d, "max_page": %d}`, jsonData, offset/limit+1, (count-1)/limit+1)))
}

func HandleGetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	tx, err := common.Db.Begin()
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
