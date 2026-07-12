package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestIsValidFileName(t *testing.T) {
	tests := []struct {
		file string
		want bool
	}{
		// Valid filenames
		{"2024-enero-1.txt", true},
		{"2024-febrero-2.txt", true},
		{"2024-marzo-3.txt", true},
		{"2024-abril-4.txt", true},
		{"2024-mayo-5.txt", true},
		{"2024-junio-6.txt", true},
		{"2024-julio-7.txt", true},
		{"2024-agosto-8.txt", true},
		{"2024-septiembre-9.txt", true},
		{"2024-octubre-0.txt", true},
		{"2024-noviembre-1.txt", true},
		{"2024-diciembre-9.txt", true},
		{"0000-enero-0.txt", true},

		// Invalid: wrong format
		{"", false},
		{"2024-enero-1", false},
		{"2024-enero-1.pdf", false},
		{"2024-enero-10.txt", false},
		{"2024-enero.txt", false},
		{"2024-enero-1.txt.bak", false},

		// Invalid: wrong year
		{"24-enero-1.txt", false},
		{"20240-enero-1.txt", false},
		{"-enero-1.txt", false},
		{"abcd-enero-1.txt", false},

		// Invalid: wrong month
		{"2024-january-1.txt", false},
		{"2024-yanero-1.txt", false},
		{"2024-Enero-1.txt", false}, // regex is case-sensitive
		{"2024-01-1.txt", false},
		{"2024-ene-1.txt", false},

		// Invalid: wrong day
		{"2024-enero--1.txt", false},
		{"2024-enero-123.txt", false},
		{"2024-enero-a.txt", false},

		// Invalid: wrong order
		{"enero-1-2025.txt", false},
		{"1-enero-2025.txt", false},
		{"1-2025-enero.txt", false},
		{"2025-1-enero.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			got := isValidFileName(tt.file)
			if got != tt.want {
				t.Errorf("isValidFileName(%q) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}

// validFileName does three things:
//  1. Validates the filename (loops until valid via readInput)
//  2. Checks the new name doesn't already exist (via stat)
//  3. Renames the file if the name changed (via rename)
//
// To test this function without touching the real filesystem or prompting a
// real user, we pass MOCK implementations of readInput, stat, and rename.
// A mock is a fake function that behaves exactly like the real one but
// without side effects — e.g., instead of reading from the terminal,
// it returns a pre-programmed string.
//
// We also pass a mock bufio.Reader for the retry prompt in the rename loop.
//
// The table-driven pattern here uses a struct that holds:
//   - The input filename
//   - Mock functions (readInput, stat, rename)
//   - A mock bufio.Reader for the rename retry prompt
//   - The expected output filename and error
func TestValidFileName(t *testing.T) {
	tests := []struct {
		name      string                            // descriptive name
		file      string                            // input filename
		readInput func(string) (string, error)      // mock: returns user's input
		stat      func(string) (os.FileInfo, error) // mock: checks if file exists
		rename    func(string, string) error        // mock: renames a file
		reader    *bufio.Reader                     // mock: for retry prompt
		wantFile  string                            // expected returned filename
		wantErr   error                             // expected returned error (nil means no error)
	}{
		// Case 1: Filename is already valid — no prompting, no renaming.
		// The function should return the same filename immediately.
		{
			name: "already valid, no changes needed",
			file: "2024-enero-1.txt",
			// These won't be called, but we provide non-nil stubs to avoid nil panics.
			readInput: func(s string) (string, error) { return s, nil },
			stat:      func(s string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			rename:    func(s, d string) error { return nil },
			reader:    bufio.NewReader(strings.NewReader("")),
			wantFile:  "2024-enero-1.txt",
			wantErr:   nil,
		},
		// Case 2: Invalid filename → user provides valid name → rename succeeds.
		// readInput is called once to get the corrected name.
		// rename is called to move the file to the new name.
		{
			name: "invalid then valid, rename succeeds",
			file: "foo.txt",
			readInput: func(s string) (string, error) {
				return "2024-enero-1.txt", nil
			},
			stat: func(s string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			rename: func(old, new string) error {
				return nil
			},
			reader:   bufio.NewReader(strings.NewReader("")),
			wantFile: "2024-enero-1.txt",
			wantErr:  nil,
		},
		// Case 3: Invalid → user types invalid again → then types valid → rename succeeds.
		// This tests the loop: readInput is called twice before a valid name is entered.
		{
			name: "multiple invalid inputs then valid",
			file: "bar.txt",
			readInput: func() func(string) (string, error) {
				// callCount is captured by the closure and persists across calls.
				// The outer function is called once (at struct init time) to create
				// the mock; the returned inner function is called each time
				// validFileName needs user input.
				callCount := 0
				return func(s string) (string, error) {
					callCount++
					if callCount == 1 {
						return "invalid", nil
					}
					return "2024-febrero-3.txt", nil
				}
			}(),
			stat: func(s string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			rename: func(old, new string) error {
				return nil
			},
			reader:   bufio.NewReader(strings.NewReader("")),
			wantFile: "2024-febrero-3.txt",
			wantErr:  nil,
		},
		// Case 4: Valid name the user enters already exists — rejected, prompted again.
		// stat returns nil (file exists) for the first input, then ErrNotExist for the second.
		{
			name: "name already exists, then valid",
			file: "baz.txt",
			readInput: func() func(string) (string, error) {
				// First call: return a name that already exists.
				// Second call: return a name that's available.
				callCount := 0
				return func(s string) (string, error) {
					callCount++
					if callCount == 1 {
						return "2024-enero-1.txt", nil // this one exists
					}
					return "2024-marzo-5.txt", nil // this one is free
				}
			}(),
			stat: func(s string) (os.FileInfo, error) {
				// Simulate: originalsDir + "2024-enero-1.txt" exists,
				//           originalsDir + "2024-marzo-5.txt" does not.
				if strings.Contains(s, "2024-enero-1.txt") {
					return nil, nil // nil FileInfo + nil error = file exists
				}
				return nil, os.ErrNotExist
			},
			rename: func(old, new string) error {
				return nil
			},
			reader:   bufio.NewReader(strings.NewReader("")),
			wantFile: "2024-marzo-5.txt",
			wantErr:  nil,
		},
		// Case 5: stat returns an unexpected error (not ErrNotExist).
		// The function should print the error and call readInput again.
		{
			name: "stat error, then valid",
			file: "qux.txt",
			readInput: func() func(string) (string, error) {
				callCount := 0
				return func(s string) (string, error) {
					callCount++
					if callCount == 1 {
						return "2024-abril-2.txt", nil
					}
					return "2024-junio-7.txt", nil
				}
			}(),
			stat: func() func(string) (os.FileInfo, error) {
				// First call: return an unexpected error (not ErrNotExist).
				// Second call: return ErrNotExist (name is available).
				callCount := 0
				return func(s string) (os.FileInfo, error) {
					callCount++
					if callCount == 1 {
						return nil, fmt.Errorf("permission denied")
					}
					return nil, os.ErrNotExist
				}
			}(),
			rename: func(old, new string) error {
				return nil
			},
			reader:   bufio.NewReader(strings.NewReader("")),
			wantFile: "2024-junio-7.txt",
			wantErr:  nil,
		},
		// Case 6: Rename fails, user cancels.
		// The rename mock returns an error, and the reader returns "n" to cancel.
		{
			name: "rename fails, user cancels",
			file: "old.txt",
			readInput: func(s string) (string, error) {
				return "2024-mayo-4.txt", nil
			},
			stat: func(s string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			rename: func(old, new string) error {
				return fmt.Errorf("disk full")
			},
			// The retry prompt reads "n\n", so the user declines to retry.
			reader:   bufio.NewReader(strings.NewReader("n\n")),
			wantFile: "old.txt",       // original file is returned on cancel
			wantErr:  errRenameCancel, // and this error signals the cancel
		},
		// Case 7: readInput fails — the loop should continue (not crash).
		// After the error, readInput is called again and returns a valid name.
		{
			name: "readInput error, then succeeds",
			file: "bad.txt",
			readInput: func() func(string) (string, error) {
				callCount := 0
				return func(s string) (string, error) {
					callCount++
					if callCount == 1 {
						return "", fmt.Errorf("terminal closed")
					}
					return "2024-septiembre-8.txt", nil
				}
			}(),
			stat: func(s string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			rename: func(old, new string) error {
				return nil
			},
			reader:   bufio.NewReader(strings.NewReader("")),
			wantFile: "2024-septiembre-8.txt",
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, gotErr := validFileName(tt.file, tt.readInput, tt.stat, tt.rename, tt.reader)
			if gotFile != tt.wantFile {
				t.Errorf("filename: got %q, want %q", gotFile, tt.wantFile)
			}
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("error: got %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}
