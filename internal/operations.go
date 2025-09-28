package internal

import (
	"bufio"
	"cadetRevenue/internal/database"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	fileNameRe   = regexp.MustCompile(`^(enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre)-\d{1}-(\d{4})\.txt$`)
	canonRe      = regexp.MustCompile(`^Canon \d+$`)
	dayNoWorkRe  = regexp.MustCompile(`^(Lunes|Martes|Miércoles|Miercoles|Jueves|Viernes|Sábado|Sabado) \d{1,2}\/\d{1,2}: *(0|-\d+)$`)
	dayWorkRe    = regexp.MustCompile(`^(Lunes|Martes|Miércoles|Miercoles|Jueves|Viernes|Sábado|Sabado) \d{1,2}\/\d{1,2}$`)
	procedingsRe = regexp.MustCompile(`^(M|T): *(?:-\d+|\d+(?:\+\d+)*(?:-\d+)?)$`)
)

func ProcessNotes() error {
	allFiles, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("Error listing files: %w", err)
	}

	var textFiles []fs.DirEntry

	for _, file := range allFiles {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			textFiles = append(textFiles, file)
		}
	}

	if len(textFiles) == 0 {
		fmt.Println("No files to process")
		return nil
	}

	for _, file := range textFiles {
		originalFileName := file.Name()
		currentFileName := originalFileName
		for {
			//check the fileName to be the correct format
			if fileNameRe.MatchString(currentFileName) {
				break
			} else {
				fmt.Printf("'%s' is not a valid file name\n", currentFileName)
				fmt.Println("The correct format is: month-int-year.txt")
				fmt.Println("Where 'month' is a valid month written in spanish word")
				fmt.Println("Where 'int' is a number from 0 to 9")
				fmt.Println("Where 'year' is a number from 0000 to 9999")
				fmt.Printf("> ")

				input := ""
				if scanner.Scan() {
					input = scanner.Text()
				}
				input = strings.TrimSpace(input)
				newPath := filepath.Join(".", input)
				if _, err := os.Stat(newPath); err == nil {
					fmt.Printf("File name '%s' already exist, input a different one", input)
				} else if !errors.Is(err, fs.ErrNotExist) {
					fmt.Printf("Error checking if file '%s' exist: %v", input, err)
				} else {
					currentFileName = input
				}
			}
		}

		if originalFileName != currentFileName {
			if err := os.Rename(filepath.Join(".", originalFileName), filepath.Join(".", currentFileName)); err != nil {
				log.Printf("Error renaming file '%s' to '%s': %v", originalFileName, currentFileName, err)
			}
		}

		// NOTE: do i want to get the year here, or do i do it on saveNote?
		var noteEntry database.Entry
		// NOTE: i am not validating the year
		matches := fileNameRe.FindStringSubmatch(currentFileName)
		year, err := strconv.Atoi(matches[2])
		if err != nil {
			log.Printf("Error extracting the year from file '%s': %v", currentFileName, err)
		} else {
			noteEntry.Year = year
		}

		if err := checkFormat(currentFileName); err != nil {
			log.Printf("Error on checkFormat(%s) : %v", currentFileName, err)
		}

	}
	// loop trough the list and call checkFormat
	// once the correct format, call saveNote
	// once the note has been saved, erase the file from the directory
	// go to the next file on the list

	return nil
}

func checkFormat(nameFile string) error {
	data, err := os.ReadFile(nameFile)
	if err != nil {
		return err
	}

	content := strings.Split(string(data), "\n")

	if !canonRe.MatchString(content[0]) {
		// TODO: here i will use promt package to modify the ones that are wrong
	}
	// for line := range content {
	//
	// }

	return nil
}

func saveNote() {

}
