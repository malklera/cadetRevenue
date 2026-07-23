package main

import (
	"testing"

	"cadetRevenue/internal/database"
	uuid "github.com/gofrs/uuid/v5"
)

func TestProcessMovement(t *testing.T) {
	entryU7 := uuid.Must(uuid.FromString("019f8237-87e8-742b-9972-c08d896cb97c"))

	var tests = []struct {
		name       string
		entryID    uuid.UUID
		content    string
		wantCount  int
		wantAmount []int64
		wantShift  string
		wantErr    bool
	}{
		{
			name:       "single value",
			entryID:    entryU7,
			content:    "m:2000",
			wantCount:  1,
			wantAmount: []int64{2000},
			wantShift:  "m",
		},
		{
			name:       "plus signs",
			entryID:    entryU7,
			content:    "t:2000+3000+2500",
			wantCount:  3,
			wantAmount: []int64{2000, 3000, 2500},
			wantShift:  "t",
		},
		{
			name:       "expense only",
			entryID:    entryU7,
			content:    "m:-4500",
			wantCount:  1,
			wantAmount: []int64{-4500},
			wantShift:  "m",
		},
		{
			name:       "income and expense",
			entryID:    entryU7,
			content:    "t:2000+3000-1500",
			wantCount:  3,
			wantAmount: []int64{-1500, 2000, 3000},
			wantShift:  "t",
		},
		{
			name:    "trailing plus errors",
			entryID: entryU7,
			content: "t:2000+",
			wantErr: true,
		},
		{
			name:    "double plus errors",
			entryID: entryU7,
			content: "m:2000++2500",
			wantErr: true,
		},
		{
			name:    "expenses not last item",
			entryID: entryU7,
			content: "t:-200+300",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processMovement(tt.entryID, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.wantAmount) {
				t.Fatalf("got %d movements, want %d", len(got), len(tt.wantAmount))
			}
			for i, mov := range got {
				if mov.Amount != tt.wantAmount[i] {
					t.Errorf("movement[%d].Amount = %d, want %d", i, mov.Amount, tt.wantAmount[i])
				}
				if mov.EntryID != tt.entryID {
					t.Errorf("movement[%d].EntryID = %v, want %v", i, mov.EntryID, tt.entryID)
				}
				if mov.Shift != tt.wantShift {
					t.Errorf("movement[%d].Shift = %q, want %q", i, mov.Shift, tt.wantShift)
				}
				if mov.ID == uuid.Nil {
					t.Errorf("movement[%d].ID is nil UUID", i)
				}
			}
		})
	}
}

func TestCalcProfit(t *testing.T) {
	var tests = []struct {
		name       string
		canon      int64
		morning    []database.Movement
		afternoon  []database.Movement
		wantResult float64
	}{
		{
			name:       "canon 8500, no work",
			canon:      8500,
			morning:    []database.Movement{{Amount: 0}},
			afternoon:  []database.Movement{{Amount: 0}},
			wantResult: 0,
		},
		{
			name:       "empty slices",
			canon:      8500,
			morning:    []database.Movement{},
			afternoon:  []database.Movement{},
			wantResult: 0,
		},
		{
			name:       "nil slices",
			canon:      8500,
			morning:    nil,
			afternoon:  nil,
			wantResult: 0,
		},
		{
			name:       "only expenses, no income",
			canon:      8500,
			morning:    []database.Movement{{Amount: -3000}},
			afternoon:  nil,
			wantResult: -3000,
		},
		{
			name:       "income below threshold, no expenses",
			canon:      8500,
			morning:    []database.Movement{{Amount: 2000}},
			afternoon:  nil,
			wantResult: 1500,
		},
		{
			name:       "income below threshold, with expenses",
			canon:      8500,
			morning:    []database.Movement{{Amount: 2000}, {Amount: -500}},
			afternoon:  nil,
			wantResult: 1000,
		},
		{
			name:       "income exactly at threshold canon*4",
			canon:      8500,
			morning:    []database.Movement{{Amount: 34000}},
			afternoon:  nil,
			wantResult: 25500,
		},
		{
			name:       "income above threshold",
			canon:      8500,
			morning:    []database.Movement{{Amount: 40000}},
			afternoon:  nil,
			wantResult: 31500,
		},
		{
			name:       "income above threshold with expenses",
			canon:      8500,
			morning:    []database.Movement{{Amount: 20000}},
			afternoon:  []database.Movement{{Amount: 20000}, {Amount: -1000}},
			wantResult: 30500,
		},
		{
			name:       "small income integer division",
			canon:      8500,
			morning:    []database.Movement{{Amount: 3}},
			afternoon:  nil,
			wantResult: 2.25,
		},
		{
			name:       "small income above integer division floor",
			canon:      8500,
			morning:    []database.Movement{{Amount: 5}},
			afternoon:  nil,
			wantResult: 3.75,
		},
		{
			name:       "income only in afternoon",
			canon:      8500,
			morning:    nil,
			afternoon:  []database.Movement{{Amount: 5000}},
			wantResult: 3750,
		},
		{
			name:       "mixed shifts below threshold",
			canon:      10000,
			morning:    []database.Movement{{Amount: 3000}},
			afternoon:  []database.Movement{{Amount: 2000}, {Amount: -200}},
			wantResult: 3550,
		},
		{
			name:       "income with multiple expenses below threshold",
			canon:      8500,
			morning:    []database.Movement{{Amount: 10000}, {Amount: -1000}, {Amount: -500}},
			afternoon:  []database.Movement{{Amount: -300}},
			wantResult: 5700,
		},
		{
			name:       "canon 0, income below threshold",
			canon:      0,
			morning:    []database.Movement{{Amount: 1000}},
			afternoon:  nil,
			wantResult: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcProfit(tt.canon, tt.morning, tt.afternoon)
			if got != tt.wantResult {
				t.Errorf("got: %f, want: %f", got, tt.wantResult)
			}
		})
	}
}
