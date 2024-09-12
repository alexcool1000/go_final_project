package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"
	"todoList/internal/models"
	"todoList/internal/services"

	"github.com/golang-jwt/jwt/v5"
	_ "modernc.org/sqlite"
)

func ApiPostTaskHandle(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	var task models.Task
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
	id, err := services.AddTask(task)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.Error500()) {
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
}

func ApiGetTaskHandle(res http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	task, err := services.GetTask(id)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.Error500()) {
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
}
func ApiPutTaskHandle(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	var task models.Task
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	err = services.UpdateTask(task)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.Error500()) {
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
}
func ApiDeleteTaskHandle(res http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	err := services.DeleteTask(id)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.Error500()) {
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
}

func ApiTasksHandle(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		search := req.FormValue("search")
		tasks, err := services.GetTasks(search)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, services.Error500()) {
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
			tasks = make([]models.Task, 0)
		}
		var m = map[string][]models.Task{"tasks": tasks}
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

func ApiDoneTask(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		id := req.FormValue("id")
		err := services.DoneTask(id)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, services.Error500()) {
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

func ApiSignIn(res http.ResponseWriter, req *http.Request) {
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

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
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
		next(res, req)
	})
}

func ApiNextDateHandle(res http.ResponseWriter, req *http.Request) {

	now := req.FormValue("now")
	date := req.FormValue("date")
	repeat := req.FormValue("repeat")

	nowTime, err := time.Parse("20060102", now)
	if err != nil {
		res.Write([]byte(err.Error()))
		return
	}

	nextDateValue, err := services.NextDate(nowTime, date, repeat)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(err.Error()))
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(nextDateValue))
}
