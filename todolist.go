package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"

	_ "modernc.org/sqlite"
)

func main() {
	checkTable()
	port := os.Getenv("TODO_PORT")
	if len(port) == 0 {
		port = "7540"
	}
	serv := fmt.Sprintf("0.0.0.0:%s", port)
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/api/nextdate", apiNextDateHandle)
	http.HandleFunc("/api/task", apiTaskHandle)
	http.HandleFunc("/api/tasks", apiTasksHandle)
	http.HandleFunc("/api/task/done", apiDoneTask)
	http.HandleFunc("/api/signin", apiSignIn)
	err := http.ListenAndServe(serv, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func checkTable() {
	dbFile := os.Getenv("TODO_DBFILE")
	if len(dbFile) == 0 {
		appPath, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		dbFile = filepath.Join(filepath.Dir(appPath), "scheduler.db")
	}
	_, err := os.Stat(dbFile)
	var install bool
	if err != nil {
		install = true
	}
	if install {
		_, err = os.Create(dbFile)
		if err != nil {
			log.Fatal(err)
		}
		dB, err := sql.Open("sqlite", dbFile)
		if err != nil {
			log.Fatal(err)
		}
		defer dB.Close()
		_, err = dB.Exec("CREATE TABLE scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT, date VARCHAR(8) NOT NULL DEFAULT '', title VARCHAR(128) NOT NULL DEFAULT '', comment VARCHAR(128) NOT NULL DEFAULT '', repeat VARCHAR(128) NOT NULL DEFAULT '')")
		if err != nil {
			log.Fatal(err)
		}
		_, err = dB.Exec("CREATE INDEX scheduler_date ON scheduler (date)")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func apiNextDateHandle(res http.ResponseWriter, req *http.Request) {

	now := req.FormValue("now")
	date := req.FormValue("date")
	repeat := req.FormValue("repeat")

	nowTime, err := time.Parse("20060102", now)
	if err != nil {
		res.Write([]byte(err.Error()))
		return
	}

	nextDateValue, err := nextDate(nowTime, date, repeat)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(err.Error()))
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(nextDateValue))
}

type task struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func apiTaskHandle(res http.ResponseWriter, req *http.Request) {
	todoPass := os.Getenv("TODO_PASSWORD")
	if len(todoPass) > 0 {
		tokenCookie, err := req.Cookie("token")
		if err != nil {
			log.Println(err)
			http.Error(res, "authentification required", http.StatusUnauthorized)
			return
		}
		if !checkToken(tokenCookie.Value, todoPass) {
			log.Println("token is not valid")
			http.Error(res, "authentification required", http.StatusUnauthorized)
			return
		}
	}
	switch req.Method {
	case "POST":
		var buf bytes.Buffer
		var task task
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(buf.Bytes(), &task)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := addTask(task)
		if err != nil {
			status := http.StatusBadRequest
			if err.Error() == "internal server error" {
				status = http.StatusInternalServerError
			}
			res.WriteHeader(status)
			var m = map[string]string{"error": err.Error()}
			respBody, err := json.Marshal(m)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			res.Write(respBody)
			return
		}
		var m = map[string]string{"id": id}
		respBody, err := json.Marshal(m)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		res.Write(respBody)
	case "GET":
		id := req.FormValue("id")
		task, err := getTask(id)
		if err != nil {
			status := http.StatusBadRequest
			if err.Error() == "internal server error" {
				status = http.StatusInternalServerError
			}
			res.WriteHeader(status)
			var m = map[string]string{"error": err.Error()}
			respBody, err := json.Marshal(m)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			res.Write(respBody)
			return
		}
		respBody, err := json.Marshal(task)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		res.Write(respBody)
	case "PUT":
		var buf bytes.Buffer
		var task task
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		err = updateTask(task)
		if err != nil {
			status := http.StatusBadRequest
			if err.Error() == "internal server error" {
				status = http.StatusInternalServerError
			}
			res.WriteHeader(status)
			var m = map[string]string{"error": err.Error()}
			respBody, err := json.Marshal(m)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			res.Write(respBody)
			return
		}
		respBody := []byte("{}")
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		res.Write(respBody)
	case "DELETE":
		id := req.FormValue("id")
		err := deleteTask(id)
		if err != nil {
			status := http.StatusBadRequest
			if err.Error() == "internal server error" {
				status = http.StatusInternalServerError
			}
			res.WriteHeader(status)
			var m = map[string]string{"error": err.Error()}
			respBody, err := json.Marshal(m)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			res.Write(respBody)
			return
		}
		respBody := []byte("{}")
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		res.Write(respBody)
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func apiTasksHandle(res http.ResponseWriter, req *http.Request) {
	todoPass := os.Getenv("TODO_PASSWORD")
	if len(todoPass) > 0 {
		tokenCookie, err := req.Cookie("token")
		if err != nil {
			log.Println(err)
			http.Error(res, "authentification required", http.StatusUnauthorized)
			return
		}
		if !checkToken(tokenCookie.Value, todoPass) {
			log.Println("token is not valid")
			http.Error(res, "authentification required", http.StatusUnauthorized)
			return
		}
	}
	switch req.Method {
	case "GET":
		search := req.FormValue("search")
		tasks, err := getTasks(search)
		if err != nil {
			status := http.StatusBadRequest
			if err.Error() == "internal server error" {
				status = http.StatusInternalServerError
			}
			res.WriteHeader(status)
			var m = map[string]string{"error": err.Error()}
			respBody, err := json.Marshal(m)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			res.Write(respBody)
			return
		}
		if tasks == nil {
			tasks = make([]task, 0)
		}
		var m = map[string][]task{"tasks": tasks}
		respBody, err := json.Marshal(m)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		res.Write(respBody)
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func apiDoneTask(res http.ResponseWriter, req *http.Request) {
	todoPass := os.Getenv("TODO_PASSWORD")
	if len(todoPass) > 0 {
		tokenCookie, err := req.Cookie("token")
		if err != nil {
			log.Println(err)
			http.Error(res, "authentification required", http.StatusUnauthorized)
			return
		}
		if !checkToken(tokenCookie.Value, todoPass) {
			log.Println("token is not valid")
			http.Error(res, "authentification required", http.StatusUnauthorized)
			return
		}
	}
	switch req.Method {
	case "POST":
		id := req.FormValue("id")
		err := doneTask(id)
		if err != nil {
			status := http.StatusBadRequest
			if err.Error() == "internal server error" {
				status = http.StatusInternalServerError
			}
			res.WriteHeader(status)
			var m = map[string]string{"error": err.Error()}
			respBody, err := json.Marshal(m)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			res.Write(respBody)
			return
		}
		respBody := []byte("{}")
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		res.Write(respBody)
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func apiSignIn(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		m := make(map[string]string)
		var buf bytes.Buffer
		buf.ReadFrom(req.Body)
		err := json.Unmarshal(buf.Bytes(), &m)
		if err != nil {
			log.Println(err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		todoPass := os.Getenv("TODO_PASSWORD")
		if len(todoPass) == 0 {
			res.WriteHeader(http.StatusInternalServerError)
			var m = map[string]string{"error": "Пароль не установлен"}
			log.Println(m)
			respBody, err := json.Marshal(m)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			res.Write(respBody)
			return
		}
		pass, ok := m["password"]
		if !ok {
			res.WriteHeader(http.StatusInternalServerError)
			var m = map[string]string{"error": "Пароль не заполнен"}
			log.Println(m)
			respBody, err := json.Marshal(m)
			if err != nil {
				log.Println(err)
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			res.Write(respBody)
			return
		}
		if todoPass != pass {
			res.WriteHeader(http.StatusInternalServerError)
			var m = map[string]string{"error": "Пароль неверный"}
			log.Println(m)
			respBody, err := json.Marshal(m)
			if err != nil {
				log.Println(err)
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			res.Write(respBody)
			return
		}
		token := jwt.New(jwt.SigningMethodHS256)
		key := pass + "final"
		signedToken, err := token.SignedString([]byte(key))
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			var m = map[string]string{"error": "internal server error"}
			respBody, err := json.Marshal(m)
			if err != nil {
				log.Println(err)
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			res.Write(respBody)
			return
		}
		var mResp = map[string]string{"token": signedToken}
		respBody, err := json.Marshal(mResp)
		if err != nil {
			log.Println(err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		res.Write(respBody)
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func checkToken(token, todoPass string) bool {

	if len(token) == 0 {
		return false
	}
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		key := todoPass + "final"
		return []byte(key), nil
	})
	if err != nil {
		log.Println(err)
		return false
	}
	if !jwtToken.Valid {
		return false
	}

	return true
}
