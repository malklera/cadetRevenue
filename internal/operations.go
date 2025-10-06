package internal

import (
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

// return a slice of file.Name().
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

// accept the name of a file, allow to modify, return nil if all operations were
// execute correctly
func checkFormatNote(nameNote string) error {
	data, err := os.ReadFile(nameNote)
	if err != nil {
		return fmt.Errorf("error reading file '%s' : %w", nameNote, err)
	}

	content := strings.Split(string(data), "\n")
	var newContent []string

	if !canonRe.MatchString(content[0]) {
		// TODO: here i will use liner package to modify the ones that are wrong
		// how do i add a line?
		// either i have something wrong, or i have nothing and the first line is "valid"
		for {
			fmt.Println("Current first line:")
			fmt.Println(content[0])
			fmt.Println("Choose operation")
			fmt.Println("1- Add line above")
			fmt.Println("2- Edit line")
			fmt.Print("> ")

			opt := ""
			if scanner.Scan() {
				opt = scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				log.Printf("error reading input: %v", err)
			} else {
				switch opt {
				case "1":
					// NOTE: How is the user suppose to know what the canon should be?
					for {
						fmt.Println("New line:")
						fmt.Print("> ")
						line := ""
						if scanner.Scan() {
							line = scanner.Text()
						}
						if err := scanner.Err(); err != nil {
							log.Printf("error reading input: %v", err)
						} else {
							if canonRe.MatchString(line) {
								newContent = append(newContent, line)
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
							log.Printf("error on input: %v", err)
						} else {
							if canonRe.MatchString(input) {
								newContent = append(newContent, input)
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
	return nil
}

func saveNote() {

}
