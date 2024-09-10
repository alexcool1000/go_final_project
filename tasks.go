package main

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type task struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func addTask(task task) (string, error) {
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
		date, err = nextDate(now, date, task.Repeat)
		if err != nil {
			return "", err
		}
	} else if date < nowString && len(task.Repeat) == 0 {
		date = nowString
	}

	db, err := openDb()
	if err != nil {
		return "", err
	}
	defer db.Close()

	res, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		log.Println(err)
		return "", error500()
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return "nil", error500()
	}
	return strconv.Itoa(int(id)), nil
}

func getTasks(search string) ([]task, error) {

	query := "SELECT id, date, title, comment, repeat FROM scheduler"
	var rows *sql.Rows
	var tasks []task
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
		rows, err = db.Query(query, sql.Named("search", search))
		if errors.Is(err, sql.ErrNoRows) {
			return tasks, nil
		}
		if err != nil {
			return nil, err
		}
	} else {
		query += " ORDER BY date LIMIT 50"
		rows, err = db.Query(query)
		if err == sql.ErrNoRows {
			return tasks, nil
		}
		if err != nil {
			log.Println(err)
			return nil, error500()
		}
	}
	defer rows.Close()

	for rows.Next() {
		task := task{}
		err = rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			log.Println(err)
			return nil, error500()
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, error500()
	}
	return tasks, nil
}

func getTask(id string) (task, error) {
	var task task
	if len(id) == 0 {
		return task, errors.New("не указан id")
	}
	db, err := openDb()
	if err != nil {
		return task, err
	}
	defer db.Close()
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
		return task, errors.New("не верный формат id")
	}
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id"
	row := db.QueryRow(query, sql.Named("id", idInt))
	err = row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if errors.Is(err, sql.ErrNoRows) {
		return task, errors.New("задача не найдена")
	}
	if err != nil {
		log.Println(err)
		return task, error500()
	}
	return task, nil
}

func updateTask(task task) error {
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
		date, err = nextDate(now, date, task.Repeat)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	db, err := openDb()
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()

	res, err := db.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
		sql.Named("id", id),
		sql.Named("date", date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		log.Println(err)
		return error500()
	}
	rowAffected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return error500()
	}
	if rowAffected == 0 {
		return errors.New("задача не найдена")
	}
	return nil
}

func doneTask(id string) error {
	task, err := getTask(id)
	if err != nil {
		return err
	}
	if len(task.Repeat) == 0 {
		return deleteTask(id)
	}
	newDate, err := nextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		return err
	}
	task.Date = newDate
	return updateTask(task)
}

func deleteTask(id string) error {
	if len(id) == 0 {
		return errors.New("не заполнен id задачи")
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
		return errors.New("не корректный id задачи")
	}

	db, err := openDb()
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	res, err := db.Exec("DELETE FROM scheduler WHERE id = :id",
		sql.Named("id", idInt))
	if err != nil {
		log.Println(err)
		return error500()
	}
	rowAffected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return error500()
	}
	if rowAffected == 0 {
		return errors.New("задача не найдена")
	}
	return nil
}

func openDb() (*sql.DB, error) {
	dbFile := os.Getenv("TODO_DBFILE")
	if len(dbFile) == 0 {
		appPath, err := os.Executable()
		if err != nil {
			log.Println(err)
			return nil, error500()
		}
		dbFile = filepath.Join(filepath.Dir(appPath), "scheduler.db")
	}
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Println(err)
		return nil, error500()
	}
	return db, nil
}

func error500() error {
	return errors.New("internal server error")
}
