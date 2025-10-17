// Package database contains the struct and functions to interact with the database
package database

import (
	"context"
	"database/sql"
	"log"
	_ "modernc.org/sqlite"
	"time"
)

type Entry struct {
	ID       int64
	Date     time.Time
	Canon    int
	IncomeM  int
	IncomeT  int
	Expenses int
}

// OpenDB create a DB and set up a schema if needed, returns a context and db
// to operate with, if there are error on the set-up it return it
func OpenDB() (context.Context, *sql.DB, error) {
	fileDB := "entries.db"
	db, err := sql.Open("sqlite", fileDB)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			log.Printf("Error closing '%s' : %v", fileDB, closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if pingErr := db.PingContext(ctx); pingErr != nil {
		return nil, nil, err
	}
	if err = createSchema(ctx, db); err != nil {
		return nil, nil, err
	}
	return ctx, db, nil
}

// createSchema creates the base table if it do not exist
func createSchema(ctx context.Context, db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS entry (
		id INTEGER PRIMARY KEY,
		date TEXT NOT NULL,
		canon INTEGER NOT NULL,
		incomeM INTEGER NOT NULL,
		incomeT INTEGER NOT NULL,
		expenses INTEGER NOT NULL,
	);`
	_, err := db.ExecContext(ctx, query)
	return err
}

// AddEntry accept an [database.Entry] and insert it to the given DB
func AddEntry(ctx context.Context, db *sql.DB, entry Entry) error {
	query := `
	INSERT INTO entry (
		date, canon, incomeM, incomeT, expenses) VALUES (?, ?, ?, ?, ?);`
	_, err := db.ExecContext(ctx, query, entry.Date, entry.Canon,
		entry.IncomeM, entry.IncomeT, entry.Expenses)
	if err != nil {
		return err
	}
	return nil
}

// ShowAll is a temporary function that returns all entries of the DB
func ShowAll(ctx context.Context, db *sql.DB) ([]Entry, error) {
	var entries []Entry
	query := `
	SELECT id, date, canon, incomeM, incomeT, expenses
	FROM entry;`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.Date, &entry.Canon, &entry.IncomeM, &entry.IncomeT, &entry.Expenses); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
