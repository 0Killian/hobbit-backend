package models

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

type Unit int

const (
	UnitNone Unit = iota
	UnitDistance
	UnitReps
	UnitTime
)

var unitStrings = map[Unit]string{
	UnitNone:     "none",
	UnitDistance: "distance",
	UnitReps:     "reps",
	UnitTime:     "time",
}

var unitValues = map[string]Unit{
	"none":     UnitNone,
	"distance": UnitDistance,
	"reps":     UnitReps,
	"time":     UnitTime,
}

var (
	ErrInvalidUnit = errors.New("invalid unit")
)

func UnitFromString(s string) (Unit, error) {
	if u, ok := unitValues[s]; ok {
		return u, nil
	}
	return UnitNone, ErrInvalidUnit
}

func (unit Unit) MarshalJSON() ([]byte, error) {
	return []byte("\"" + unitStrings[unit] + "\""), nil
}

type taskFromQuery struct {
	TaskID           string
	Quantity         int
	Unit             string
	Name             string
	Description      string
	Frequency        string
	ExperienceGained int
	IsPublic         bool
	UserID           *string
}

type Task struct {
	TaskID           string  `json:"task_id"`
	Quantity         int     `json:"quantity"`
	Unit             Unit    `json:"unit"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Frequency        string  `json:"frequency"`
	ExperienceGained int     `json:"experience_gained"`
	IsPublic         bool    `json:"is_public"`
	UserID           *string `json:"user_id"`
}

func makeTask(task taskFromQuery) Task {
	return Task{
		TaskID:           task.TaskID,
		Quantity:         task.Quantity,
		Unit:             unitValues[task.Unit],
		Name:             task.Name,
		Description:      task.Description,
		Frequency:        task.Frequency,
		ExperienceGained: task.ExperienceGained,
		IsPublic:         task.IsPublic,
		UserID:           task.UserID,
	}
}

func makeTaskFromQuery(task Task) taskFromQuery {
	return taskFromQuery{
		TaskID:           task.TaskID,
		Quantity:         task.Quantity,
		Unit:             unitStrings[task.Unit],
		Name:             task.Name,
		Description:      task.Description,
		Frequency:        task.Frequency,
		ExperienceGained: task.ExperienceGained,
		IsPublic:         task.IsPublic,
		UserID:           task.UserID,
	}
}

type TaskFilter struct {
	Name              *string
	Description       *string
	Categories        []string
	UserID            *string
	Completed         *bool
	CompletionTimeMin *time.Time
	CompletionTimeMax *time.Time
}

type TaskSortBy string

const (
	TaskSortByName           TaskSortBy = "name"
	TaskSortByCompletionTime TaskSortBy = "complete_timestamp"
)

func FetchOneTask(conn *sql.Tx, taskID string) (Task, error) {
	var task taskFromQuery
	row := conn.QueryRow("select task.task_id, quantity, unit, name, description, frequency, experience_gained, is_public, user_id from task left join user_task on user_task.task_id = task.task_id where task.task_id = $1", taskID)
	err := row.Scan(&task.TaskID, &task.Quantity, &task.Unit, &task.Name, &task.Description, &task.Frequency, &task.ExperienceGained, &task.IsPublic, &task.UserID)

	if err != nil {
		return Task{}, err
	}

	return makeTask(task), nil
}

func CountTasks(conn *sql.Tx) (int, error) {
	var count int
	err := conn.QueryRow("select count(*) from task").Scan(&count)
	return count, err
}

func FetchAllTasks(conn *sql.Tx, filter TaskFilter, sortBy *TaskSortBy, limit int, offset int) ([]Task, int, error) {
	var tasks []Task

	query := "select task.task_id, quantity, unit, name, description, frequency, experience_gained, is_public, user_id from task"
	joins := " left join user_task on task.task_id = user_task.task_id"
	where := ""
	paramIndex := 1

	if filter.Name != nil {
		where += " and name like $" + strconv.Itoa(paramIndex)
		paramIndex++
	}

	if filter.Description != nil {
		where += " and description like $" + strconv.Itoa(paramIndex)
		paramIndex++
	}

	if len(filter.Categories) > 0 {
		joins += " inner join task_category on task.task_id = task_category.category_id inner join category on task_category.category_id = category.category_id"
		where += " and category.name in ("
		for i, _ := range filter.Categories {
			if i > 0 {
				where += ", "
			}
			where += "$" + strconv.Itoa(paramIndex)
			paramIndex++
		}
		where += ")"
	}

	if filter.UserID != nil {
		where += " and (is_public = true or user_task.user_id = $" + strconv.Itoa(paramIndex) + ")"
		paramIndex++
	} else {
		where += " and is_public = true"
	}

	completionTimeJoin := filter.CompletionTimeMin != nil || filter.CompletionTimeMax != nil || filter.Completed != nil
	completionTimeJoin = completionTimeJoin || (sortBy != nil && *sortBy == TaskSortByCompletionTime)
	if completionTimeJoin {
		joins += " left join task_completion on user_task.user_task_id = task_completion.user_task_id"
		if filter.Completed == nil {
			if filter.CompletionTimeMin != nil {
				where += " and task_completion.complete_timestamp >= $" + strconv.Itoa(paramIndex)
				paramIndex++
			}
			if filter.CompletionTimeMax != nil {
				where += " and task_completion.complete_timestamp <= $" + strconv.Itoa(paramIndex)
				paramIndex++
			}
		} else {
			if *filter.Completed {
				where += " and task_completion.complete_timestamp is not null"
			} else {
				where += " and task_completion.complete_timestamp is null"
			}
		}
	}

	if where != "" {
		query += joins + " where" + where[4:]
	}

	if sortBy != nil {
		switch *sortBy {
		case TaskSortByCompletionTime:
			query += " order by task_completion.complete_timestamp desc"
		default:
			query += " order by name"
		}
	}
	query += " limit $" + strconv.Itoa(paramIndex) + " offset $" + strconv.Itoa(paramIndex+1)
	args := make([]interface{}, paramIndex+1)
	paramIndex = 0

	if filter.Name != nil {
		args[paramIndex] = "%" + *filter.Name + "%"
		paramIndex++
	}

	if filter.Description != nil {
		args[paramIndex] = "%" + *filter.Description + "%"
		paramIndex++
	}

	for i, _ := range filter.Categories {
		args[paramIndex] = filter.Categories[i]
		paramIndex++
	}

	if filter.UserID != nil {
		args[paramIndex] = *filter.UserID
		paramIndex++
	}

	if filter.Completed != nil {
		if filter.CompletionTimeMin != nil {
			args[paramIndex] = filter.CompletionTimeMin
			paramIndex++
		}

		if filter.CompletionTimeMax != nil {
			args[paramIndex] = filter.CompletionTimeMax
			paramIndex++
		}
	}

	args[paramIndex] = limit
	args[paramIndex+1] = offset

	fmt.Println(query, args)

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}

	for rows.Next() {
		var task taskFromQuery
		err := rows.Scan(&task.TaskID, &task.Quantity, &task.Unit, &task.Name, &task.Description, &task.Frequency, &task.ExperienceGained, &task.IsPublic, &task.UserID)

		if err != nil {
			return nil, 0, err
		}

		tasks = append(tasks, makeTask(task))
	}

	query = "select count(*) from task" + joins + where
	fmt.Println(query, args)

	var total int
	err = conn.QueryRow(query, args[:paramIndex]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func CreateTask(conn *sql.Tx, task Task) error {
	var t = makeTaskFromQuery(task)
	fmt.Println(task)
	_, err := conn.Exec("insert into task (task_id, quantity, unit, name, description, frequency, experience_gained, is_public) values ($1, $2, $3, $4, $5, $6, $7, $8)",
		t.TaskID, t.Quantity, t.Unit, t.Name, t.Description, t.Frequency, t.ExperienceGained, t.IsPublic)
	if err != nil {
		return err
	}
	if task.UserID != nil {
		fmt.Println(*task.UserID, t.TaskID)
		_, err = conn.Exec("insert into user_task (user_id, task_id) values ($1, $2)", *task.UserID, t.TaskID)
	}
	return err
}

func UpdateTask(conn *sql.Tx, task Task) error {
	var t = makeTaskFromQuery(task)
	_, err := conn.Exec("update task set quantity = $2, unit = $3, name = $4, description = $5, frequency = $6, experience_gained = $7, is_public = $8 where task_id = $1",
		t.TaskID, t.Quantity, t.Unit, t.Name, t.Description, t.Frequency, t.ExperienceGained, t.IsPublic)
	return err
}

func CompleteTask(conn *sql.Tx, userID string, taskID string, completionTime time.Time) error {
	row := conn.QueryRow("select user_task_id, experience_gained from user_task inner join task on task.task_id = user_task.task_id where task_id = $1 and user_id = $2", taskID, userID)

	var userTaskID string
	var experienceGained int
	err := row.Scan(&userTaskID, &experienceGained)
	if err != nil {
		return err
	}

	_, err = conn.Exec("insert into task_completion (user_task_id, complete_timestamp) values ($1, $2)", userTaskID, completionTime)
	if err != nil {
		return err
	}

	row = conn.QueryRow("select rank from user_experience where user_id = $1", userID)

	var rank float64
	err = row.Scan(&rank)
	if err != nil {
		return err
	}

	nextRankThreshold := math.Floor(rank) * 1000
	rank += float64(experienceGained) / nextRankThreshold

	_, err = conn.Exec("update user_experience set rank = $1 where user_id = $2", rank, userID)
	return err
}
