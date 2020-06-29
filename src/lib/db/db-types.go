package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	Host string
	Port string
	User string
	Db   string
}

func (d *Database) getPass() string {
	return os.Getenv("MYSQL_ROOT_PASSWORD")
}

func (d *Database) Open() *sql.DB {
	pass := d.getPass()
	conn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		d.User,
		pass,
		d.Host,
		d.Port,
		d.Db,
	)

	db, err := sql.Open("mysql", conn)
	if err != nil {
		panic(err)
	}

	return db
}

func (d *Database) Query(q string) *sql.Rows {
	db := d.Open()

	rows, err := db.Query(q)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	return rows
}
