// Package internal contains all functions used by the application
package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func Menu() {
	for {
		fmt.Println("Operations (use numbers to choose e.g. 1, 2, 3)")
		fmt.Println("1- Process notes")
		fmt.Println("2- Show notes")
		fmt.Println("3- Exit")
		fmt.Print("> ")

		opt, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading input: %v", err)
		} else {
			opt = strings.TrimSpace(opt)
			switch opt {
			case "1":
				if err := checkFileNames(); err != nil {
					if err == errNoFiles {
						fmt.Println("There are no notes to process")
					} else {
						log.Printf("error checking the format of the file names: %v\n", err)
					}
				} else {
					listNotes, err := listFiles()
					if err != nil {
						log.Printf("error listing files: %v", err)
					} else {
						for _, note := range listNotes {
							if err := checkFormatNote(note); err != nil {
								log.Printf("error checking the format of '%s' : %v", note, err)
							}
						}
					}
				}
				fmt.Println("process the note, add the results to a db")
				fmt.Println("move the note with the correct format to 'originals' directory")
			case "2":
				fmt.Println("call showMenu()")
			case "3":
				innerFor := true
				for innerFor {
					fmt.Println("Confirm exit? (y/n)")
					fmt.Print("> ")
					opt, err := reader.ReadString('\n')
					if err != nil {
						log.Printf("error reading input: %v", err)
					} else {
						opt = strings.TrimSpace(opt)
						switch opt {
						case "y", "Y":
							os.Exit(0)
						case "n", "N":
							innerFor = false
						default:
							fmt.Printf("'%s' is an invalid option.\n", opt)
						}
					}
				}
			default:
				fmt.Printf("'%s' is an invalid option.\n", opt)
			}
		}
	}
}
