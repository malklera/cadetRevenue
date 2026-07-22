package main

import (
	// "context"
	// "database/sql"
	"fmt"
	// "os"
	// "path/filepath"
	"strconv"
	"strings"
	"time"

	"cadetRevenue/internal/database"
	uuid "github.com/gofrs/uuid/v5"
)

// processNotes process all notes in `formatedDir`, move the notes correctly
// formated to `processedDir`.
// func processNotes(target string) error {
// 	path := filepath.Join(target, formatedDir)
// 	listNotes, err := listFiles(path)
// 	if err != nil {
// 		return fmt.Errorf("listFiles(%s): %w", path, err)
// 	}
//
// 	ctx := context.Background()
// 	db, err := sql.Open("sqlite3", "entries.db")
// 	if err != nil {
// 		return fmt.Errorf("sql.Open(\"sqlite3\", \"entries.db\"): %w", err)
// 	}
// 	defer db.Close()
//
// 	for n, note := range listNotes {
// 		orgNote := filepath.Join(formatedDir, note)
// 		data, err := os.ReadFile(orgNote)
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "os.ReadFile(%s): %w", orgNote, err)
// 			continue
// 		}
//
// 		entries, movements, err := processNote(note, data)
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "error processing note '%s': %v\n", listNotes[n], err)
// 			continue
// 		}
//
// 		if moveNote {
// 			if err := os.Rename(filepath.Join(formatedDir, listNotes[n]), filepath.Join(processedDir, listNotes[n])); err != nil {
// 				fmt.Fprintf(os.Stderr, "error moving formated note to the processed directory: %v\n", err)
// 			}
// 		}
// 	}
// 	return nil
// }

// processNote accept the name of a file and its content, extract the data from it,
// return the entries and their respective movements, returns at any error,
// the notes are supposed to be formated
func processNote(nameNote string, data []byte) ([]database.Entry, []database.Movement, error) {
	fmt.Println()
	fmt.Println("Processing:", nameNote)

	content := strings.Split(string(data), "\n")

	year := fileNameRe.FindStringSubmatch(nameNote)[1]

	canon := int64(0)

	entries := make([]database.Entry, 0, 6)
	movements := make([]database.Movement, 0, 12)
	n := 0
	for n < len(content) {
		entryU7, err := uuid.NewV7()
		if err != nil {
			return nil, nil, fmt.Errorf("uuid.NewV7(): %w", err)
		}
		entry := database.Entry{ID: entryU7, Canon: canon}

		date := year + "-"

		switch {
		case canonRe.MatchString(content[n]):
			line := strings.Split(content[n], " ")

			c, err := strconv.ParseInt(line[1], 10, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("canonRe.MatchString(%s): strconv.ParseInt(%s, 10, 64): %w", content[n], line[1], err)
			}
			canon = c
			n++
		case dayWorkRe.MatchString(content[n]):
			_, cDate, _ := strings.Cut(content[n], " ")
			day, month, _ := strings.Cut(cDate, "/")
			date += month + "-" + day
			pDate, err := time.Parse(time.DateOnly, date)
			if err != nil {
				return nil, nil, fmt.Errorf("dayWorkRe.MatchString(%s): time.Parse(%s, %s): %w", content[n], time.DateOnly, date, err)
			}
			entry.Date = pDate
			// Read day-date and advance counter
			n++
			morning, err := processMovement(entry.ID, content[n])
			if err != nil {
				return nil, nil, fmt.Errorf("dayWorkRe.MatchString(%s): processProcedings(%s): %w", content[n-1], content[n], err)
			}
			n++
			afternoon, err := processMovement(entry.ID, content[n])
			if err != nil {
				return nil, nil, fmt.Errorf("dayWorkRe.MatchString(%s): processProcedings(%s): %w", content[n-2], content[n], err)
			}
			n++
			// TODO: calculate profit
			entries = append(entries, entry)
			movements = append(movements, morning...)
			movements = append(movements, afternoon...)
		default:
			return nil, nil, fmt.Errorf("line '%s' of file '%s' has the wrong format", content[n], nameNote)
		}
	}
	return entries, movements, nil
}

// processMovement take in a valid string of `morningRe` or `afternoonRe`,
// and extract its values
func processMovement(entryID uuid.UUID, content string) ([]database.Movement, error) {
	var items []database.Movement
	line := strings.Split(content, ":")
	shift := line[0]

	switch {
	case strings.Contains(line[1], "-"):
		hasExp := strings.Split(line[1], "-")
		// NOTE: only allow for one -int, do not need to calculate len
		// i do have to do it because of m:-4500
		exp, err := strconv.ParseInt(hasExp[len(hasExp)-1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("strings.Contains(%s, \"-\"): strconv.ParseInt(%s, 10, 64): %w", line[1], hasExp[len(hasExp)-1], err)
		}

		procU7, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("uuid.NewV7(): %w", err)
		}

		proc := database.Movement{ID: procU7, EntryID: entryID, Shift: shift, Amount: exp * -1}
		items = append(items, proc)
		if len(hasExp) > 1 && hasExp[0] != "" {
			for p := range strings.SplitSeq(hasExp[0], "+") {
				amount, err := strconv.ParseInt(p, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("strings.Contains(%s, \"-\"): strconv.ParseInt(%s, 10, 64): %w", line[1], p, err)
				}

				procU7, err := uuid.NewV7()
				if err != nil {
					return nil, fmt.Errorf("uuid.NewV7(): %w", err)
				}

				proc := database.Movement{ID: procU7, EntryID: entryID, Shift: shift, Amount: amount}
				items = append(items, proc)
			}
		}
	case strings.Contains(line[1], "+"):
		for p := range strings.SplitSeq(line[1], "+") {
			amount, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("strings.Contains(%s, \"+\"): strconv.ParseInt(%s, 10, 64): %w", line[1], p, err)
			}

			procU7, err := uuid.NewV7()
			if err != nil {
				return nil, fmt.Errorf("uuid.NewV7(): %w", err)
			}

			proc := database.Movement{ID: procU7, EntryID: entryID, Shift: shift, Amount: amount}
			items = append(items, proc)
		}
	default:
		amount, err := strconv.ParseInt(line[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("default: strconv.ParseInt(%s, 10, 64): %w", line[1], err)
		}

		procU7, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("uuid.NewV7(): %w", err)
		}

		proc := database.Movement{ID: procU7, EntryID: entryID, Shift: shift, Amount: amount}
		items = append(items, proc)
	}
	return items, nil
}

func calcProfit(canon int64, morning []database.Movement, afternoon []database.Movement) float64 {
	expenses := int64(0)
	income := int64(0)
	for _, i := range morning {
		if i.Amount > int64(0) {
			income += i.Amount
		} else {
			expenses += i.Amount
		}
	}

	for _, i := range afternoon {
		if i.Amount > int64(0) {
			income += i.Amount
		} else {
			expenses += i.Amount
		}
	}

	if income > 0 {
		if income < canon*4 {
			sub := float64(income - (income / 4))
			return float64(income) - sub - float64(expenses)
		} else {
			return float64(canon) - float64(expenses)
		}
	}
	return float64(expenses)
}

// func saveNote(ctx context.Context, db *sql.DB, queries *database.Queries, entries []database.Entry, movements []database.Movement) error {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()
// 	qtx := queries.WithTx(tx)
// 	for _, e := range entries {
// 		err := qtx.CreateEntry(ctx, database.CreateEntryParams{
// 			ID:     e.ID,
// 			Date:   e.Date,
// 			Canon:  e.Canon,
// 			Profit: e.Profit,
// 		})
// 	}
// }
