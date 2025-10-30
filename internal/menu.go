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
				showOptions()
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

// showOptions allows you to choose between different queries to run
func showOptions() {
	// options are all months of a year on chronological order, showing the net profit
	// of each month
	// or choose a year and month and show that one
	// or show all entries(leave the ShowAll() for now
	for {
		fmt.Println()
		fmt.Println("What to show")
		fmt.Println("1- All months")
		fmt.Println("2- A specific month")
		fmt.Println("3- All entries")
		fmt.Println("4- Exit")
		fmt.Print("> ")

		opt, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading input: %v\n", err)
		} else {
			opt = strings.TrimSpace(opt)
			switch opt {
			case "1":
				// TODO: show the same format as in specific month, but do not ask
				// for user input about which ones, just show all available
			case "2":
				dbInstance, err := database.New()
				if err != nil {
					log.Printf("error opening the database: %v", err)
				} else {
					defer dbInstance.Close()
					years, err := database.GetYears(dbInstance)
					if err != nil {
						log.Printf("error geting the available years: %v\n", err)
					} else {
						for {
							fmt.Println()
							fmt.Println("The available years are:")
							for _, y := range years {
								fmt.Println(y)
							}
							fmt.Println()
							fmt.Println("Choose a year (input the year as displayed)")
							fmt.Println("Input 'q' to exit")
							fmt.Print("> ")
							optY, err := reader.ReadString('\n')
							if err != nil {
								log.Printf("error reading input: %v\n", err)
							} else {
								optY = strings.TrimSpace(optY)
								if optY == "q" {
									break
								} else {
									valid := false
									for _, y := range years {
										if y == optY {
											valid = true
										}
									}
									if valid {
										months, err := database.GetMonths(dbInstance, optY)
										if err != nil {
											log.Printf("error geting the available months for year '%s': %v\n", optY, err)
										} else {
											for {
												fmt.Println()
												fmt.Println("The available months are:")
												for _, m := range months {
													fmt.Println(m)
												}
												fmt.Println()
												fmt.Println("Choose a month (input the month as displayed)")
												fmt.Println("Input 'q' to exit")
												fmt.Print("> ")
												optM, err := reader.ReadString('\n')
												if err != nil {
													log.Printf("error reading input: %v\n", err)
												} else {
													optM = strings.TrimSpace(optM)
													if optM == "q" {
														break
													} else {
														valid := false
														for _, m := range months {
															if m == optM {
																valid = true
															}
														}
														if valid {
															entries, err := database.GetEntries(dbInstance, optY, optM)
															if err != nil {
																log.Printf("error getting the entries for year '%s' month '%s': %v\n", optY, optM, err)
															} else {
																// TODO: check if the month has entries for all the working days
																fmt.Println("Date	- Net profit")
																fmt.Printf("%s-%s	- $ %d\n", optY, optM, netRevenue(entries))
															}
														} else {
															fmt.Printf("'%s' is an invalid option\n", optM)
														}
													}
												}
											}
										}
									} else {
										fmt.Printf("'%s' is an invalid option\n", optY)
									}
								}
							}
						}
					}
				}
			case "3":
				dbInstance, err := database.New()
				if err != nil {
					log.Printf("error opening the database: %v", err)
				} else {
					defer dbInstance.Close()
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
				return
			default:
				fmt.Printf("'%s' is an invalid option.\n", opt)
			}
		}
	}
}
