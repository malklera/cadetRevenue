package main

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"time"

	"cadetRevenue/internal/database"
)

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

	dates, err := getAllDates(ctx, queries)
	if err != nil {
		return fmt.Errorf("getAllDates(ctx, queries): %w", err)
	}

	years := []int{}
	months := []time.Month{}
	for _, d := range dates {
		if !slices.Contains(years, d.Year()) {
			years = append(years, d.Year())
			months = slices.Delete(months, 0, len(months))

			fmt.Println()
			fmt.Print(d.Year())
		}

		if !slices.Contains(months, d.Month()) {
			months = append(months, d.Month())
			fmt.Println()
			fmt.Printf("\t%v", d.Month())
			fmt.Println()
			fmt.Printf("\t\t")
		}

		fmt.Printf("%d, ", d.Day())
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
