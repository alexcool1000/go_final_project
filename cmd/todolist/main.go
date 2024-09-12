package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"todoList/internal/database"
	"todoList/internal/transport/rest"

	_ "modernc.org/sqlite"
)

func main() {
	database.OpenDB()
	defer database.Db.Close()
	port := os.Getenv("TODO_PORT")
	if len(port) == 0 {
		port = "7540"
	}
	webDir := os.Getenv("TODO_WEBDIR")
	if len(webDir) == 0 {
		webDir = "../../web"
	}

	serv := fmt.Sprintf("0.0.0.0:%s", port)
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("GET /api/nextdate", rest.ApiNextDateHandle)
	http.HandleFunc("GET /api/task", rest.Auth(rest.ApiGetTaskHandle))
	http.HandleFunc("POST /api/task", rest.Auth(rest.ApiPostTaskHandle))
	http.HandleFunc("PUT /api/task", rest.Auth(rest.ApiPutTaskHandle))
	http.HandleFunc("DELETE /api/task", rest.Auth(rest.ApiDeleteTaskHandle))
	http.HandleFunc("/api/tasks", rest.Auth(rest.ApiTasksHandle))
	http.HandleFunc("/api/task/done", rest.Auth(rest.ApiDoneTask))
	http.HandleFunc("/api/signin", rest.ApiSignIn)
	err := http.ListenAndServe(serv, nil)
	if err != nil {
		log.Fatal(err)
	}
}
