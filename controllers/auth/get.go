package authController

import (
	"encoding/json"
	"log"
	"net/http"
	"server/common"
	"server/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/context"
)

func HandleGet(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleGet")
	tx, err := common.Db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, err := context.Get(r, "user").(jwt.MapClaims).GetSubject()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := models.FetchOneUserByCloudIamSub(tx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userRaw, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(userRaw)
}
