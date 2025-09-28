package database

import (
	_ "modernc.org/sqlite"
)

type Entry struct {
	ID       int64
	Year     int
	Month    int
	Day      int
	Canon    int
	Income   int
	Expenses int
}
