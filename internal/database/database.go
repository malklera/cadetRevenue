package database

import (
	"fmt"
	"log"
	"context"
	_ "modernc.org/sqlite"
	"database/sql"
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
