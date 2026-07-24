package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/malklera/sliner/pkg/liner"
	_ "github.com/mattn/go-sqlite3"
)

var (
	canonRe     = regexp.MustCompile(`^canon \d+$`)
	dayNoWorkRe = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) \d{1,2}\/\d{1,2}: *(0|-\d+)$`)
	dayWorkRe   = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) \d{1,2}\/\d{1,2}$`)
	morningRe   = regexp.MustCompile(`^m:\s*(?:-?\d+|\d+(?:\+\d+)*(?:-\d+)?)$`)
	afternoonRe = regexp.MustCompile(`^t:\s*(?:-?\d+|\d+(?:\+\d+)*(?:-\d+)?)$`)

	// fileNameRe defines the valid filename format: `YYYY-MES-D.txt` where MES is
	// a Spanish month name (e.g., enero, febrero, ..., diciembre) and D is a
	// single digit (0-9). Example: 2024-enero-3.txt
	fileNameRe = regexp.MustCompile(`^(\d{4})-(enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre)-\d{1}\.txt$`)
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
		if err := createEnv(target); err != nil {
			fmt.Fprintf(os.Stderr, "error creating the needed directories at '%s': %v\n", target, err)
			os.Exit(1)
		}
		fmt.Printf("Enviroment successfully created at '%s'\n", target)
	case "format":
		formatCmd.Parse(os.Args[2:])
		err := formatNotes(target)
		switch err {
		case errNoFiles:
			fmt.Fprintf(os.Stderr, "There are no files to format.\n")
		case nil:
			return
		default:
			fmt.Fprintf(os.Stderr, "error formating notes at '%s': %v\n", target, err)
		}
		os.Exit(1)
	case "process":
		processCmd.Parse(os.Args[2:])
		if err := processNotes(target); err != nil {
			fmt.Fprintf(os.Stderr, "Error formating notes at `%s`: %v\n", target, err)
			os.Exit(1)
		}
	case "show":
		fmt.Println("show")
		showCmd.Parse(os.Args[2:])
		// if err := showEntries(target, year, month, day); err != nil {
		// 	fmt.Fprintf(os.Stderr, "showEntries(%s, %d, %d, %d): %v\n", target, year, month, day, err)
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

// listFiles return a slice of `file.Name()`
func listFiles(dir string) ([]string, error) {
	if dir != originalsDir && dir != formatedDir && dir != processedDir {
		return nil, errInvalidDir
	}
	allFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("os.ReadDir(%s): %w", dir, err)
	}

	textFiles := make([]string, 0, len(allFiles))

	for _, file := range allFiles {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			textFiles = append(textFiles, file.Name())
		}
	}

	if len(textFiles) == 0 {
		return nil, errNoFiles
	}

	return textFiles, nil
}
