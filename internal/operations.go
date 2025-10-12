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
	fileNameRe     = regexp.MustCompile(`^(enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre)-\d{1}-(\d{4})\.txt$`)
	canonRe        = regexp.MustCompile(`^canon \d+$`)
	dayNoWorkRe    = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) \d{1,2}\/\d{1,2}: *(0|-\d+)$`)
	dayWorkRe      = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) \d{1,2}\/\d{1,2}$`)
	dayWorkCanonRe = regexp.MustCompile(`^(lunes|martes|miércoles|miercoles|jueves|viernes|sábado|sabado) (\d{1,2}\/\d{1,2}) (canon \d+)$`)

	procedingsRe = regexp.MustCompile(`^(m|t): *(?:-\d+|\d+(?:\+\d+)*(?:-\d+)?)$`)
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

// Take the name of a file, check that it is the correct format, if not ask the
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

// return a slice of [file.Name()].
func listFiles(dir string) ([]string, error) {
	allFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error listing files: %w", err)
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

// accept the name of a file, allow to modify or cancel the modification, return
// nil if all operations were execute correctly
func checkFormatNote(nameNote string) error {
	orgNote := filepath.Join(originalsDir, nameNote)
	data, err := os.ReadFile(orgNote)
	if err != nil {
		return fmt.Errorf("error reading file '%s' : %w", nameNote, err)
	}
	fmt.Println()
	fmt.Println("Procesing:", nameNote)

	// Fix potential panic when checking empty files
	if string(data) == "" {
		fmt.Printf("'%s' is empty\n", nameNote)
		return errSkipNote
	}
	content := strings.Split(strings.ToLower(string(data)), "\n")

	// the .Split leave me with a final empty string element
	content = content[:len(content)-1]
	newContent := ""
	n := 0

	for {
		if canonRe.MatchString(content[n]) {
			newContent += content[n] + "\n"
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
			newContent += day[0] + "\n"
			newContent += "M:" + day[1] + "\n"
			newContent += "T:0" + "\n"
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
			newContent += content[n] + "\n"
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
			newContent += subStrings[3] + "\n"
			newContent += subStrings[1] + subStrings[2] + "\n"
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

		case procedingsRe.MatchString(content[n]):
			newContent += content[n] + "\n"
			switch {
			case n+1 == len(content):
				newContent, _ = strings.CutSuffix(newContent, "\n")
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
		tempName := filepath.Join(formatedDir, tempFile.Name())
		formatNote := filepath.Join(formatedDir, nameNote)

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
							if err := os.Rename(tempName, formatNote); err != nil {
								fmt.Println()
								fmt.Println("File:", tempName)
								log.Printf("error renaming '%s' to '%s' : %v\n", tempName, formatNote, err)
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

// Accept the name of a file, extract the data from it, return a [Entry] struct
// func extractData(file string) (database.Entry, error) {
//
// }
