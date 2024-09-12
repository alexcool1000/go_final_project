package services

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"
	"todoList/internal/database"
	"todoList/internal/models"
)

func AddTask(task models.Task) (string, error) {
	if len(task.Title) == 0 {
		return "", errors.New("не заполнен заголовок задачи")
	}
	now := time.Now()
	nowString := now.Format("20060102")
	date := task.Date
	if len(date) == 0 {
		date = now.Format("20060102")
	}
	_, err := time.Parse("20060102", date)
	if err != nil {
		log.Println(err)
		return "", errors.New("некорректный формат даты")
	}
	if date < nowString && len(task.Repeat) > 0 {
		date, err = NextDate(now, date, task.Repeat)
		if err != nil {
			return "", err
		}
	} else if date < nowString && len(task.Repeat) == 0 {
		date = nowString
	}

	res, err := database.Db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		log.Println(err)
		return "", Error500()
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return "nil", Error500()
	}
	return strconv.Itoa(int(id)), nil
}

func GetTasks(search string) ([]models.Task, error) {

	query := "SELECT id, date, title, comment, repeat FROM scheduler"
	var rows *sql.Rows
	var tasks []models.Task
	var err error
	if len(search) > 0 {
		date, err := time.Parse("02.01.2006", search)
		if err == nil {
			query += " WHERE date = :search"
			search = date.Format("20060102")
		} else {
			query += " WHERE title LIKE :search OR comment LIKE :search"
			search = "%" + search + "%"
		}
		query += " ORDER BY date LIMIT 50"
		rows, err = database.Db.Query(query, sql.Named("search", search))
		if errors.Is(err, sql.ErrNoRows) {
			return tasks, nil
		}
		if err != nil {
			return nil, err
		}
	} else {
		query += " ORDER BY date LIMIT 50"
		rows, err = database.Db.Query(query)
		if err == sql.ErrNoRows {
			return tasks, nil
		}
		if err != nil {
			log.Println(err)
			return nil, Error500()
		}
	}
	defer rows.Close()

	for rows.Next() {
		task := models.Task{}
		err = rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			log.Println(err)
			return nil, Error500()
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, Error500()
	}
	return tasks, nil
}

func GetTask(id string) (models.Task, error) {
	var task models.Task
	if len(id) == 0 {
		return task, errors.New("не указан id")
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
		return task, errors.New("не верный формат id")
	}
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id"
	row := database.Db.QueryRow(query, sql.Named("id", idInt))
	err = row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if errors.Is(err, sql.ErrNoRows) {
		return task, errors.New("задача не найдена")
	}
	if err != nil {
		log.Println(err)
		return task, Error500()
	}
	return task, nil
}

func UpdateTask(task models.Task) error {
	if len(task.Title) == 0 {
		return errors.New("не заполнен заголовок задачи")
	}
	if len(task.Id) == 0 {
		return errors.New("не заполнен id задачи")
	}
	id, err := strconv.Atoi(task.Id)
	if err != nil {
		log.Println(err)
		return errors.New("не корректный id задачи")
	}
	now := time.Now()
	nowString := now.Format("20060102")
	date := task.Date
	if len(date) == 0 {
		date = now.Format("20060102")
	}
	_, err = time.Parse("20060102", date)
	if err != nil {
		log.Println(err)
		return errors.New("не корректная дата задачи")
	}
	if date < nowString && len(task.Repeat) > 0 {
		date, err = NextDate(now, date, task.Repeat)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	res, err := database.Db.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
		sql.Named("id", id),
		sql.Named("date", date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		log.Println(err)
		return Error500()
	}
	rowAffected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return Error500()
	}
	if rowAffected == 0 {
		return errors.New("задача не найдена")
	}
	return nil
}

func DoneTask(id string) error {
	task, err := GetTask(id)
	if err != nil {
		return err
	}
	if len(task.Repeat) == 0 {
		return DeleteTask(id)
	}
	newDate, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		return err
	}
	task.Date = newDate
	return UpdateTask(task)
}

func DeleteTask(id string) error {
	if len(id) == 0 {
		return errors.New("не заполнен id задачи")
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
		return errors.New("не корректный id задачи")
	}

	res, err := database.Db.Exec("DELETE FROM scheduler WHERE id = :id",
		sql.Named("id", idInt))
	if err != nil {
		log.Println(err)
		return Error500()
	}
	rowAffected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return Error500()
	}
	if rowAffected == 0 {
		return errors.New("задача не найдена")
	}
	return nil
}

func Error500() error {
	return errors.New("internal server error")
}
