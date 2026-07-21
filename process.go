package main

import (
	// "context"
	// "database/sql"
	"fmt"
	// "os"
	// "path/filepath"
	"strconv"
	"strings"
	// "time"

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
// 			return fmt.Errorf("os.ReadFile(%s): %w", orgNote, err)
// 		}
//
// 		entries, err := processNote(note, data)
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "error processing note '%s': %v\n", listNotes[n], err)
// 			continue
// 		}
// 		// TODO: check this function again, do i want to open the database for each note or for all?
// 		moveNote := true
// 		for _, e := range entries {
// 			if err := db.AddEntry(dbInstance, e); err != nil {
// 				fmt.Fprintf(os.Stderr, "error adding entry '%v' to the database: %v\n", e.Date, err)
// 				moveNote = false
// 				break
// 			}
// 		}
// 		if moveNote {
// 			if err := os.Rename(filepath.Join(formatedDir, listNotes[n]), filepath.Join(processedDir, listNotes[n])); err != nil {
// 				fmt.Fprintf(os.Stderr, "error moving formated note to the processed directory: %v\n", err)
// 			}
// 		}
// 	}
// 	return nil
// }

// processNote accept the name of a file, extract the data from it, return a
// slice of struct `Entry`, returns at any error, the notes are supposed
// to be formated
// func processNote(nameNote string, data []byte) ([]database.Entry, error) {
// 	fmt.Println()
// 	fmt.Println("Processing:", nameNote)
//
// 	content := strings.Split(string(data), "\n")
//
// 	year := fileNameRe.FindStringSubmatch(nameNote)[1]
//
// 	canon := int64(0)
//
// 	entries := make([]database.Entry, 0, 6)
// 	movements := make([]database.Movement, 0, 12)
// 	n := 0
// 	for n < len(content) {
// 		entryU7, err := uuid.NewV7()
// 		if err != nil {
// 			return nil, fmt.Errorf("uuid.NewV7(): %w", err)
// 		}
// 		entry := database.Entry{ID: entryU7, Canon: canon}
// 		movU7, err := uuid.NewV7()
// 		if err != nil {
// 			return nil, fmt.Errorf("uuid.NewV7(): %w", err)
// 		}
// 		movement := database.Movement{ID: movU7, EntryID: entry.ID}
// 		date := year + "-"
// 		switch {
// 		case canonRe.MatchString(content[n]):
// 			line := strings.Split(content[n], " ")
//
// 			c, err := strconv.ParseInt(line[1], 10, 64)
// 			if err != nil {
// 				return nil, fmt.Errorf("canonRe.MatchString(%s): strconv.ParseInt(%s, 10, 64): %w", content[n], line[1], err)
// 			}
// 			canon = c
// 			n++
// 		case dayWorkRe.MatchString(content[n]):
// 			_, cDate, _ := strings.Cut(content[n], " ")
// 			day, month, _ := strings.Cut(cDate, "/")
// 			date += month + "-" + day
// 			pDate, err := time.Parse(time.DateOnly, date)
// 			if err != nil {
// 				return nil, fmt.Errorf("dayWorkRe.MatchString(%s): time.Parse(%s, %s): %w", content[n], time.DateOnly, date, err)
// 			}
// 			entry.Date = pDate
// 			// Read day-date and advance counter
// 			n++
// 			morning, expensesM, err := processMovement(content[n])
// 			if err != nil {
// 				return nil, fmt.Errorf("dayWorkRe.MatchString(%s): processProcedings(%s): %w", content[n-1], content[n], err)
// 			}
// 			n++
// 			expensesT := 0
// 			entry.IncomeT, expensesT, err = processMovement(content[n])
// 			if err != nil {
// 				return nil, fmt.Errorf("dayWorkRe.MatchString(%s): processProcedings(%s): %w", content[n-2], content[n], err)
// 			}
// 			n++
// 			entry.Expenses = expensesM + expensesT
// 			entries = append(entries, entry)
// 		default:
// 			return nil, fmt.Errorf("line '%s' of file '%s' has the wrong format", content[n], nameNote)
// 		}
// 	}
// 	return entries, nil
// }

// processMovement take in a valid string of `morningRe` or `afternoonRe`,
// and extract its values
func processMovement(entryID uuid.UUID, content string) ([]database.Movement, error) {
	var items []database.Movement
	line := strings.Split(content, ":")
	shift := line[0]

	// m:0
	// t:2000
	// t:2000+3000+2500
	// m:-4500
	// t:2000+3000+2500-3300
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
