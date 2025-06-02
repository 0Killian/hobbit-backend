package authController

import (
	"net/http"

	"github.com/gorilla/context"
)

func HandleGet(w http.ResponseWriter, r *http.Request) {
	user, err := context.Get(r, "user").(jwt.MapClaims)

}
