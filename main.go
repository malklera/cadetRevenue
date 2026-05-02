package main

import (
	"bufio"
	"cadetRevenue/db"
	"errors"
	"flag"
	"fmt"
	"github.com/malklera/sliner/pkg/liner"
	"io/fs"
	"log"
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
	setUp := ""
	format := false
	process := false

	// Each flag added has to be added to flagPassed and added to flag.Visit
	flagPassed := map[string]bool{"setup": false, "format": false, "process": false}

	flag.StringVar(&setUp, "setup", "", "Create needed directories and the db.")
	flag.StringVar(&setUp, "s", "", "Create needed directories and the db. (shorthand)")

	flag.BoolVar(&format, "format", false, "Format notes.")
	flag.BoolVar(&format, "f", false, "Format notes. (shorthand)")

	flag.BoolVar(&process, "process", false, "Process notes.")
	flag.BoolVar(&process, "p", false, "Process notes. (shorthand)")

	// flag.StringVar(&show, "show", "", "Show entries. ")

	flag.Parse()
	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(0)
	}

	if flag.NArg() > 0 {
		log.Println("Invalid arguments:")
		fmt.Println(flag.Args())
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}

	flag.Visit(func(fn *flag.Flag) {
		switch f := fn.Name; f {
		case "setup":
			flagPassed["setup"] = true
		case "s":
			flagPassed["setup"] = true
		case "format":
			flagPassed["format"] = true
		case "f":
			flagPassed["format"] = true
		case "process":
			flagPassed["process"] = true
		case "p":
			flagPassed["process"] = true
		default:
			break
		}
	})

	// setUp can only be on its own, can't really format if i still did not copy the notes
	// flatPassed["setUp"] == true && (flagPassed["format"] != false || flagPassed["process"] != false)
	// error

	// format and process can be together, format has to run first

	switch {
	case flagPassed["setup"]:
		if flagPassed["format"] || flagPassed["process"] {
			fmt.Println("Error. -setup <path> can only be on its own.")
			flag.Usage()
			os.Exit(1)
		}

		targetDir, err := createEnv(setUp)
		if err != nil {
			fmt.Printf("Error creating the environment : %v", err)
			os.Exit(1)
		}

		_, err = db.New(targetDir)
		if err != nil {
			fmt.Printf("Error creating the database : %v", err)
			os.Exit(1)
		}

		fmt.Println("Setup succesfull.")
	case flagPassed["format"]:
		notes, err := listFiles(originalsDir)
		switch err {
		case nil:
			for n := range notes {
				note, err := checkFileName(notes[n])
				switch err {
				case nil:
					err := checkFormatNote(note)
					switch err {
					case nil:
						break
					case errSkipNote:
						fmt.Printf("Formating of '%s' skipped\n", note)
					default:
						log.Printf("error checking the format of '%s' : %v\n", note, err)
					}
				case errRenameCancel:
					fmt.Printf("The renaming of '%s' was canceled\n", note)
				default:
					log.Printf("error checking the name of '%s' : %v\n", note, err)
				}
			}
			if flagPassed["process"] {
				// process notes here
			}
		case errNoFiles:
			fmt.Println("There are no files to format.")
			os.Exit(0)
		default:
			log.Printf("error listFiles(%s) : %v\n", originalsDir, err)
		}
	case flagPassed["process"]:
		listNotes, err := listFiles(formatedDir)
		switch err {
		case nil:
			for n := range listNotes {
				entry, err := processNote(listNotes[n])
				if err != nil {
					log.Printf("error processing note '%s' : %v", listNotes[n], err)
				} else {
					dbInstance, err := db.New("")
					moveFile := false
					if err != nil {
						log.Printf("error opening the database: %v", err)
					} else {
						defer dbInstance.Close()
						for _, e := range entry {
							if err := db.AddEntry(dbInstance, e); err != nil {
								log.Printf("error adding entry '%v' to the database: %v", e.Date, err)
							} else {
								moveFile = true
							}
						}
					}
					// move the file from formated to processed
					if moveFile {
						if err := os.Rename(filepath.Join(formatedDir, listNotes[n]), filepath.Join(processedDir, listNotes[n])); err != nil {
							log.Printf("error moving formated note to the processed directory: %v", err)
						}
					}
				}
			}
		case errNoFiles:
			fmt.Println("There are no files to process")
		default:
			log.Printf("error listing files: %v\n", err)
		}
	default:
		log.Fatalf("Invalid flag.")
	}
}

// createEnv takes a path and creates the originalsDir, formatedDir, processedDir, return
// any error encountered.
func createEnv(target string) (string, error) {
	var err error
	if target == "" {
		target, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	err = os.MkdirAll(target, 0777)
	if err != nil {
		return target, fmt.Errorf("os.MkdirAll(%s, 0777) : %w", target, err)
	}

	op := filepath.Join(target, originalsDir)
	err = os.Mkdir(op, 0777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return target, fmt.Errorf("os.Mkdir(%s), 0777 : %w", op, err)
	}

	fp := filepath.Join(target, formatedDir)
	err = os.Mkdir(fp, 0777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return target, fmt.Errorf("os.Mkdir(%s), 0777 : %w", fp, err)
	}

	pp := filepath.Join(target, processedDir)
	err = os.Mkdir(pp, 0777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return target, fmt.Errorf("os.Mkdir(%s), 0777 : %w", pp, err)
	}

	return target, nil
}

// listFiles return a slice of [file.Name()]
func listFiles(dir string) ([]string, error) {
	if dir != originalsDir && dir != formatedDir && dir != processedDir {
		return nil, errInvalidDir
	}
	allFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf(" > os.ReadDir(%s) : %w", dir, err)
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

// checkFileName take the name of a file, check that it is the correct format, if not ask the
// user for input, return a correctly formated file name
func checkFileName(file string) (string, error) {
	line := liner.NewLiner()
	defer line.Close()

	currentFileName := ""
	renameFor := true

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
					log.Printf("error on input: %v\n", err)
				} else {
					if _, err := os.Stat(filepath.Join(originalsDir, input)); err == nil {
						fmt.Printf("File name '%s' already exist, input a different one\n", input)
					} else if !errors.Is(err, fs.ErrNotExist) {
						log.Printf("error checking if file '%s' exist: %v\n", input, err)
					} else {
						currentFileName = input
					}
				}
			}
		}

		if file == currentFileName {
			renameFor = false
		} else {
			retry := true
			for retry {
				if err := os.Rename(filepath.Join(originalsDir, file), filepath.Join(originalsDir, currentFileName)); err != nil {
					log.Printf("error renaming file '%s' to '%s': %v\n", file, currentFileName, err)
					fmt.Println("Do you want to retry? (y/n)")
					fmt.Print("> ")
					opt, err := reader.ReadString('\n')
					if err != nil {
						log.Printf("error reading input: %v\n", err)
					} else {
						opt = strings.TrimSpace(opt)
						switch opt {
						case "y", "Y":
							break
						case "n", "N":
							return file, errRenameCancel
						default:
							fmt.Printf("'%s' is an invalid option.\n", opt)
						}
					}
				} else {
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

// checkFormatNote accept the name of a file, allow to modify or cancel the modification, return
// nil if all operations were execute correctly
func checkFormatNote(nameNote string) error {
	orgNote := filepath.Join(originalsDir, nameNote)
	data, err := os.ReadFile(orgNote)
	if err != nil {
		return fmt.Errorf("checkFormatNote(%s) > os.ReadFile(%s) : %w", nameNote, orgNote, err)
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
	newContent := ""
	n := 0

	for {
		if canonRe.MatchString(content[n]) {
			newContent += content[n] + "\n"
			n++
			break
		} else {
			if content[n] == "" {
				n++
			} else {
				fmt.Println()
				fmt.Println("File:", nameNote)
				fmt.Println("Current first line:")
				fmt.Println(content[n])
				fmt.Println("Choose operation")
				fmt.Println("1- Add line above")
				fmt.Println("2- Edit line")
				fmt.Println("3- Erase line")
				fmt.Print("> ")
				opt, err := reader.ReadString('\n')
				if err != nil {
					log.Printf("error reading input: %v\n", err)
				} else {
					opt = strings.TrimSpace(opt)
					switch opt {
					case "1":
						// NOTE: How is the user suppose to know what the canon should be?
						for {
							fmt.Println("New line:")
							fmt.Print("> ")
							line, err := reader.ReadString('\n')
							if err != nil {
								log.Printf("error reading input: %v\n", err)
							} else {
								line = strings.TrimSpace(line)
								if canonRe.MatchString(line) {
									newContent += line + "\n"
									break
								} else {
									fmt.Printf("'%s' is an invalid line\n", line)
								}
							}
						}
					case "2":
						line := liner.NewLiner()
						defer line.Close()
						for {
							input, err := line.PrefilledInput(content[n], -1)
							if err != nil {
								log.Printf("error on input: %v\n", err)
							} else {
								if canonRe.MatchString(input) {
									newContent += input + "\n"
									break
								} else {
									fmt.Printf("'%s' is an invalid line\n", input)
								}
							}
						}
					case "3":
						n++
					default:
						fmt.Printf("'%s' is an invalid option.\n", opt)
					}
				}
			}
		}
	}

	// check each line after the first, for non-valid ones allow user to erase or modify
	for n < len(content) {
		switch {
		case content[n] == "":
			n++
		case canonRe.MatchString(content[n]):
			newContent += content[n] + "\n"
			n++
		case dayNoWorkRe.MatchString(content[n]):
			day := strings.Split(content[n], ":")
			newContent += checkPadding(day[0]) + "\n"
			newContent += "m:" + day[1] + "\n"
			newContent += "t:0" + "\n"
			switch {
			case n+1 == len(content):
				newContent, _ = strings.CutSuffix(newContent, "\n")
			case canonRe.MatchString(content[n+1]):
				break
			case dayNoWorkRe.MatchString(content[n+1]):
				break
			case dayWorkRe.MatchString(content[n+1]):
				break
			case dayWorkCanonRe.MatchString(content[n+1]):
				break
			default:
				// error, the next line is invalid
				proceed := true
				for proceed {
					fmt.Println()
					fmt.Println("File:", nameNote)
					fmt.Println("Current line:")
					fmt.Println(content[n])
					fmt.Println("The line below is invalid")
					fmt.Println(content[n+1])
					fmt.Println("Choose what to do")
					fmt.Println("1- Erase line")
					fmt.Println("2- Leave it(will be prompted to modify it later)")
					fmt.Print("> ")
					opt, err := reader.ReadString('\n')
					if err != nil {
						log.Printf("error reading input: %v\n", err)
					} else {
						opt = strings.TrimSpace(opt)
						switch opt {
						case "1":
							// Advance the counter, jump over the next line
							n++
							proceed = false
						case "2":
							proceed = false
						default:
							fmt.Printf("'%s' is an invalid option.\n", opt)
						}
					}
				}
			}
			n++
		case dayWorkRe.MatchString(content[n]):
			newContent += checkPadding(content[n]) + "\n"
			switch {
			case n+1 > len(content):
				fmt.Println()
				fmt.Println("File:", nameNote)
				fmt.Println("Current line:")
				fmt.Println(content[n])
				fmt.Println("There are no entries for procedings, will be filled with 0")
				newContent += "M:0" + "\n"
				newContent += "T:0" + "\n"
			case procedingsRe.MatchString(content[n+1]):
				break
			default:
				// error, the next line is invalid
				proceed := true
				for proceed {
					fmt.Println()
					fmt.Println("File:", nameNote)
					fmt.Println("Current line:")
					fmt.Println(content[n])
					fmt.Println("The line below is invalid")
					fmt.Println(content[n+1])
					fmt.Println("Choose what to do")
					fmt.Println("1- Erase line")
					fmt.Println("2- Leave it(will be prompted to modify it later)")
					fmt.Print("> ")
					opt, err := reader.ReadString('\n')
					if err != nil {
						log.Printf("error reading input: %v\n", err)
					} else {
						opt = strings.TrimSpace(opt)
						switch opt {
						case "1":
							// Advance the counter, jump over the next line
							n++
							proceed = false
						case "2":
							proceed = false
						default:
							fmt.Printf("'%s' is an invalid option.\n", opt)
						}
					}
				}
			}
			n++
		case dayWorkCanonRe.MatchString(content[n]):
			subStrings := dayWorkCanonRe.FindStringSubmatch(content[n])
			newContent += subStrings[3] + "\n" + checkPadding(subStrings[1]+" "+subStrings[2]) + "\n"
			switch {
			case n+1 > len(content):
				fmt.Println()
				fmt.Println("File:", nameNote)
				fmt.Println("Current line:")
				fmt.Println(content[n])
				fmt.Println("There are no entries for procedings, will be filled with 0")
				newContent += "m:0" + "\n"
				newContent += "t:0" + "\n"
			case procedingsRe.MatchString(content[n+1]):
				break
			default:
				// error, the next line is invalid
				proceed := true
				for proceed {
					fmt.Println()
					fmt.Println("File:", nameNote)
					fmt.Println("Current line:")
					fmt.Println(content[n])
					fmt.Println("The line below is invalid")
					fmt.Println(content[n+1])
					fmt.Println("Choose what to do")
					fmt.Println("1- Erase line")
					fmt.Println("2- Leave it(will be prompted to modify it later)")
					fmt.Print("> ")
					opt, err := reader.ReadString('\n')
					if err != nil {
						log.Printf("error reading input: %v\n", err)
					} else {
						opt = strings.TrimSpace(opt)
						switch opt {
						case "1":
							// Advance the counter, jump over the next line
							n++
							proceed = false
						case "2":
							proceed = false
						default:
							fmt.Printf("'%s' is an invalid option.\n", opt)
						}
					}
				}
			}
			n++

		case procedingsRe.MatchString(content[n]):
			newContent += content[n] + "\n"
			switch {
			case n+1 == len(content):
				newContent += "T:0"
			case procedingsRe.MatchString(content[n+1]):
				break
			case canonRe.MatchString(content[n+1]):
				break
			case dayNoWorkRe.MatchString(content[n+1]):
				break
			case dayWorkRe.MatchString(content[n+1]):
				break
			case dayWorkCanonRe.MatchString(content[n+1]):
				break
			default:
				// error, the next line is invalid
				proceed := true
				for proceed {
					fmt.Println()
					fmt.Println("File:", nameNote)
					fmt.Println("Current line:")
					fmt.Println(content[n])
					fmt.Println("The line below is invalid")
					fmt.Println(content[n+1])
					fmt.Println("Choose what to do")
					fmt.Println("1- Erase line")
					fmt.Println("2- Leave it(will be prompted to modify it later)")
					fmt.Print("> ")
					opt, err := reader.ReadString('\n')
					if err != nil {
						log.Printf("error reading input: %v\n", err)
					} else {
						opt = strings.TrimSpace(opt)
						switch opt {
						case "1":
							// Advance the counter, jump over the next line
							n++
							proceed = false
						case "2":
							proceed = false
						default:
							fmt.Printf("'%s' is an invalid option.\n", opt)
						}
					}
				}
			}
			n++
		default:
			// Non valid line
			proceed := true
			for proceed {
				fmt.Println()
				fmt.Println("File:", nameNote)
				fmt.Println("Current line:")
				fmt.Println(content[n])
				fmt.Println("The line is invalid")
				fmt.Println("Choose what to do")
				fmt.Println("1- Erase line")
				fmt.Println("2- Modify")
				fmt.Println("3- Skip note (need to manually change something about the note)")
				fmt.Print("> ")
				opt, err := reader.ReadString('\n')
				if err != nil {
					log.Printf("error reading input: %v\n", err)
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
							input, err := line.PrefilledInput(content[n], -1)
							if err != nil {
								log.Printf("error on input: %v\n", err)
							} else if validLine(input) {
								newContent += input + "\n"
								break
							} else {
								fmt.Printf("'%s'\n is not a valid line\n", input)
							}
						}
						proceed = false
					case "3":
						return errSkipNote
					default:
						fmt.Printf("'%s' is an invalid option.\n", opt)
					}
				}
			}
			n++
		}
	}

	// Move the file

	// Loop to os.CreateTemp()
	for {
		tempFile, err := os.CreateTemp(formatedDir, nameNote)
		tempName := tempFile.Name()
		formatedNote := filepath.Join(formatedDir, nameNote)

		if err != nil {
			fmt.Println()
			fmt.Println("File:", tempName)
			log.Printf("error creating temporary file: %v\n", err)
			fmt.Println("Do you want to retry? (y/n)")
			fmt.Print("> ")
			opt, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("error reading input: %v\n", err)
			} else {
				opt = strings.TrimSpace(opt)
				switch opt {
				case "y", "Y":
					break
				case "n", "N":
					return errSkipNote
				default:
					fmt.Printf("'%s' is an invalid option.\n", opt)
				}
			}
		} else {
			defer os.Remove(tempName)
			for {
				if _, err := tempFile.Write([]byte(newContent)); err != nil {
					fmt.Println()
					fmt.Println("File:", tempName)
					log.Printf("error writing to the temporary file: %v\n", err)
					fmt.Println("Do you want to retry? (y/n)")
					fmt.Print("> ")
					opt, err := reader.ReadString('\n')
					if err != nil {
						log.Printf("error reading input: %v\n", err)
					} else {
						opt = strings.TrimSpace(opt)
						switch opt {
						case "y", "Y":
							break
						case "n", "N":
							return errSkipNote
						default:
							fmt.Printf("'%s' is an invalid option.\n", opt)
						}
					}
				} else {
					// data was writen
					if err := tempFile.Close(); err != nil {
						log.Printf("error closing temporary file after writing: %v", err)
					} else {
						for {
							if err := os.Rename(tempName, formatedNote); err != nil {
								fmt.Println()
								fmt.Println("File:", tempName)
								log.Printf("error renaming '%s' to '%s' : %v\n", tempName, formatedNote, err)
								fmt.Println("Do you want to retry? (y/n)")
								fmt.Print("> ")
								opt, err := reader.ReadString('\n')
								if err != nil {
									log.Printf("error reading input: %v\n", err)
								} else {
									opt = strings.TrimSpace(opt)
									switch opt {
									case "y", "Y":
										break
									case "n", "N":
										return errSkipNote
									default:
										fmt.Printf("'%s' is an invalid option.\n", opt)
									}
								}
							} else {
								// os.Rename was successfull
								// has to Remove originals/nameNote
								for {
									if err := os.Remove(orgNote); err != nil {
										fmt.Println()
										fmt.Println("File:", orgNote)
										log.Printf("error removing '%s' : %v\n", orgNote, err)
										fmt.Println("Do you want to retry? (y/n)")
										fmt.Print("> ")
										opt, err := reader.ReadString('\n')
										if err != nil {
											log.Printf("error reading input: %v\n", err)
										} else {
											opt = strings.TrimSpace(opt)
											switch opt {
											case "y", "Y":
												break
											case "n", "N":
												return errSkipNote
											default:
												fmt.Printf("'%s' is an invalid option.\n", opt)
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

// checkPadding adds 0 pading to single digit dates
func checkPadding(day string) string {
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
	case canonRe.MatchString(line), dayNoWorkRe.MatchString(line),
		dayWorkRe.MatchString(line), procedingsRe.MatchString(line):
		return true
	default:
		return false
	}
}

// processNote accept the name of a file, extract the data from it, return a
// slice of struct [Entry], at any error it gets returned, the notes are supposed
// to be formated
func processNote(nameNote string) ([]db.Entry, error) {
	fmt.Println()
	fmt.Println("Processing:", nameNote)

	// Get the year
	date := fileNameRe.FindStringSubmatch(nameNote)
	year := date[2]

	orgNote := filepath.Join(formatedDir, nameNote)
	data, err := os.ReadFile(orgNote)
	if err != nil {
		return nil, err
	}
	content := strings.Split(string(data), "\n")

	canon := 0

	entries := make([]db.Entry, 0, 6)
	n := 0
	for n < len(content) {
		var entry db.Entry
		entry.Canon = canon
		dateS := year + "-"
		switch {
		case canonRe.MatchString(content[n]):
			line := strings.Split(content[n], " ")

			canon, err = strconv.Atoi(line[1])
			if err != nil {
				return nil, err
			}
			n++
		case dayWorkRe.MatchString(content[n]):
			line := strings.Split(content[n], " ")
			date := strings.Split(line[1], "/")
			dateS += date[1] + "-" + date[0]
			entry.Date, err = time.Parse(time.DateOnly, dateS)
			if err != nil {
				return nil, err
			}
			n++
			// here process the procedings and advance the counter
			expensesM := 0
			entry.IncomeM, expensesM, err = processProcedings(content[n])
			if err != nil {
				return nil, err
			}
			n++
			expensesT := 0
			entry.IncomeT, expensesT, err = processProcedings(content[n])
			if err != nil {
				return nil, err
			}
			entry.Expenses = expensesM + expensesT
			n++
			entries = append(entries, entry)
		default:
			return nil, fmt.Errorf("line '%s' of file '%s' has the wrong format", content[n], nameNote)
		}
	}
	return entries, nil
}

// processProcedings take in a valid string of procedingsRe, returns income,
// expenses, error
func processProcedings(content string) (int, int, error) {
	line := strings.Split(content, ":")
	lineP := strings.TrimSpace(line[1])

	procedings := 0
	expenses := 0
	switch {
	case strings.Contains(lineP, "-"):
		hasExp := strings.Split(lineP, "-")
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
