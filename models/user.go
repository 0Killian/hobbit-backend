package models

import (
	"database/sql"
)

type User struct {
	UserID      string  `json:"id"`
	CloudIamSub string  `json:"cloud_iam_sub"`
	Rank        float32 `json:"rank"`
}

type UserSortBy string

const (
	UserSortByRank UserSortBy = "rank"
)

func FetchOneUser(conn *sql.Tx, userID string) (User, error) {
	var user User
	row := conn.QueryRow("select u.user_id, u.cloud_iam_sub, ue.rank from \"user\" u inner join user_experience ue on u.user_id = ue.user_id where user_id = $1", userID)
	err := row.Scan(&user.UserID, &user.CloudIamSub, &user.Rank)
	return user, err
}

func FetchOneUserByCloudIamSub(conn *sql.Tx, cloudIamSub string) (User, error) {
	var user User
	row := conn.QueryRow("select u.user_id, u.cloud_iam_sub, ue.rank from \"user\" u inner join user_experience ue on u.user_id = ue.user_id where cloud_iam_sub = $1", cloudIamSub)
	err := row.Scan(&user.UserID, &user.CloudIamSub, &user.Rank)
	return user, err
}

func CountUsers(conn *sql.Tx) (int, error) {
	var count int
	err := conn.QueryRow("select count(*) from user").Scan(&count)
	return count, err
}

func FetchAllUsers(conn *sql.Tx, sortBy *UserSortBy, limit int, offset int) ([]User, error) {
	var users []User

	query := "select u.user_id, u.cloud_iam_sub, ue.rank from \"user\" u inner join ue on u.user_id = ue.user_id"
	if sortBy != nil {
		query += " order by " + string(*sortBy)
	}
	query += " limit $1 offset $2"

	rows, err := conn.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var user User
		err = rows.Scan(&user.UserID, &user.CloudIamSub, &user.Rank)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func CreateUser(conn *sql.Tx, user User) error {
	_, err := conn.Exec("insert into \"user\" (user_id, cloud_iam_sub) values ($1, $2)", user.UserID, user.CloudIamSub)
	if err != nil {
		return err
	}

	_, err = conn.Exec("insert into user_experience (user_id, rank) values ($1, $2)", user.UserID, user.Rank)
	return err
}

func Update(conn *sql.Tx, user User) error {
	_, err := conn.Exec("update \"user\" set cloud_iam_sub = $1 where user_id = $2", user.CloudIamSub, user.UserID)
	if err != nil {
		return err
	}

	_, err = conn.Exec("update user_experience set rank = $1 where user_id = $2", user.Rank, user.UserID)
	return err
}
