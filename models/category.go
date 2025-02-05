package models

import (
	"database/sql"
)

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func FetchOneCategory(conn *sql.Tx, categoryID string) (Category, error) {
	var category Category
	row := conn.QueryRow("select category_id, name from category where category_id = $1", categoryID)
	err := row.Scan(&category.ID, &category.Name)
	return category, err
}

func FetchAllCategories(conn *sql.Tx, limit int, offset int) ([]Category, error) {
	var categories []Category
	rows, err := conn.Query("select category_id, name from category limit $1 offset $2", limit, offset)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var category Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func CountCategories(conn *sql.Tx) (int, error) {
	var count int
	err := conn.QueryRow("select count(*) from category").Scan(&count)
	return count, err
}

func CreateCategory(conn *sql.Tx, category Category) error {
	_, err := conn.Exec("insert into category (category_id, name) values ($1, $2)", category.ID, category.Name)
	return err
}
