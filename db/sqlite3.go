package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3" // comment needed for blank import by go lint
)

// Row is the struct representation of the schema of annotations table
type Row struct {
	AlertHash   string
	StartsID    int64
	StartsAt    int64
	EndsID      int64
	EndsAt      int64
	RegionID    int64
	Alertname   string
	Description string
	Status      string
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
			alert_hash TEXT PRIMARY KEY, 
			starts_id  BIGINT DEFAULT 0,
			starts_at BIGINT DEFAULT 0,
			ends_id BIGINT DEFAULT 0,
			ends_at BIGINT DEFAULT 0,
			region_id BIGINT DEFAULT 0, 
			alertname TEXT DEFAULT NULL,
			description TEXT DEFAULT NULL,
			status TEXT DEFAULT NULL)`)

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

	//fmt.Printf("%s\n", q)
	statement, err := con.Prepare(q)
	if err != nil {
		return
	}

	statement.Exec()
	return
}

// Select row from annotations based on key: val
func Select(con *sql.DB, vals map[string]string) ([]Row, error) {

	var r []Row
	q := "SELECT alert_hash, starts_id, starts_at, ends_id, ends_at, region_id, alertname, description, status FROM annotations where "
	for k, v := range vals {
		q += k
		q += "="
		q += v
	}
	//fmt.Printf("%s\n", q)
	rows, err := con.Query(q)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		tempRow := Row{}
		rows.Scan(&(tempRow.AlertHash), &(tempRow.StartsID), &(tempRow.StartsAt),
			&(tempRow.EndsID), &(tempRow.EndsAt),
			&(tempRow.RegionID), &(tempRow.Alertname), &(tempRow.Description), &(tempRow.Status))
		r = append(r, tempRow)
	}
	return r, nil
}

// UpdateWithHash updates the an entry with starting and ending annotation ids
func UpdateWithHash(con *sql.DB, vals map[string]string, hash string) (err error) {
	q := "UPDATE annotations SET "
	i := 0
	for k, v := range vals {
		q += k
		q += "="
		q += v
		if i != len(vals)-1 {
			q += ","
		}
		i++
	}
	q += fmt.Sprintf(" WHERE alert_hash = %q;", hash)
	statement, err := con.Prepare(q)
	if err != nil {
		return
	}

	statement.Exec()
	return
}

// UpdateWithEndsAt updates the status of a row as obselete if ends_at is older than 1 month
func UpdateWithEndsAt(con *sql.DB, vals map[string]string) (err error) {
	q := "UPDATE annotations SET "
	i := 0
	for k, v := range vals {
		q += k
		q += "="
		q += v
		if i != len(vals)-1 {
			q += ","
		}
		i++
	}
	q += " where strftime('%Y-%m-%d %H:%M:%f', ends_at/1000, 'unixepoch') < datetime('now','-31 days');"
	statement, err := con.Prepare(q)
	if err != nil {
		return
	}

	statement.Exec()
	return
}

// DeleteWithHash delete an entry where status matches st
func DeleteWithHash(con *sql.DB, hash string) (err error) {
	q := fmt.Sprintf("DELETE FROM annotations WHERE alert_hash = %q;", hash)

	statement, err := con.Prepare(q)
	if err != nil {
		return
	}

	statement.Exec()
	return
}
