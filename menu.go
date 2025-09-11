package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func menu() {
	for {
		fmt.Println("Operations (use numbers to choose e.g. 1, 2, 3)")
		fmt.Println("1- Get notes")
		fmt.Println("2- Process notes")
		fmt.Println("3- Show notes")
		fmt.Println("4- Exit")
		fmt.Print("> ")

		opt, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
		} else {
			opt = strings.TrimSpace(opt)
			switch opt {
			case "1":
			case "2":
			case "3":
			case "4":
				innerFor := true
				for innerFor {
					fmt.Println("Confirm exit? (y/n)")
					fmt.Print("> ")

					opt, err := reader.ReadString('\n')
					if err != nil {
						fmt.Printf("Error reading input: %v\n", err)
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
