package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	fileNameRe   = regexp.MustCompile(`^(enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre)-\d{1}-\d{4}.txt$`)
	canonRe      = regexp.MustCompile(`^Canon \d+$`)
	dayNoWorkRe  = regexp.MustCompile(`^(Lunes|Martes|Miércoles|Miercoles|Jueves|Viernes|Sábado|Sabado) \d{1,2}\/\d{1,2}: *(0|-\d+)$`)
	dayWorkRe    = regexp.MustCompile(`^(Lunes|Martes|Miércoles|Miercoles|Jueves|Viernes|Sábado|Sabado) \d{1,2}\/\d{1,2}$`)
	procedingsRe = regexp.MustCompile(`^(M|T): *(?:-\d+|\d+(?:\+\d+)*(?:-\d+)?)$`)
)

type Entry struct {
	ID       int64
	Year     int
	Month    int
	Day      int
	Canon    int
	Income   int
	Expenses int
}

func processNotes() {
	files, err := os.ReadDir(".")
	if err != nil {
		log.Printf("Error listing files: %v", err)
	}

	var textFiles []os.DirEntry

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".txt") {
			textFiles = append(textFiles, file)
		}
	}

	for _, file := range textFiles {
		//check the fileName to be the correct format
		if validFileName := fileNameRe.MatchString(file.Name()); validFileName {
			// get the year from here, the month and number do not matter
		} else {
			fmt.Printf("'%s' is not a valid file name\n", file.Name())
		}
		if err := checkFormat(file.Name()); err != nil {
			log.Printf("Error on checkFormat(%s) : %v", file.Name(), err)
		}
	}
	// loop trough the list and call checkFormat
	// once the correct format, call saveNote
	// once the note has been saved, erase the file from the directory
	// go to the next file on the list

}

func checkFormat(nameFile string) error {
	data, err := os.ReadFile(nameFile)
	if err != nil {
		return err
	}

	content := strings.Split(string(data), "\n")

	return nil
}

func saveNote() {

}
