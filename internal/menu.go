// Package internal contains all functions used by the application
package internal

import (
	"bufio"
	"cadetRevenue/internal/database"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func Menu() {
	for {
		fmt.Println()
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
			case "2":
				listNotes, err := listFiles(formatedDir)
				switch err {
				case nil:
					for n := range listNotes {
						entry, err := processNote(listNotes[n])
						if err != nil {
							log.Printf("error processing note '%s' : %v", listNotes[n], err)
						} else {
							dbInstance, err := database.New()
							moveFile := false
							if err != nil {
								log.Printf("error opening the database: %v", err)
							} else {
								defer dbInstance.Close()
								for _, e := range entry {
									if err := database.AddEntry(dbInstance, e); err != nil {
										log.Printf("error adding entry '%v' to the database: %v", e.Date, err)
									} else {
										moveFile = true
									}
								}
							}
							// move the file from formated to processed
							if moveFile {
								if err := os.Rename(filepath.Join(formatedDir, listNotes[n]), filepath.Join(processedDir, listNotes[n])); err != nil {
									log.Printf("error moving formated note to the processed directory: %v", err)
								}
							}
						}
					}
				case errNoFiles:
					fmt.Println("There are no files to process")
				default:
					log.Printf("error listing files: %v\n", err)
				}
			case "3":
				dbInstance, err := database.New()
				defer dbInstance.Close()
				if err != nil {
					log.Printf("error opening the database: %v", err)
				} else {
					entries, err := database.ShowAll(dbInstance)
					if err != nil {
						log.Printf("error on ShowAll: %v", err)
					} else {
						for _, entry := range entries {
							fmt.Println(entry)
						}
					}
				}
			case "4":
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
