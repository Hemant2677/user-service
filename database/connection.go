package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)
var Db *sql.DB

func init() {
	var err error
	cdb := "postgres://hemant:5689@localhost/postgres?sslmode=disable"
	Db, err = sql.Open("postgres", cdb)

	if err != nil {
		panic(err)
	}

	if err = Db.Ping(); err != nil {
		panic(err)
	}
	// Confirming database connection
	fmt.Println("The database is connected")
}
