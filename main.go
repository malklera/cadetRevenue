package main

import (
	"bufio"
	"cadetRevenue/db"
	"errors"
	"flag"
	"fmt"
	"github.com/malklera/sliner/pkg/liner"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	fileNameRe     = regexp.MustCompile(`^(enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre)-\d{1}-(\d{4})\.txt$`)
	canonRe        = regexp.MustCompile(`^canon \d+$`)
	dayNoWorkRe    = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) \d{1,2}\/\d{1,2}: *(0|-\d+)$`)
	dayWorkRe      = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) \d{1,2}\/\d{1,2}$`)
	dayWorkCanonRe = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) (\d{1,2}\/\d{1,2}) (canon \d+)$`)
	procedingsRe   = regexp.MustCompile(`^(m|t): *(?:-\d+|\d+(?:\+\d+)*(?:-\d+)?)$`)
)

// Indicates that there are no .txt files on the current directory
var errNoFiles = errors.New("there are no files to process")

// Indicates that the user canceled the renaming of a file
var errRenameCancel = errors.New("renaming canceled")

// Indicates skiping the formatting of the note
var errSkipNote = errors.New("skip formatting of note")

// Indicates that the given directory is invalid
var errInvalidDir = errors.New("the given directory is invalid")

var originalsDir = "originals"
var formatedDir = "formated"
var processedDir = "processed"

var reader = bufio.NewReader(os.Stdin)

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
		// TODO: have to create the db too, or just try this way and see what happens
	case "format":
		formatCmd.Parse(os.Args[2:])
		err := formatNotes(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formating notes at '%s': %v", target, err)
			os.Exit(1)
		}
	case "process":
		processCmd.Parse(os.Args[2:])
		err := processNotes(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formating notes at `%s`: %v", target, err)
			os.Exit(1)
		}
	case "show":
		fmt.Println("show")
		showCmd.Parse(os.Args[2:])
		// thinkg what to do here
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
	// TODO: if there is an error, delete the created directories
	err := os.MkdirAll(target, 0777)
	if err != nil {
		return fmt.Errorf("os.MkdirAll(%s, 0777): %w", target, err)
	}

	op := filepath.Join(target, originalsDir)
	err = os.Mkdir(op, 0777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("os.Mkdir(%s), 0777: %w", op, err)
	}

	fp := filepath.Join(target, formatedDir)
	err = os.Mkdir(fp, 0777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("os.Mkdir(%s), 0777: %w", fp, err)
	}

	pp := filepath.Join(target, processedDir)
	err = os.Mkdir(pp, 0777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
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
		note, err := validFileName(notes[n])
		switch err {
		case nil:
			err := formatNote(note)
			switch err {
			case nil:
				// NOTE: break or continue?
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

// listFiles return a slice of [file.Name()]
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

// validFileName take the name of a file, check that it is the correct format, if not ask the
// user for input, return a correctly formated file name
func validFileName(file string) (string, error) {
	line := liner.NewLiner()
	defer line.Close()

	currentFileName := ""
	renameFor := true

	// TODO: refactor this
	for renameFor {
		currentFileName = file
		for {
			//check the fileName to be the correct format
			if fileNameRe.MatchString(currentFileName) {
				break
			} else {
				// NOTE: Should ask the user if it want to rename the file?
				fmt.Println()
				fmt.Printf("'%s' is not a valid file name\n", currentFileName)
				fmt.Println("The correct format is: month-int-year.txt")
				fmt.Println("Where 'month' is a valid month written in Spanish word")
				fmt.Println("Where 'int' is a number from 0 to 9")
				fmt.Println("Where 'year' is a number from 0000 to 9999")
				fmt.Printf("> ")

				input, err := line.PrefilledInput(currentFileName, -1)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error on input: %v\n", err)
				} else {
					if _, err := os.Stat(filepath.Join(originalsDir, input)); err == nil {
						fmt.Printf("File name '%s' already exist, input a different one\n", input)
					} else if !errors.Is(err, fs.ErrNotExist) {
						fmt.Fprintf(os.Stderr, "error checking if file '%s' exist: %v\n", input, err)
					} else {
						currentFileName = input
					}
				}
			}
		}

		if file == currentFileName {
			// use break instead
			renameFor = false
		} else {
			retry := true
			for retry {
				if err := os.Rename(filepath.Join(originalsDir, file), filepath.Join(originalsDir, currentFileName)); err != nil {
					fmt.Fprintf(os.Stderr, "error renaming file '%s' to '%s': %v\n", file, currentFileName, err)
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
							return file, errRenameCancel
						default:
							fmt.Fprintf(os.Stderr, "'%s' is an invalid option.\n", opt)
						}
					}
				} else {
					fmt.Println()
					fmt.Printf("File '%s' succesfull renamed to '%s'\n", file, currentFileName)
					retry = false
					renameFor = false
				}
			}
		}
	}
	return currentFileName, nil
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
	// TODO: replace this for another thing
	content := strings.Split(strings.ToLower(string(data)), "\n")

	// the .Split leave me with a final empty string element
	if content[len(content)-1] == "" {
		content = content[:len(content)-1]
	}

	newContent, lineN := validFirstLine(nameNote, content)

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
				// TODO: this is wrong
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
				// TODO: Colaps all break into one case
			case procedingsRe.MatchString(content[lineN+1]):
				break
			case canonRe.MatchString(content[lineN+1]):
				break
			case dayNoWorkRe.MatchString(content[lineN+1]):
				break
			case dayWorkRe.MatchString(content[lineN+1]):
				break
			case dayWorkCanonRe.MatchString(content[lineN+1]):
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
				} else {
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
func validFirstLine(nameNote string, content []string) (string, int) {
	newContent := ""
	lineN := 0

	for {
		// TODO: use return more often instead of if/else
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
			for {
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
			}
		case "2":
			line := liner.NewLiner()
			defer line.Close()
			for {
				input, err := line.PrefilledInput(content[lineN], -1)
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
			}
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
	parts := strings.Split(date, "/")
	d, err := strconv.Atoi(parts[0])
	// TODO: put both checks in the same if
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
func processNote(nameNote string) ([]db.Entry, error) {
	fmt.Println()
	fmt.Println("Processing:", nameNote)

	orgNote := filepath.Join(formatedDir, nameNote)
	data, err := os.ReadFile(orgNote)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile(%s): %w", orgNote, err)
	}
	content := strings.Split(string(data), "\n")

	// TODO: when i change the name of the file, this has to be updated
	// validate the name first
	// Get the year
	year := fileNameRe.FindStringSubmatch(nameNote)[2]

	canon := 0

	entries := make([]db.Entry, 0, 6)
	n := 0
	for n < len(content) {
		var entry db.Entry
		entry.Canon = canon
		date := year + "-"
		switch {
		case canonRe.MatchString(content[n]):
			line := strings.Split(content[n], " ")

			canon, err = strconv.Atoi(line[1])
			if err != nil {
				return nil, fmt.Errorf("canonRe.MatchString(%s): strconv.Atoi(%s): %w", content[n], line[1], err)
			}
			n++
		case dayWorkRe.MatchString(content[n]):
			_, cDate, _ := strings.Cut(content[n], " ")
			day, month, _ := strings.Cut(cDate, "/")
			date += month + "-" + day
			entry.Date, err = time.Parse(time.DateOnly, date)
			if err != nil {
				return nil, fmt.Errorf("dayWorkRe.MatchString(%s): time.Parse(%s, %s): %w", content[n], time.DateOnly, date, err)
			}
			n++
			// here process the procedings and advance the counter
			expensesM := 0
			entry.IncomeM, expensesM, err = processProcedings(content[n])
			if err != nil {
				return nil, fmt.Errorf("dayWorkRe.MatchString(%s): processProcedings(%s): %w", content[n-1], content[n], err)
			}
			n++
			expensesT := 0
			entry.IncomeT, expensesT, err = processProcedings(content[n])
			if err != nil {
				return nil, fmt.Errorf("dayWorkRe.MatchString(%s): processProcedings(%s): %w", content[n-2], content[n], err)
			}
			n++
			entry.Expenses = expensesM + expensesT
			entries = append(entries, entry)
		default:
			return nil, fmt.Errorf("line '%s' of file '%s' has the wrong format", content[n], nameNote)
		}
	}
	return entries, nil
}

// processProcedings take in a valid string of `procedingsRe`, returns income,
// expenses, error
func processProcedings(content string) (int, int, error) {
	// NOTE: i am currently loosing information, should i return a []int for
	// procedings and expenses?
	line := strings.Split(content, ":")
	lineP := strings.TrimSpace(line[1])

	procedings := 0
	expenses := 0
	// TODO: refactor this into a recursive function that returns []int
	switch {
	case strings.Contains(lineP, "-"):
		hasExp := strings.Split(lineP, "-")
		// NOTE: only allow for one -int, do not need to calculate len
		exp, err := strconv.Atoi(hasExp[len(hasExp)-1])
		if err != nil {
			return 0, 0, err
		}
		expenses = exp * -1
		if len(hasExp) > 1 && hasExp[0] != "" {
			for p := range strings.SplitSeq(hasExp[0], "+") {
				proc, err := strconv.Atoi(p)
				if err != nil {
					return 0, 0, err
				}
				procedings += proc
			}
		}
	case strings.Contains(lineP, "+"):
		for p := range strings.SplitSeq(lineP, "+") {
			proc, err := strconv.Atoi(p)
			if err != nil {
				return 0, 0, err
			}
			procedings += proc
		}
	default:
		proc, err := strconv.Atoi(lineP)
		if err != nil {
			return 0, 0, err
		}
		procedings += proc
	}
	return procedings, expenses, nil
}

// processNotes process all notes in `formatedDir`, only returns an error if there
// was a problem with `listFiles()`.
func processNotes(target string) error {
	path := filepath.Join(target, formatedDir)
	listNotes, err := listFiles(path)
	if err != nil {
		return fmt.Errorf("listFiles(%s): %w", path, err)
	}

	for n := range listNotes {
		entries, err := processNote(listNotes[n])
		if err != nil {
			// TODO: use continue instead of if/else
			fmt.Fprintf(os.Stderr, "error processing note '%s': %v", listNotes[n], err)
		} else {
			// TODO: check this function again, do i want to open the database for each note or for all?
			dbInstance, err := db.New("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "error opening the database: %v", err)
			} else {
				defer dbInstance.Close()
				for _, e := range entries {
					if err := db.AddEntry(dbInstance, e); err != nil {
						fmt.Fprintf(os.Stderr, "error adding entry '%v' to the database: %v", e.Date, err)
					} else {
						if err := os.Rename(filepath.Join(formatedDir, listNotes[n]), filepath.Join(processedDir, listNotes[n])); err != nil {
							fmt.Fprintf(os.Stderr, "error moving formated note to the processed directory: %v", err)
						}
					}
				}
			}
		}
	}
	return nil
}
