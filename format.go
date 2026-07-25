package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
				fmt.Printf("Formating of '%s' skipped.\n", note)
			default:
				fmt.Fprintf(os.Stderr, "error checking the format of '%s': %v\n", note, err)
			}
		case errRenameCancel:
			fmt.Printf("The renaming of '%s' was canceled.\n", note)
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

	// check each line after the first, for non-valid ones allow user to erase or modify
	for lineN < len(content) {
		n, line, err := formatLine(nameNote, content, lineN, bufio.NewReader(os.Stdin))
		if err != nil {
			return err
		}
		newContent += line
		lineN += n
	}

	newContent, _ = strings.CutSuffix(newContent, "\n")

	if err := moveFormated(nameNote, newContent); err != nil {
		return err
	}

	fmt.Println("Formated:", nameNote)
	return nil
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
func nextLineInvalid(nameNote string, currentLine string, nextLine string, reader *bufio.Reader) int {
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
	case canonRe.MatchString(line), morningRe.MatchString(line),
		afternoonRe.MatchString(line):
		return true
	case dayNoWorkRe.MatchString(line), dayWorkRe.MatchString(line):
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

// formatLine check that `lineN` and `lineN+1` of `content` are valid, return
// the number of lines to advance, the valid line to save, the only error it may
// return is `errSkipNote`
func formatLine(nameNote string, content []string, lineN int, reader *bufio.Reader) (int, string, error) {
	line := ""
	switch {
	case content[lineN] == "":
		return 1, "", nil
	case canonRe.MatchString(content[lineN]):
		line += content[lineN] + "\n"
		switch {
		case lineN+1 == len(content):
			return 1, "", nil
		case dayWorkRe.MatchString(content[lineN+1]),
			dayNoWorkRe.MatchString(content[lineN+1]):
			return 1, line, nil
		default:
			n := nextLineInvalid(nameNote, content[lineN], content[lineN+1], reader)
			return n + 1, line, nil
		}
	case dayNoWorkRe.MatchString(content[lineN]):
		day := strings.Split(content[lineN], ":")
		line += addPadding(day[0]) + "\n"
		// remove whitespace
		line += "m:" + strings.ReplaceAll(day[1], " ", "") + "\n"
		line += "t:0\n"
		switch {
		case lineN+1 == len(content),
			dayWorkRe.MatchString(content[lineN+1]),
			dayNoWorkRe.MatchString(content[lineN+1]),
			canonRe.MatchString(content[lineN+1]):
			return 1, line, nil
		default:
			n := nextLineInvalid(nameNote, content[lineN], content[lineN+1], reader)
			return n + 1, line, nil
		}
	case dayWorkRe.MatchString(content[lineN]):
		line += addPadding(content[lineN]) + "\n"
		switch {
		case lineN+1 == len(content):
			line += "m:0\n"
			line += "t:0\n"
			return 1, line, nil
		case morningRe.MatchString(content[lineN+1]):
			return 1, line, nil
		default:
			n := nextLineInvalid(nameNote, content[lineN], content[lineN+1], reader)
			return n + 1, line, nil
		}
	case morningRe.MatchString(content[lineN]):
		line += strings.ReplaceAll(content[lineN], " ", "") + "\n"
		switch {
		case lineN+1 == len(content),
			canonRe.MatchString(content[lineN+1]),
			dayNoWorkRe.MatchString(content[lineN+1]),
			dayWorkRe.MatchString(content[lineN+1]):
			line += "t:0\n"
			return 1, line, nil
		case afternoonRe.MatchString(content[lineN+1]):
			return 1, line, nil
		default:
			n := nextLineInvalid(nameNote, content[lineN], content[lineN+1], reader)
			return n + 1, line, nil
		}
	case afternoonRe.MatchString(content[lineN]):
		line += strings.ReplaceAll(content[lineN], " ", "") + "\n"
		switch {
		case lineN+1 == len(content),
			canonRe.MatchString(content[lineN+1]),
			dayNoWorkRe.MatchString(content[lineN+1]),
			dayWorkRe.MatchString(content[lineN+1]):
			return 1, line, nil
		default:
			n := nextLineInvalid(nameNote, content[lineN], content[lineN+1], reader)
			return n + 1, line, nil
		}
	default:
		line, err := invalidLine(nameNote, content[lineN], linerInput, reader)
		return 1, line, err
	}
}

// invalidLine show an invalid line and prompt the user to choose what to do,
// return a valid line or error
func invalidLine(nameNote string,
	content string,
	readInput func(string) (string, error),
	reader *bufio.Reader,
) (string, error) {
	for {
		fmt.Println()
		fmt.Println("File:", nameNote)
		fmt.Println("Current line:")
		fmt.Println(content)
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
			return "", nil
		case "2":
			for {
				fmt.Println("Modify the line and press Enter")
				input, err := readInput(content)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error on input: %v\n", err)
					continue
				}
				if validLine(input) {
					return input + "\n", nil
				}
				fmt.Printf("'%s'\n is not a valid line\n", input)
			}
		case "3":
			return "", errSkipNote
		default:
			fmt.Fprintf(os.Stderr, "'%s' is an invalid option.\n", opt)
		}
	}
}

func moveFormated(nameNote string, content string) error {
	formatedNote, err := os.Create(filepath.Join(formatedDir, nameNote))
	if err != nil {
		return fmt.Errorf("os.Create(filepath.Join(%s, %s)): %w", formatedDir, nameNote, err)
	}
	defer formatedNote.Close()

	if _, err := formatedNote.Write([]byte(content)); err != nil {
		return fmt.Errorf("formatedNote.Write([]byte(%s))): %w", content, err)
	}

	orgNote := filepath.Join(originalsDir, nameNote)
	if err := os.Remove(orgNote); err != nil {
		return fmt.Errorf("os.Remove(%s): %w", orgNote, err)
	}
	return nil
}
