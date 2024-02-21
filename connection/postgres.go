package connection

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func NewPostgresConnection(dsn string) *sql.DB {

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}
