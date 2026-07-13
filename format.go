package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/malklera/sliner/pkg/liner"
)

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

// nextLineInvalid prompt choosing if erasing the line below(return 1) or
// leaving it for later modification(return 0)
func nextLineInvalid(nameNote string, currentLine string, nextLine string) int {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println()
		fmt.Println("File:", nameNote)
		fmt.Println("Current line:")
		fmt.Println(currentLine)
		fmt.Println("The line below is invalid")
		fmt.Println(nextLine)
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
