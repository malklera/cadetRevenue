package internal

import (
	"cadetRevenue/internal/database"
	"errors"
	"fmt"
	"github.com/malklera/sliner/pkg/liner"
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
					fmt.Println("Where 'month' is a valid month written in spanish word")
					fmt.Println("Where 'int' is a number from 0 to 9")
					fmt.Println("Where 'year' is a number from 0000 to 9999")
					fmt.Printf("> ")

					input, err := line.PrefilledInput(currentFileName, -1)
					if err != nil {
						log.Printf("error on input: %v", err)
					} else {
						newPath := filepath.Join(".", input)
						if _, err := os.Stat(newPath); err == nil {
							fmt.Printf("File name '%s' already exist, input a different one", input)
						} else if !errors.Is(err, fs.ErrNotExist) {
							log.Printf("error checking if file '%s' exist: %v", input, err)
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
					log.Printf("error renaming file '%s' to '%s': %v", originalFileName, currentFileName, err)
				}
			}
		}
	}
	return nil
}

// Important to only call this after checkFileNames()
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

func checkFormatNote(nameNote string) error {
	data, err := os.ReadFile(nameNote)
	if err != nil {
		return fmt.Errorf("error reading file '%s' : %w", nameNote, err)
	}

	content := strings.Split(string(data), "\n")

	if !canonRe.MatchString(content[0]) {
		// TODO: here i will use promt package to modify the ones that are wrong
	}

	return nil
}

func saveNote() {

}
