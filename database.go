package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
)

func openDB() {
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
		db, err = sql.Open("sqlite", dbFile)
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.Exec("CREATE TABLE scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT, date VARCHAR(8) NOT NULL DEFAULT '', title VARCHAR(128) NOT NULL DEFAULT '', comment VARCHAR(128) NOT NULL DEFAULT '', repeat VARCHAR(128) NOT NULL DEFAULT '')")
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.Exec("CREATE INDEX scheduler_date ON scheduler (date)")
		if err != nil {
			log.Fatal(err)
		}
	}
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
	}
}
