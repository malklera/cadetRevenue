package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cadetRevenue/internal/database"
)

func showProfitMonth(date time.Time, nextDate time.Time) error {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", "entries.db")
	if err != nil {
		return fmt.Errorf("sql.Open(\"sqlite3\", \"entries.db\"): %w", err)
	}
	defer db.Close()

	queries := database.New(db)

	profit, err := getProfitMonth(ctx, queries, date, nextDate)
	if err != nil {
		return fmt.Errorf("getProfitMonth(ctx, queries, %v, %v): %w", date, nextDate, err)
	}
	fmt.Println("Month\t\tProfit")
	fmt.Printf("%s\t%.2f\n", date.Format(time.DateOnly), profit.Float64)
	return nil
}

func getProfitMonth(ctx context.Context, queries *database.Queries, month time.Time, nextMonth time.Time) (sql.NullFloat64, error) {
	profit, err := queries.ProfitMonth(ctx, database.ProfitMonthParams{Date: month, Date_2: nextMonth})
	if err != nil {
		return sql.NullFloat64{}, err
	}
	if !profit.Valid {
		return sql.NullFloat64{}, fmt.Errorf("no profits in %v", month)
	}
	return profit, nil
}

func showProfitDay(day time.Time) error {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", "entries.db")
	if err != nil {
		return fmt.Errorf("sql.Open(\"sqlite3\", \"entries.db\"): %w", err)
	}
	defer db.Close()

	queries := database.New(db)

	profit, err := getProfitDay(ctx, queries, day)
	if err != nil {
		return fmt.Errorf("getProfitDay(ctx, db, %v): %w", day, err)
	}
	fmt.Println("Day\t\tProfit")
	fmt.Printf("%s\t%.2f\n", day.Format(time.DateOnly), profit)
	return nil
}

func getProfitDay(ctx context.Context, queries *database.Queries, day time.Time) (float64, error) {
	profit, err := queries.ProfitDay(ctx, day)
	if err != nil {
		return 0, err
	}
	return profit, nil
}
