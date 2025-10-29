// Package database contains the struct and functions to interact with the database
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"time"
)

type Entry struct {
	ID       int64     `db:"id"`
	Date     time.Time `db:"date"`
	Canon    int       `db:"canon"`
	IncomeM  int       `db:"incomeM"`
	IncomeT  int       `db:"incomeT"`
	Expenses int       `db:"expenses"`
}

const (
	fileDB = "internal/database/entries.db"
)

// New create a DB and set up a schema if needed, returns db
// to operate with, if there are error on the set-up it returns it
func New() (*sql.DB, error) {
	db, err := sql.Open("sqlite", fileDB)
	if err != nil {
		return nil, fmt.Errorf("error on sql.Open(): %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if pingErr := db.PingContext(ctx); pingErr != nil {
		return nil, fmt.Errorf("error on db.PringContext(): %w", err)
	}

	if err = createSchema(db); err != nil {
		return nil, fmt.Errorf("error on createSchema(): %w", err)
	}
	return db, nil
}

// createSchema creates the base table if it do not exist
func createSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS entry (
		id INTEGER PRIMARY KEY,
		date TEXT NOT NULL,
		canon INTEGER NOT NULL,
		incomeM INTEGER NOT NULL,
		incomeT INTEGER NOT NULL,
		expenses INTEGER NOT NULL
	);`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// AddEntry accept an [database.Entry] and insert it to the given DB
func AddEntry(db *sql.DB, entry Entry) error {
	query := `
	INSERT INTO entry (
		date, canon, incomeM, incomeT, expenses) VALUES (?, ?, ?, ?, ?);`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	dateStr := entry.Date.Format(time.DateTime)
	_, err := db.ExecContext(ctx, query, dateStr, entry.Canon,
		entry.IncomeM, entry.IncomeT, entry.Expenses)
	if err != nil {
		return err
	}
	return nil
}

// ShowAll is a temporary function that returns all entries of the DB
func ShowAll(db *sql.DB) ([]Entry, error) {
	var entries []Entry
	query := `
	SELECT id, date, canon, incomeM, incomeT, expenses
	FROM entry;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry Entry
		dateStr := ""
		if err := rows.Scan(&entry.ID, &dateStr, &entry.Canon, &entry.IncomeM, &entry.IncomeT, &entry.Expenses); err != nil {
			return nil, err
		}
		tempDate, err := time.Parse(time.DateTime, dateStr)
		if err != nil {
			log.Printf("error parsing date '%s' of row '%d'", dateStr, entry.ID)
		} else {
			entry.Date = tempDate
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// GetYears return all available years on the database
func GetYears(db *sql.DB) ([]string, error) {
	query := `
		select distinct strftime('%Y', date)
		from entry
		order by date;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var years []string
	for rows.Next() {
		year := ""
		if err := rows.Scan(&year); err != nil {
			return nil, err
		}
		years = append(years, year)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return years, nil
}

// GetMonths return all available months of a given year
func GetMonths(db *sql.DB, year string) ([]string, error) {
	// WARN: erase this print
	fmt.Println("year:", year)
	query := `
		select distinct strftime('%m', date)
		from entry
		where strftime('%Y', date) = ?
		order by date;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var months []string
	for rows.Next() {
		month := ""
		if err := rows.Scan(&month); err != nil {
			return nil, err
		}
		// WARN: erase this print
		fmt.Println("month:", month)
		months = append(months, month)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return months, nil
}
