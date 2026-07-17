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
	canonRe        = regexp.MustCompile(`^canon \d+$`)
	dayNoWorkRe    = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) \d{1,2}\/\d{1,2}: *(0|-\d+)$`)
	dayWorkRe      = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) \d{1,2}\/\d{1,2}$`)
	morningRe      = regexp.MustCompile(`^m: *(?:-\d+|\d+(?:\+\d+)*(?:-\d+)?)$`)
	afternoonRe    = regexp.MustCompile(`^t: *(?:-\d+|\d+(?:\+\d+)*(?:-\d+)?)$`)
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
