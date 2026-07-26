package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cadetRevenue/internal/database"
)

// TODO: this is wrong, i should be able to do it in the db

// showAll print to stdout a formated list of all year, month, day availables
// in the database
func showAll() error {
	ctx := context.Background()
	db, err := sql.Open("sqlite3", "entries.db")
	if err != nil {
		return fmt.Errorf("sql.Open(\"sqlite3\", \"entries.db\"): %w", err)
	}
	defer db.Close()

	queries := database.New(db)

	entries, err := getAllEntries(ctx, queries)
	if err != nil {
		return fmt.Errorf("getAllEntries(ctx, queries): %w", err)
	}

	for _, e := range entries {
		fmt.Println(e)
	}

	dates, err := getAllDates(ctx, queries)
	if err != nil {
		return fmt.Errorf("getAllDates(ctx, queries): %w", err)
	}

	for _, d := range dates {
		fmt.Println(d)
	}
	return nil
}

func getAllDates(ctx context.Context, queries *database.Queries) ([]time.Time, error) {
	dates, err := queries.ListAvailableDates(ctx)
	if err != nil {
		return nil, err
	}
	return dates, err
}

func getAllEntries(ctx context.Context, queries *database.Queries) ([]database.Entry, error) {
	entries, err := queries.ListAllEntries(ctx)
	if err != nil {
		return nil, err
	}
	return entries, err

}
