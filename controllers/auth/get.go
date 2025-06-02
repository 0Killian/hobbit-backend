package authController

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/context"
)

func HandleGet(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(jwt.MapClaims)

	userRaw, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(userRaw)
}
