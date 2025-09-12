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
	canon_re = regexp.MustCompile(`[a-zA-ZáéíóúÁÉÍÓÚñÑ]+ \d+`)
    change_canon_re = regexp.MustCompile(`[a-zA-ZáéíóúÁÉÍÓÚñÑ]+ \d+\/\d+ [a-zA-ZáéíóúÁÉÍÓÚñÑ]+ \d+`)
    day1_re = regexp.MustCompile(`[a-zA-ZáéíóúÁÉÍÓÚñÑ]+ \d+\/\d+`)
    day2_re = regexp.MustCompile(`[a-zA-ZáéíóúÁÉÍÓÚñÑ]+ \d+\/\d+:-?\d+`)
    morning_re = regexp.MustCompile(`M:((\d+(\+\d+)*(-\d+)?)|(-\d+))`)
    afternoon_re = regexp.MustCompile(`T:-?\d+([+-]\d+)*`)
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
