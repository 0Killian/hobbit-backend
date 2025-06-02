package taskController

import (
	"encoding/json"
	"net/http"
	"server/common"
	"server/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/context"
)

func HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	body := r.Body

	var payload createTaskPayload
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

	tx, err := common.Db.Begin()
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
