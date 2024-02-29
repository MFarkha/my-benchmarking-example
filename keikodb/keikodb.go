package keikodb

import (
	"database/sql"
	"log"

	"github.com/fatih/color"
	"github.com/mattn/go-sqlite3"
)

// Try to create a new database.
// Returns `true` if a new database was created, and `false` if it already exists
// and contains the appropriate tables.
//
// This function will abort the program if there is a failure while creating the database.
func MakeNew(db *sql.DB) bool {
	_, err := db.Exec(`
		CREATE TABLE hits (
			id     INTEGER PRIMARY KEY,
			name   TEXT UNIQUE,
			count  INTEGER
		);
	`)
	if err != nil {
		if sqlError, ok := err.(sqlite3.Error); ok {
			// code 1 == "table already exists"
			if sqlError.Code != 1 {
				log.Fatal(sqlError)
			} else {
				return false
			}
		} else {
			log.Fatal(err)
		}
	}
	return true
}

// Increments the number of hits by `1` for the given `name`.
// The `name` provided must be the name of the file as it exists
// on disk. Any non-existent entries will be added automatically
// and have a hit count set to `1`.
func IncrementHitCount(db *sql.DB, name string) error {
	_, err := db.Exec(`INSERT INTO
		hits(name, count) VALUES(?, 1)
    ON CONFLICT(name) DO UPDATE SET count = count + 1`, name)

	if err != nil {
		display := color.New(color.FgRed).SprintFunc()
		log.Printf("%s\n", display(err))
		return err
	}

	return nil
}

func SetHitCount(db *sql.DB, name string, count int) error {
	_, err := db.Exec(`INSERT INTO
		hits(name, count) VALUES(?, ?)
	ON CONFLICT(name) DO UPDATE SET count = ?`, name, count, count)

	if err != nil {
		display := color.New(color.FgRed).SprintFunc()
		log.Printf("%s\n", display(err))
		return err
	}
	return nil
}

func GetHits(db *sql.DB) (map[string]int, error) {
	rows, err := db.Query(`SELECT name,count FROM hits`)
	if err != nil {
		display := color.New(color.FgRed).SprintFunc()
		log.Printf("%s\n", display(err))
		return nil, err
	}
	results := make(map[string]int)

	for rows.Next() {
		var (
			name  string
			count int
		)
		err := rows.Scan(&name, &count)
		if err != nil {
			display := color.New(color.FgRed).SprintFunc()
			log.Printf("%s\n", display(err))
			return nil, err
		}
		results[name] = count
	}
	return results, nil
}
