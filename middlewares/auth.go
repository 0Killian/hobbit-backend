package middlewares

import (
	"database/sql"
	"fmt"
	"net/http"
	"server/common"
	"server/models"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/context"
	_ "github.com/lib/pq"
)

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		if !strings.HasPrefix(bearer, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		accessToken := bearer[7:]

		if common.Rdb != nil {
			exists, err := common.Rdb.Exists(common.Ctx, "user:"+accessToken).Result()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if exists == 1 {
				user := models.User{}
				common.Rdb.Get(common.Ctx, "user:"+accessToken).Scan(&user)

				context.Set(r, "user", user)

				next(w, r)
				return
			}
		}

		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return common.PublicKey, nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		context.Set(r, "user", token.Claims)

		tx, err := common.Db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sub, err := token.Claims.(jwt.MapClaims).GetSubject()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user, err := models.FetchOneUserByCloudIamSub(tx, sub)
		if err != nil {
			if err == sql.ErrNoRows {
				user = models.User{
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

		if common.Rdb != nil {
			present, err := common.Rdb.Exists(common.Ctx, "user:"+user.CloudIamSub).Result()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if present == 0 {
				err = common.Rdb.Set(common.Ctx, "user:"+user.CloudIamSub, user, 0).Err()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
