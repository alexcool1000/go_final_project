package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func main() {
	openDB()
	defer db.Close()
	port := os.Getenv("TODO_PORT")
	if len(port) == 0 {
		port = "7540"
	}
	serv := fmt.Sprintf("0.0.0.0:%s", port)
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("GET /api/nextdate", apiNextDateHandle)
	http.HandleFunc("GET /api/task", auth(apiGetTaskHandle))
	http.HandleFunc("POST /api/task", auth(apiPostTaskHandle))
	http.HandleFunc("PUT /api/task", auth(apiPutTaskHandle))
	http.HandleFunc("DELETE /api/task", auth(apiDeleteTaskHandle))
	http.HandleFunc("/api/tasks", auth(apiTasksHandle))
	http.HandleFunc("/api/task/done", auth(apiDoneTask))
	http.HandleFunc("/api/signin", apiSignIn)
	err := http.ListenAndServe(serv, nil)
	if err != nil {
		log.Fatal(err)
	}
}
