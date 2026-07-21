package main

import (
	"testing"

	"cadetRevenue/internal/database"
	uuid "github.com/gofrs/uuid/v5"
)

// TODO: fix this
// func TestProcessProcedings(t *testing.T) {
// 	var contents = []struct {
// 		content  string
// 		income   int
// 		expenses int
// 		err      error
// 	}{
// 		{"M:2000", 2000, 0, nil},
// 		{"T:2000+2000", 4000, 0, nil},
// 		{"m:-4500", 0, -4500, nil},
// 		{"M:2500+2200+2500+6000+2000+4000+4000+2000-14800", 25200, -14800, nil},
// 		{"t:2000+5000+2000+2000+2000+3000+2500-3300", 18500, -3300, nil},
// 		{"M:2000++2500+4500+4500+4000-2000", 0, 0, strconv.ErrSyntax},
// 		{"T:-2000+3000", 0, 0, strconv.ErrSyntax}, // TODO: this will change when i update the regex
// 		// {"T:-2000-3000", 0, 0, strconv.ErrSyntax}, // TODO: this will change when i update the regex
// 		{"T:2000+", 0, 0, strconv.ErrSyntax}, // TODO: this will change when i update the regex
// 	}
//
// 	for _, tt := range contents {
// 		t.Run(tt.content, func(t *testing.T) {
// 			in, ex, er := processProcedings(tt.content)
// 			if in != tt.income || ex != tt.expenses || !errors.Is(er, tt.err) {
// 				t.Errorf("got '%d, %d, %v'; want '%d, %d, %v'", in, ex, er, tt.income, tt.expenses, tt.err)
// 			}
// 		})
// 	}
// }

func TestProcessMovement(t *testing.T) {
	entryU7 := uuid.Must(uuid.FromString("019f8237-87e8-742b-9972-c08d896cb97c"))
	movU7 := uuid.Must(uuid.FromString("019f8244-fd05-7cfe-b31e-21a6f73eeefb"))
	var tests = []struct {
		name    string
		entryID uuid.UUID
		content string
		wantMov []database.Movement
		wantErr error
	}{
		{"no sign", entryU7, "m:2000", []database.Movement{database.Movement{ID: movU7, EntryID: entryU7, Shift: "m", Amount: int64(2000)}}, nil},
		{"one +", entryU7, "t:2000+2000", []database.Movement{database.Movement{ID: movU7, EntryID: entryU7, Shift: "m", Amount: int64(2000)}, database.Movement{ID: movU7, EntryID: entryU7, Shift: "m", Amount: int64(2000)}}, nil},
		{"m:-4500", 0, -4500, nil},
		{"M:2500+2200+2500+6000+2000+4000+4000+2000-14800", 25200, -14800, nil},
		{"t:2000+5000+2000+2000+2000+3000+2500-3300", 18500, -3300, nil},
		{"M:2000++2500+4500+4500+4000-2000", 0, 0, strconv.ErrSyntax},
		{"T:-2000+3000", 0, 0, strconv.ErrSyntax}, // TODO: this will change when i update the regex
		// {"T:-2000-3000", 0, 0, strconv.ErrSyntax}, // TODO: this will change when i update the regex
		{"T:2000+", 0, 0, strconv.ErrSyntax}, // TODO: this will change when i update the regex

	}
}
