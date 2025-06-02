package taskController

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"server/common"
	"server/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

func HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	body := r.Body

	tx, err := common.Db.Begin()
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

	var payload createTaskPayload
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
