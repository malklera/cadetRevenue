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
		fmt.Println("1- Format notes")
		fmt.Println("2- Process notes")
		fmt.Println("3- Show notes")
		fmt.Println("4- Exit")
		fmt.Print("> ")

		opt, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading input: %v\n", err)
		} else {
			opt = strings.TrimSpace(opt)
			switch opt {
			case "1":
				listNotes, err := listFiles(originalsDir)
				switch err {
				case nil:
					for n := range listNotes {
						note, err := checkFileName(listNotes[n])
						switch err {
						case nil:
							err := checkFormatNote(note)
							switch err {
							case nil:
								break
							case errSkipNote:
								fmt.Printf("Formating of '%s' skipped\n", note)
							default:
								log.Printf("error checking the format of '%s' : %v\n", note, err)
							}
						case errRenameCancel:
							fmt.Printf("The renaming of '%s' was canceled\n", note)
						default:
							log.Printf("error checking the name of '%s' : %v\n", note, err)
						}
					}
				case errNoFiles:
					fmt.Println("There are no files to format")
				default:
					log.Printf("error listing files: %v\n", err)
				}
				fmt.Println("add the results to a db")
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
						log.Printf("error reading input: %v\n", err)
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
