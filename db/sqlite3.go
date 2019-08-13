package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3" // comment needed for blank import by go lint
)

type Row struct {
	starts_hash string
	starts_id   int64
	starts_at   int64
	ends_hash   string
	ends_id     int64
	ends_at     int64
	region_id   int64
	alertname   string
	description string
	status      string
}

// Connect opens a new connection
func Connect(sqliteDB string) (con *sql.DB, err error) {
	// prepare database
	con, err = sql.Open("sqlite3", sqliteDB)
	if err != nil {
		panic(err)
	}
	return
}

// Close closes a connection
func Close(con *sql.DB) {
	con.Close()
}

// Init creates the initial database schema and indices
func Init(con *sql.DB) (err error) {

	// create annotations table
	statement, err := con.Prepare(
		`CREATE TABLE IF NOT EXISTS annotations (
			starts_hash TEXT PRIMARY KEY, 
			starts_id  BIGINT DEFAULT 0,
			starts_at BIGINT DEFAULT 0,
			ends_hash TEXT DEFAULT NULL,
			ends_id BIGINT DEFAULT 0,
			ends_at BIGINT DEFAULT 0,
			region_id BIGINT DEFAULT 0, 
			alertname TEXT DEFAULT NULL,
			description TEXT NULL,
			status TEXT NOT NULL)`)

	statement.Exec()

	// create index on region_id
	statement, err = con.Prepare(
		`CREATE INDEX IF NOT EXISTS region_id_idx ON annotations (region_id)`)

	statement.Exec()

	statement, err = con.Prepare(
		`CREATE INDEX IF NOT EXISTS status_idx ON annotations (status)`)

	statement.Exec()

	return
}

// Insert inserts a new raw with colums the maps keys and values the maps values
func Insert(con *sql.DB, vals map[string]string) (err error) {

	q := "INSERT INTO annotations ("

	keys := make([]string, 0, len(vals))
	values := make([]string, 0, len(vals))

	for k, v := range vals {
		keys = append(keys, k)
		values = append(values, v)
	}
	q += strings.Join(keys, ",")
	q += ") VALUES ("
	q += strings.Join(values, ",")
	q += ");"

	fmt.Printf("%s\n", q)
	statement, err := con.Prepare(q)
	if err != nil {
		return
	}

	statement.Exec()
	return
}

// Select row from annotations based on key: val
func Select(con *sql.DB, vals map[string]string) (*Row, error) {

	r := &Row{}
	q := "SELECT starts_hash, alertname, description, status FROM annotations where "
	for k, v := range vals {
		q += k
		q += "="
		q += v
	}
	fmt.Printf("%s\n", q)
	rows, err := con.Query(q)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		rows.Scan(&(r.starts_hash), &(r.alertname), &(r.description), &(r.status))
	}
	return r, nil
}

// func updateAlert() {

// }

// func deleteAlert() {

// }
