// Package database contains the struct and functions to interact with the database
package database

import (
	_ "modernc.org/sqlite"
)

type Entry struct {
	ID       int64
	Year     int
	Month    int
	Day      string
	Canon    int
	IncomeM   int
	IncomeT   int
	Expenses int
}
