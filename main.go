package main

import (
	// "context"
	// "database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	// "time"

	"cadetRevenue/internal/database"
	"github.com/malklera/sliner/pkg/liner"
	_ "github.com/mattn/go-sqlite3"
)

var (
	canonRe        = regexp.MustCompile(`^canon \d+$`)
	dayNoWorkRe    = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) \d{1,2}\/\d{1,2}: *(0|-\d+)$`)
	dayWorkRe      = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) \d{1,2}\/\d{1,2}$`)
	dayWorkCanonRe = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) (\d{1,2}\/\d{1,2}) (canon \d+)$`)
	procedingsRe   = regexp.MustCompile(`^(m|t): *(?:-\d+|\d+(?:\+\d+)*(?:-\d+)?)$`)
)

var (
	// Indicates that there are no .txt files on the current directory
	errNoFiles = errors.New("there are no files to process")
	// Indicates that the user canceled the renaming of a file
	errRenameCancel = errors.New("renaming canceled")
	// Indicates skiping the formatting of the note
	errSkipNote = errors.New("skip formatting of note")
	// Indicates that the given directory is invalid
	errInvalidDir = errors.New("the given directory is invalid")
)

const (
	originalsDir = "originals"
	formatedDir  = "formated"
	processedDir = "processed"
)

// linerInput acepts a string and allow the user to change it
var linerInput = func(current string) (string, error) {
	line := liner.NewLiner()
	defer line.Close()
	return line.PrefilledInput(current, -1)
}

func main() {
	target := "."
	year := 0
	month := 0
	day := 0

	setupCmd := flag.NewFlagSet("setup", flag.ExitOnError)
	setupCmd.StringVar(&target, "target", ".", "Create the needed directories and the database at --target.")
	setupCmd.StringVar(&target, "t", ".", "Create the needed directories and the database at -t.")

	formatCmd := flag.NewFlagSet("format", flag.ExitOnError)
	formatCmd.StringVar(&target, "target", ".", "Format the notes at '--target/originals'.")
	formatCmd.StringVar(&target, "t", ".", "Format the notes at '-t/originals'. (shorthand)")

	processCmd := flag.NewFlagSet("process", flag.ExitOnError)
	processCmd.StringVar(&target, "target", ".", "Extract data from '--target/formated' and save to the database.")
	processCmd.StringVar(&target, "t", ".", "Extract data from '-t/formated' and save to the database. (shorthand)")

	showCmd := flag.NewFlagSet("show", flag.ExitOnError)

	showCmd.StringVar(&target, "target", ".", "Show data from '--target/entries.db'.")
	showCmd.StringVar(&target, "t", ".", "Show data from '-t/entries.db'. (shorthand)")

	showCmd.IntVar(&year, "year", 0, "Year to show, use two or four numbers. Empty or 0 show all available.")
	showCmd.IntVar(&year, "y", 0, "Year to show, use two or four numbers. Empty or 0 show all available. (shorthand)")

	showCmd.IntVar(&month, "month", 0, "Month to show, use one or two digits. Empty or 0 show all available.")
	showCmd.IntVar(&month, "m", 0, "Month to show, use one or two digits. Empty or 0 show all available. (shorthand)")

	showCmd.IntVar(&day, "day", 0, "Day to show, use numbers, one or two digits. Empty or 0 show all available.")
	showCmd.IntVar(&day, "d", 0, "Day to show, use numbers, one or two digits. Empty or 0 show all available. (shorthand)")

	if len(os.Args) < 2 {
		setupCmd.Usage()
		formatCmd.Usage()
		processCmd.Usage()
		showCmd.Usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "setup":
		// NOTE: Do i error when passing not valid flags or just ignore?
		setupCmd.Parse(os.Args[2:])
		err := createEnv(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating the needed directories at '%s': %v\n", target, err)
			os.Exit(1)
		}
		fmt.Printf("Enviroment successfully created at '%s'\n", target)
	case "format":
		formatCmd.Parse(os.Args[2:])
		err := formatNotes(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formating notes at '%s': %v", target, err)
			os.Exit(1)
		}
	case "process":
		processCmd.Parse(os.Args[2:])
		// err := processNotes(target)
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Error formating notes at `%s`: %v", target, err)
		// 	os.Exit(1)
		// }
	case "show":
		fmt.Println("show")
		showCmd.Parse(os.Args[2:])
		// err := showEntries(target, year, month, day)
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "showEntries(%s, %d, %d, %d)", target, year, month, day)
		// 	os.Exit(1)
		// }
	default:
		fmt.Fprintln(os.Stderr, "wrong sub-command.")
		setupCmd.Usage()
		formatCmd.Usage()
		processCmd.Usage()
		showCmd.Usage()
		os.Exit(1)
	}
}

// processNote accept the name of a file, extract the data from it, return a
// slice of struct `Entry`, returns at any error, the notes are supposed
// to be formated
// func processNote(nameNote string) ([]database.Entry, error) {
// 	fmt.Println()
// 	fmt.Println("Processing:", nameNote)
//
// 	orgNote := filepath.Join(formatedDir, nameNote)
// 	data, err := os.ReadFile(orgNote)
// 	if err != nil {
// 		return nil, fmt.Errorf("os.ReadFile(%s): %w", orgNote, err)
// 	}
// 	content := strings.Split(string(data), "\n")
//
// 	// TODO: when i change the name of the file, this has to be updated
// 	// validate the name first
// 	// Get the year
// 	year := fileNameRe.FindStringSubmatch(nameNote)[2]
//
// 	canon := int64(0)
//
// 	entries := make([]database.Entry, 0, 6)
// 	movements := make([]database.Movement, 0, 12)
// 	n := 0
// 	for n < len(content) {
// 		var entry database.Entry
// 		var movement database.Movement
// 		entry.Canon = canon
// 		date := year + "-"
// 		switch {
// 		case canonRe.MatchString(content[n]):
// 			line := strings.Split(content[n], " ")
//
// 			canon, err = strconv.ParseInt(line[1], 10, 64)
// 			if err != nil {
// 				return nil, fmt.Errorf("canonRe.MatchString(%s): strconv.Atoi(%s): %w", content[n], line[1], err)
// 			}
// 			n++
// 		case dayWorkRe.MatchString(content[n]):
// 			// CONTINUE HERE
// 			_, cDate, _ := strings.Cut(content[n], " ")
// 			day, month, _ := strings.Cut(cDate, "/")
// 			date += month + "-" + day
// 			entry.Date, err = time.Parse(time.DateOnly, date)
// 			if err != nil {
// 				return nil, fmt.Errorf("dayWorkRe.MatchString(%s): time.Parse(%s, %s): %w", content[n], time.DateOnly, date, err)
// 			}
// 			n++
// 			// here process the procedings and advance the counter
// 			expensesM := 0
// 			entry.IncomeM, expensesM, err = processProcedings(content[n])
// 			if err != nil {
// 				return nil, fmt.Errorf("dayWorkRe.MatchString(%s): processProcedings(%s): %w", content[n-1], content[n], err)
// 			}
// 			n++
// 			expensesT := 0
// 			entry.IncomeT, expensesT, err = processProcedings(content[n])
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

// processProcedings take in a valid string of `procedingsRe`, and extract its values
func processProcedings(content string, entryID int64) ([]database.Movement, error) {
	var procedings []database.Movement
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
			return nil, err
		}
		proc := database.Movement{EntryID: entryID, Shift: shift, Amount: exp * -1}
		procedings = append(procedings, proc)
		if len(hasExp) > 1 && hasExp[0] != "" {
			for p := range strings.SplitSeq(hasExp[0], "+") {
				amount, err := strconv.ParseInt(p, 10, 64)
				if err != nil {
					return nil, err
				}
				proc := database.Movement{EntryID: entryID, Shift: shift, Amount: amount}
				procedings = append(procedings, proc)
			}
		}
	case strings.Contains(line[1], "+"):
		for p := range strings.SplitSeq(line[1], "+") {
			amount, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				return nil, err
			}
			proc := database.Movement{EntryID: entryID, Shift: shift, Amount: amount}
			procedings = append(procedings, proc)
		}
	default:
		amount, err := strconv.ParseInt(line[1], 10, 64)
		if err != nil {
			return nil, err
		}
		proc := database.Movement{EntryID: entryID, Shift: shift, Amount: amount}
		procedings = append(procedings, proc)
	}
	return procedings, nil
}

// processNotes process all notes in `formatedDir`, move the notes correctly
// formated to `processedDir`.
// func processNotes(target string) error {
// 	path := filepath.Join(target, formatedDir)
// 	listNotes, err := listFiles(path)
// 	if err != nil {
// 		return fmt.Errorf("listFiles(%s): %w", path, err)
// 	}
// 	ctx := context.Background()
// 	db, err := sql.Open("sqlite3", "entries.db")
// 	if err != nil {
// 		return fmt.Errorf("sql.Open(\"sqlite3\", \"entries.db\"): %w", err)
// 	}
// 	defer db.Close()
//
// 	for n := range listNotes {
// 		entries, err := processNote(listNotes[n])
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

// func showEntries(target string, year int, month int, day int) error {
// 	dbInstance, err := db.New(target)
// 	if err != nil {
// 		return fmt.Errorf("db.New(%s): %v", target, err)
// 	}
// 	// TODO: has to change this
// 	entries, err := db.ShowAll(dbInstance)
// 	if err != nil {
// 		return fmt.Errorf("db.ShowAll(dbInstance): %v", err)
// 	}
// 	displayEntries(entries)
// 	return nil
// }

// displayEntries pretty print to stdout the given entry
// func displayEntries(entries []db.Entry) {
// 	fmt.Println("Date	Canon	In. Morning	In. Afternoon	Expenses")
// 	for _, e := range entries {
// 		fmt.Printf("%v	%d	%d	%d	%d\n", e.Date, e.Canon, e.IncomeM, e.IncomeT, e.Expenses)
// 	}
// }
