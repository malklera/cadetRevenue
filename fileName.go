package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/malklera/sliner/pkg/liner"
)

var (
	fileNameRe = regexp.MustCompile(`^(\d{4})-(enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre)-\d{1}\.txt$`)
	fileInput  = func(current string) (string, error) {
		line := liner.NewLiner()
		defer line.Close()
		return line.PrefilledInput(current, -1)
	}
	fileStat   = os.Stat
	fileRename = os.Rename
)

// TODO: write better doc

// validFileName
func validFileName(file string,
	readInput func(string) (string, error),
	stat func(string) (os.FileInfo, error),
	rename func(string, string) error,
) (string, error) {
	currentFileName := file

	for !fileNameRe.MatchString(currentFileName) {
		fmt.Println()
		fmt.Printf("'%s' is not a valid file name\n", currentFileName)
		fmt.Println("The correct format is: year-month-int.txt")
		fmt.Println("Where 'year' is a number from 0000 to 9999")
		fmt.Println("Where 'month' is a valid month written in Spanish word")
		fmt.Println("Where 'int' is a number from 0 to 9")
		fmt.Printf("> ")

		input, err := readInput(currentFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error on input: %v\n", err)
			continue
		}
		if _, err := os.Stat(filepath.Join(originalsDir, input)); err == nil {
			fmt.Printf("File name '%s' already exist, input a different one\n", input)
		} else if !errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "error checking if file '%s' exist: %v\n", input, err)
		} else {
			currentFileName = input
		}
	}

	if file == currentFileName {
		return file, nil
	}

	for {
		if err := rename(filepath.Join(originalsDir, file), filepath.Join(originalsDir, currentFileName)); err != nil {
			fmt.Fprintf(os.Stderr, "error renaming file '%s' to '%s': %v\n", file, currentFileName, err)
			fmt.Println("Do you want to retry? (y/n)")
			fmt.Print("> ")
			opt, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
				continue
			}

			opt = strings.TrimSpace(opt)
			switch opt {
			case "y", "Y":
				continue
			case "n", "N":
				return file, errRenameCancel
			default:
				fmt.Fprintf(os.Stderr, "'%s' is an invalid option.\n", opt)
				continue
			}
		}
		break
	}

	fmt.Println()
	fmt.Printf("File '%s' succesfull renamed to '%s'\n", file, currentFileName)
	return currentFileName, nil
}
