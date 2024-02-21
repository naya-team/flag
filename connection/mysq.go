package connection

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func NewMysqlConnection(dsn string) *sql.DB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}
