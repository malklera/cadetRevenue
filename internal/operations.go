// Package internal has all functions used to interact with the program
package internal

import (
	// "cadetRevenue/internal/database"
	"errors"
	"fmt"
	"github.com/malklera/sliner/pkg/liner"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	fileNameRe   = regexp.MustCompile(`^(enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre)-\d{1}-(\d{4})\.txt$`)
	canonRe      = regexp.MustCompile(`^Canon \d+$`)
	dayNoWorkRe  = regexp.MustCompile(`^(Lunes|Martes|Miércoles|Miercoles|Jueves|Viernes|Sábado|Sabado) \d{1,2}\/\d{1,2}: *(0|-\d+)$`)
	dayWorkRe    = regexp.MustCompile(`^(Lunes|Martes|Miércoles|Miercoles|Jueves|Viernes|Sábado|Sabado) \d{1,2}\/\d{1,2}$`)
	procedingsRe = regexp.MustCompile(`^(M|T): *(?:-\d+|\d+(?:\+\d+)*(?:-\d+)?)$`)
)

var errNoFiles = errors.New("there are no files to process")

func checkFileNames() error {
	allFiles, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("error listing files: %w", err)
	}

	var textFiles []fs.DirEntry

	for _, file := range allFiles {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			textFiles = append(textFiles, file)
		}
	}

	if len(textFiles) == 0 {
		return errNoFiles
	}

	line := liner.NewLiner()
	defer line.Close()

	for _, file := range textFiles {
		originalFileName := file.Name()
		renameFor := true

		for renameFor {
			currentFileName := originalFileName
			for {
				//check the fileName to be the correct format
				if fileNameRe.MatchString(currentFileName) {
					renameFor = false
					break
				} else {
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
						newPath := filepath.Join(".", input)
						if _, err := os.Stat(newPath); err == nil {
							fmt.Printf("File name '%s' already exist, input a different one\n", input)
						} else if !errors.Is(err, fs.ErrNotExist) {
							log.Printf("error checking if file '%s' exist: %v\n", input, err)
						} else {
							currentFileName = input
						}
					}
				}
			}

			if originalFileName == currentFileName {
				renameFor = false
			} else {
				if err := os.Rename(filepath.Join(".", originalFileName), filepath.Join(".", currentFileName)); err != nil {
					log.Printf("error renaming file '%s' to '%s': %v\n", originalFileName, currentFileName, err)
				}
			}
		}
	}
	return nil
}

// return a slice of [file.Name()].
// Important to only call this after [checkFileNames()]
func listFiles() ([]string, error) {
	allFiles, err := os.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("error listing files: %w", err)
	}

	var textFiles []string

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

// accept the name of a file, allow to modify, return nil if all operations were
// execute correctly
func checkFormatNote(nameNote string) error {
	data, err := os.ReadFile(nameNote)
	if err != nil {
		return fmt.Errorf("error reading file '%s' : %w", nameNote, err)
	}

	content := strings.Split(string(data), "\n")
	// the .Split leave me with a final empty string element
	content = content[:len(content)-1]
	newContent := ""

	if canonRe.MatchString(content[0]) {
		newContent += content[0] + "\n"
	} else {
		for {
			fmt.Println("File:", nameNote)
			fmt.Println("Current first line:")
			fmt.Println(content[0])
			fmt.Println("Choose operation")
			fmt.Println("1- Add line above")
			fmt.Println("2- Edit line")
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
						input, err := line.PrefilledInput(content[0], -1)
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
				default:
					fmt.Printf("'%s' is an invalid option.\n", opt)
				}
			}
		}
	}

	// check each line after the first, for non-valid ones allow user to erase or modify
	for n := 1; n < len(content); n++ {
		switch {
		case canonRe.MatchString(content[n]):
			newContent += content[n] + "\n"
		case dayNoWorkRe.MatchString(content[n]):
			day := strings.Split(content[n], ":")
			newContent += day[0] + "\n"
			newContent += "M:" + day[1] + "\n"
			newContent += "T:0" + "\n"
			switch {
			case n+1 == len(content):
				newContent, _ = strings.CutSuffix(newContent, "\n")
			case canonRe.MatchString(content[n+1]):
				continue
			case dayNoWorkRe.MatchString(content[n+1]):
				continue
			case dayWorkRe.MatchString(content[n+1]):
				continue
			default:
				// error, the next line is invalid
				proceed := true
				for proceed {
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
		case dayWorkRe.MatchString(content[n]):
			newContent += content[n] + "\n"
			switch {
			case n+1 > len(content):
				fmt.Println("File:", nameNote)
				fmt.Println("Current line:")
				fmt.Println(content[n])
				fmt.Println("There are no entries for procedings, will be filled with 0")
				newContent += "M:0" + "\n"
				newContent += "T:0" + "\n"
			case procedingsRe.MatchString(content[n+1]):
				continue
			default:
				// error, the next line is invalid
				proceed := true
				for proceed {
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
		case procedingsRe.MatchString(content[n]):
			newContent += content[n] + "\n"
			switch {
			case n+1 == len(content):
				newContent, _ = strings.CutSuffix(newContent, "\n")
			case procedingsRe.MatchString(content[n+1]):
				continue
			case canonRe.MatchString(content[n+1]):
				continue
			case dayNoWorkRe.MatchString(content[n+1]):
				continue
			case dayWorkRe.MatchString(content[n+1]):
				continue
			default:
				// error, the next line is invalid
				proceed := true
				for proceed {
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
		default:
			// Non valid line
			proceed := true
			for proceed {
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
						// Advance the counter, jump over the next line
						n++
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
						fmt.Printf("%s was skipped\n", nameNote)
						return nil
					default:
						fmt.Printf("'%s' is an invalid option.\n", opt)
					}
				}
			}
		}
	}
	// Was the note modified?
	if string(data) != newContent {
		retry := true
		for retry {
			if err := os.WriteFile(nameNote, []byte(newContent), 0666); err != nil {
				fmt.Println("File:", nameNote)
				log.Printf("error saving the note: %v\n", err)
				fmt.Println("Do you want to retry? (y/n)")
				fmt.Print("> ")
				opt, err := reader.ReadString('\n')
				if err != nil {
					log.Printf("error reading input: %v\n", err)
				} else {
					opt = strings.TrimSpace(opt)
					switch opt {
					case "y", "Y":
						continue
					case "n", "N":
						fmt.Println("The modifications were not saved")
						retry = false
					default:
						fmt.Printf("'%s' is an invalid option.\n", opt)
					}
				}
			} else {
				fmt.Printf("Modifications for '%s' successfully saved\n", nameNote)
				retry = false
			}
		}
	}
	return nil
}

// Evaluate if the given line conform to any of the declared regex's
func validLine(line string) bool {
	switch {
	case canonRe.MatchString(line), dayNoWorkRe.MatchString(line),
		dayWorkRe.MatchString(line), procedingsRe.MatchString(line):
		return true
	default:
		return false
	}
}
