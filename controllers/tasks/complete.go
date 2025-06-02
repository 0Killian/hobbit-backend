package taskController

import (
	"database/sql"
	"fmt"
	"net/http"
	"server/common"
	"server/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

func HandleCompleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	tx, err := common.Db.Begin()
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
