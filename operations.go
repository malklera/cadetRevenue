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
	fileNameRe = regexp.MustCompile(`^(Enero|Febrero|Marzo|Abril|Mayo|Junio|Julio|Agosto|Septiembre|Octubre|Noviembre|Diciembre) \d{4}.txt$`)
	canonRe = regexp.MustCompile(`^Canon \d+$`)
    dayNoWorkRe = regexp.MustCompile(`^(Lunes|Martes|Miércoles|Miercoles|Jueves|Viernes|Sábado|Sabado) \d{1,2}\/\d{1,2}: *(0|-\d+)$`)
    dayWorkRe = regexp.MustCompile(`^(Lunes|Martes|Miércoles|Miercoles|Jueves|Viernes|Sábado|Sabado) \d{1,2}\/\d{1,2}$`)
    procedingsRe = regexp.MustCompile(`^(M|T): *(?:-\d+|\d+(?:\+\d+)*(?:-\d+)?)$`)
)

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
