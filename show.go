package main

// import (
//
// )
//
// func showEntries(target string, year int, month int, day int) error {
// 	dbInstance, err := db.New(target)
// 	if err != nil {
// 		return fmt.Errorf("db.New(%s): %v", target, err)
// 	}
// 	// TODO: has to change this
// 	entries, err := db.ShowAll(dbInstance)
// 	if err != nil {
// 		return fmt.Errorf("db.ShowAll(dbInstance): %v", err)
// 	}
// 	displayEntries(entries)
// 	return nil
// }
//
// // displayEntries pretty print to stdout the given entry
// func displayEntries(entries []db.Entry) {
// 	fmt.Println("Date	Canon	In. Morning	In. Afternoon	Expenses")
// 	for _, e := range entries {
// 		fmt.Printf("%v	%d	%d	%d	%d\n", e.Date, e.Canon, e.IncomeM, e.IncomeT, e.Expenses)
// 	}
// }
