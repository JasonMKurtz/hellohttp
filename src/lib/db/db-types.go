package db

import (
	"database/sql"
	"fmt"
	"os"
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

func (d *Database) open() *sql.DB {
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

func (d *Database) Read() string {
	db := d.open()

	var res string
	err := db.QueryRow("SELECT 1").Scan(&res)
	if err != nil {
		panic(err.Error)
	}

	return res
}
