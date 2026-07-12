package main

import (
	"bufio"
	// "context"
	// "database/sql"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
		// TODO: When updating the db.go have the creation of the db here
		// run the goose migration here??
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

// createEnv takes a path and creates the originalsDir, formatedDir, processedDir
func createEnv(target string) error {
	err := os.MkdirAll(target, 0777)
	if err != nil {
		t := os.Remove(target)
		if t != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", target, t)
		}
		return fmt.Errorf("os.MkdirAll(%s, 0777): %w", target, err)
	}

	op := filepath.Join(target, originalsDir)
	err = os.Mkdir(op, 0777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		t := os.Remove(target)
		if t != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", target, t)
		}
		o := os.Remove(op)
		if o != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", op, o)
		}
		return fmt.Errorf("os.Mkdir(%s), 0777: %w", op, err)
	}

	fp := filepath.Join(target, formatedDir)
	err = os.Mkdir(fp, 0777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		t := os.Remove(target)
		if t != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", target, t)
		}
		o := os.Remove(op)
		if o != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", op, o)
		}
		f := os.Remove(fp)
		if f != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", fp, f)
		}
		return fmt.Errorf("os.Mkdir(%s), 0777: %w", fp, err)
	}

	pp := filepath.Join(target, processedDir)
	err = os.Mkdir(pp, 0777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		t := os.Remove(target)
		if t != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", target, t)
		}
		o := os.Remove(op)
		if o != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", op, o)
		}
		f := os.Remove(fp)
		if f != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", fp, f)
		}
		p := os.Remove(pp)
		if p != nil {
			fmt.Fprintf(os.Stderr, "os.Remove(%s): %v", pp, p)
		}
		return fmt.Errorf("os.Mkdir(%s), 0777: %w", pp, err)
	}
	return nil
}

// formatNotes validates file names in `originalsDir` and call `checkFormatNote`
// on each, only returns an error in case of failing to `listFiles()`
func formatNotes(target string) error {
	path := filepath.Join(target, originalsDir)
	notes, err := listFiles(path)
	if err != nil {
		return fmt.Errorf("listFiles(%s): %w", path, err)
	}
	for n := range notes {
		note, err := validFileName(notes[n], linerInput, fileStat, fileRename, bufio.NewReader(os.Stdin))
		switch err {
		case nil:
			err := formatNote(note)
			switch err {
			case nil:
				break
			case errSkipNote:
				fmt.Printf("Formating of '%s' skipped\n", note)
			default:
				fmt.Fprintf(os.Stderr, "error checking the format of '%s': %v\n", note, err)
			}
		case errRenameCancel:
			fmt.Printf("The renaming of '%s' was canceled\n", note)
		default:
			fmt.Fprintf(os.Stderr, "error checking the name of '%s': %v\n", note, err)
		}
	}
	return nil
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

// NOTE: Maybe i should reformat checkFormatNote

// formatNote accept the name of a file, allow to modify or cancel the modification, return
// nil if all operations were execute correctly
func formatNote(nameNote string) error {
	orgNote := filepath.Join(originalsDir, nameNote)
	data, err := os.ReadFile(orgNote)
	if err != nil {
		return fmt.Errorf("os.ReadFile(%s): %w", orgNote, err)
	}
	fmt.Println()
	fmt.Println("Formating:", nameNote)

	// Fix potential panic when checking empty files
	if string(data) == "" {
		fmt.Printf("'%s' is empty\n", nameNote)
		return errSkipNote
	}
	content := strings.Split(strings.ToLower(string(data)), "\n")

	// the .Split leave me with a final empty string element
	if content[len(content)-1] == "" {
		content = content[:len(content)-1]
	}

	newContent, lineN := validFirstLine(nameNote, content, linerInput, bufio.NewReader(os.Stdin))

	reader := bufio.NewReader(os.Stdin)

	// check each line after the first, for non-valid ones allow user to erase or modify
	for lineN < len(content) {
		switch {
		case content[lineN] == "":
			lineN++
		case canonRe.MatchString(content[lineN]):
			newContent += content[lineN] + "\n"
			lineN++
		case dayNoWorkRe.MatchString(content[lineN]):
			day := strings.Split(content[lineN], ":")
			newContent += addPadding(day[0]) + "\n"
			newContent += "m:" + day[1] + "\n"
			newContent += "t:0" + "\n"
			switch {
			case lineN+1 == len(content):
				newContent, _ = strings.CutSuffix(newContent, "\n")
			case canonRe.MatchString(content[lineN+1]), dayNoWorkRe.MatchString(content[lineN+1]),
				dayWorkRe.MatchString(content[lineN+1]), dayWorkCanonRe.MatchString(content[lineN+1]):
				break
			default:
				// error, the next line is invalid
				lineN = lineN + nextLineInvalid(nameNote, content[lineN], content[lineN+1])
			}
			lineN++
		case dayWorkRe.MatchString(content[lineN]):
			newContent += addPadding(content[lineN]) + "\n"
			nC, j := fillNoEntry(nameNote, content, newContent, lineN)
			newContent += nC
			lineN += j
			lineN++
		case dayWorkCanonRe.MatchString(content[lineN]):
			subStrings := dayWorkCanonRe.FindStringSubmatch(content[lineN])
			newContent += subStrings[3] + "\n" + addPadding(subStrings[1]+" "+subStrings[2]) + "\n"
			nC, j := fillNoEntry(nameNote, content, newContent, lineN)
			newContent += nC
			lineN += j
			lineN++
		case procedingsRe.MatchString(content[lineN]):
			newContent += content[lineN] + "\n"
			switch {
			case lineN+1 == len(content):
				newContent += "t:0"
			case procedingsRe.MatchString(content[lineN+1]), canonRe.MatchString(content[lineN+1]),
				dayNoWorkRe.MatchString(content[lineN+1]), dayWorkRe.MatchString(content[lineN+1]),
				dayWorkCanonRe.MatchString(content[lineN+1]):
				break
			default:
				// error, the next line is invalid
				lineN = lineN + nextLineInvalid(nameNote, content[lineN], content[lineN+1])
			}
			lineN++
		default:
			// Non valid line
			proceed := true
			for proceed {
				fmt.Println()
				fmt.Println("File:", nameNote)
				fmt.Println("Current line:")
				fmt.Println(content[lineN])
				fmt.Println("The line is invalid")
				fmt.Println("Choose what to do")
				fmt.Println("1- Erase line")
				fmt.Println("2- Modify")
				fmt.Println("3- Skip note (need to manually change something about the note)")
				fmt.Print("> ")
				opt, err := reader.ReadString('\n')
				if err != nil {
					fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
					continue
				}
				opt = strings.TrimSpace(opt)
				switch opt {
				case "1":
					proceed = false
				case "2":
					line := liner.NewLiner()
					defer line.Close()
					for {
						fmt.Println("Modify the line and press Enter")
						input, err := line.PrefilledInput(content[lineN], -1)
						if err != nil {
							fmt.Fprintf(os.Stderr, "error on input: %v\n", err)
							continue
						}
						if validLine(input) {
							newContent += input + "\n"
							if lineN == len(content)-1 {
								newContent += "t:0"
							}
							break
						}
						fmt.Printf("'%s'\n is not a valid line\n", input)
					}
					proceed = false
				case "3":
					return errSkipNote
				default:
					fmt.Fprintf(os.Stderr, "'%s' is an invalid option.\n", opt)
				}
			}
			lineN++
		}
	}

	// Move the file

	// TODO: this has to be its own function

	// Loop to os.CreateTemp()
	for {
		tempFile, err := os.CreateTemp(formatedDir, nameNote)
		tempName := tempFile.Name()
		formatedNote := filepath.Join(formatedDir, nameNote)

		if err != nil {
			fmt.Println()
			fmt.Println("File:", tempName)
			fmt.Fprintf(os.Stderr, "error creating temporary file: %v\n", err)
			fmt.Println("Do you want to retry? (y/n)")
			fmt.Print("> ")
			opt, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
			} else {
				opt = strings.TrimSpace(opt)
				switch opt {
				case "y", "Y":
					break
				case "n", "N":
					return errSkipNote
				default:
					fmt.Fprintf(os.Stderr, "'%s' is an invalid option.\n", opt)
				}
			}
		} else {
			defer os.Remove(tempName)
			for {
				if _, err := tempFile.Write([]byte(newContent)); err != nil {
					fmt.Println()
					fmt.Println("File:", tempName)
					fmt.Fprintf(os.Stderr, "error writing to the temporary file: %v\n", err)
					fmt.Println("Do you want to retry? (y/n)")
					fmt.Print("> ")
					opt, err := reader.ReadString('\n')
					if err != nil {
						fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
					} else {
						opt = strings.TrimSpace(opt)
						switch opt {
						case "y", "Y":
							break
						case "n", "N":
							return errSkipNote
						default:
							fmt.Fprintf(os.Stderr, "'%s' is an invalid option.\n", opt)
						}
					}
				} else {
					// data was writen
					if err := tempFile.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "error closing temporary file after writing: %v", err)
					} else {
						for {
							if err := os.Rename(tempName, formatedNote); err != nil {
								fmt.Println()
								fmt.Println("File:", tempName)
								fmt.Fprintf(os.Stderr, "error renaming '%s' to '%s': %v\n", tempName, formatedNote, err)
								fmt.Println("Do you want to retry? (y/n)")
								fmt.Print("> ")
								opt, err := reader.ReadString('\n')
								if err != nil {
									fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
								} else {
									opt = strings.TrimSpace(opt)
									switch opt {
									case "y", "Y":
										break
									case "n", "N":
										return errSkipNote
									default:
										fmt.Fprintf(os.Stderr, "'%s' is an invalid option.\n", opt)
									}
								}
							} else {
								// os.Rename was successfull
								// has to Remove originals/nameNote
								for {
									if err := os.Remove(orgNote); err != nil {
										fmt.Println()
										fmt.Println("File:", orgNote)
										fmt.Fprintf(os.Stderr, "error removing '%s': %v\n", orgNote, err)
										fmt.Println("Do you want to retry? (y/n)")
										fmt.Print("> ")
										opt, err := reader.ReadString('\n')
										if err != nil {
											fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
										} else {
											opt = strings.TrimSpace(opt)
											switch opt {
											case "y", "Y":
												break
											case "n", "N":
												return errSkipNote
											default:
												fmt.Fprintf(os.Stderr, "'%s' is an invalid option.\n", opt)
											}
										}
									} else {
										// os.Remove successfull
										return nil
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// validFirstLine ensures the first line of a note is the canon.
func validFirstLine(nameNote string,
	content []string,
	readInput func(string) (string, error),
	reader *bufio.Reader,
) (string, int) {
	newContent := ""
	lineN := 0

	for {
		if canonRe.MatchString(content[lineN]) {
			return content[lineN] + "\n", (lineN + 1)
		}
		if content[lineN] == "" {
			lineN++
			continue
		}
		fmt.Println()
		fmt.Println("File:", nameNote)
		fmt.Println("Current first line:")
		fmt.Println(content[lineN])
		fmt.Println("Next line:")
		fmt.Println(content[lineN+1])
		fmt.Println("Choose operation")
		fmt.Println("1- Add line above")
		fmt.Println("2- Edit line")
		fmt.Println("3- Erase line")
		fmt.Print("> ")
		opt, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
			continue
		}
		opt = strings.TrimSpace(opt)
		switch opt {
		case "1":
			// NOTE: How is the user suppose to know what the canon should be?
			fmt.Println("New line:")
			fmt.Print("> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
				continue
			}
			line = strings.TrimSpace(line)
			if canonRe.MatchString(line) {
				newContent += line + "\n"
				lineN++
				return newContent, lineN
			}
			fmt.Fprintf(os.Stderr, "'%s' is an invalid line\n", line)
		case "2":
			input, err := readInput(content[lineN])
			if err != nil {
				fmt.Fprintf(os.Stderr, "error on input: %v\n", err)
				continue
			}
			if canonRe.MatchString(input) {
				newContent += input + "\n"
				lineN++
				return newContent, lineN
			}
			fmt.Fprintf(os.Stderr, "'%s' is an invalid line\n", input)
		case "3":
			lineN++
		default:
			fmt.Fprintf(os.Stderr, "'%s' is an invalid option.\n", opt)
		}
	}
}

// nextLineInvalid prompt choosing if erasing the line below(return 1) or
// leaving it for later modification(return 0)
func nextLineInvalid(nameNote string, cl string, nl string) int {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println()
		fmt.Println("File:", nameNote)
		fmt.Println("Current line:")
		fmt.Println(cl)
		fmt.Println("The line below is invalid")
		fmt.Println(nl)
		fmt.Println("Choose what to do")
		fmt.Println("1- Erase line")
		fmt.Println("2- Leave it(will be prompted to modify it later)")
		fmt.Print("> ")
		opt, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		} else {
			opt = strings.TrimSpace(opt)
			switch opt {
			case "1":
				return 1
			case "2":
				return 0
			default:
				fmt.Fprintf(os.Stderr, "'%s' is an invalid option.\n", opt)
			}
		}
	}
}

// fillNoEntry fill with 0 if there are no procedings in the next line, or prompt
// for input if `nextLineInvalid`
func fillNoEntry(nameNote string, content []string, newContent string, n int) (string, int) {
	switch {
	case n+1 > len(content):
		fmt.Println()
		fmt.Println("File:", nameNote)
		fmt.Println("Current line:")
		fmt.Println(content[n])
		fmt.Println("There are no entries for procedings, will be filled with 0")
		newContent += "m:0" + "\n"
		newContent += "t:0" + "\n"
		return newContent, 0
	case procedingsRe.MatchString(content[n+1]):
		return "", 0
	default:
		return "", nextLineInvalid(nameNote, content[n], content[n+1])
	}
}

// addPadding takes a 'Day n/m', returns it zero padded
func addPadding(day string) string {
	parts := strings.Split(day, " ")
	formatDay := parts[0] + " "
	dates := strings.Split(parts[1], "/")
	if len(dates[0]) < 2 {
		formatDay += "0" + dates[0] + "/"
	} else {
		formatDay += dates[0] + "/"
	}
	if len(dates[1]) < 2 {
		formatDay += "0" + dates[1]
	} else {
		formatDay += dates[1]
	}
	return formatDay
}

// validLine evaluate if the given line conform to any of the declared regex's
func validLine(line string) bool {
	switch {
	case canonRe.MatchString(line), procedingsRe.MatchString(line):
		return true
	case dayNoWorkRe.MatchString(line), dayWorkRe.MatchString(line),
		dayWorkCanonRe.MatchString(line):
		day := strings.Split(line, " ")
		date, _, _ := strings.Cut(day[1], ":")
		return validDate(date)
	default:
		return false
	}
}

// validDate evaluate if the given day `Day dd/mm` has sensible max numbers, dd < 31
// and mm < 13
func validDate(date string) bool {
	// TODO: change this to check for real dates instead
	parts := strings.Split(date, "/")
	d, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	if 1 > d || d > 31 {
		return false
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	if 1 > m || m > 12 {
		return false
	}
	return true
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
