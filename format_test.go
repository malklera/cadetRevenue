package main

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestValidFirstLine(t *testing.T) {
	var tests = []struct {
		name           string
		content        []string
		readInput      func(string) (string, error)
		reader         *bufio.Reader
		wantNewContent string
		wantN          int
	}{
		// case 1: valid
		{
			name:           "valid",
			content:        []string{"canon 5000", "Lunes 01/01"},
			readInput:      func(s string) (string, error) { return s, nil },
			reader:         bufio.NewReader(strings.NewReader("")),
			wantNewContent: "canon 5000\n",
			wantN:          1,
		},
		// case 2: invalid->add->valid
		{
			name:           "invalid->add->valid",
			content:        []string{"Lunes 02/02", "m:0"},
			readInput:      func(s string) (string, error) { return "", nil },
			reader:         bufio.NewReader(strings.NewReader("1\ncanon 5000\n")),
			wantNewContent: "canon 5000\n",
			// TODO: is this ok? or should it be 0
			wantN: 1,
		},
		// case 3: invalid->edit->valid
		{
			name:    "invalid->edit->valid",
			content: []string{"invalid", "Lunes 03/03", "m:100"},
			readInput: func(s string) (string, error) {
				return "canon 5000", nil
			},
			reader:         bufio.NewReader(strings.NewReader("2\n")),
			wantNewContent: "canon 5000\n",
			wantN:          1,
		},
		// case 4: invalid->erase->valid
		{
			name:           "invalid->erase->valid",
			content:        []string{"invalid", "canon 5000", "Lunes 04/04"},
			readInput:      func(s string) (string, error) { return "", nil },
			reader:         bufio.NewReader(strings.NewReader("3\n")),
			wantNewContent: "canon 5000\n",
			wantN:          2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNewContent, gotN := validFirstLine("fileName", tt.content, tt.readInput, tt.reader)
			if gotNewContent != tt.wantNewContent {
				t.Errorf("got: %s, want: %s", gotNewContent, tt.wantNewContent)
			}
			if gotN != tt.wantN {
				t.Errorf("got: %d, want: %d", gotN, tt.wantN)
			}
		})
	}
}

func TestAddPadding(t *testing.T) {
	var days = []struct {
		day  string
		want string
	}{
		{"Lunes 29/9", "Lunes 29/09"},
		{"Martes 3/12", "Martes 03/12"},
		{"Miércoles 1/1", "Miércoles 01/01"},
		{"Jueves 02/05", "Jueves 02/05"},
		{"Viernes 31/11", "Viernes 31/11"},
		{"Sábado 14/10", "Sábado 14/10"},
	}

	for _, tt := range days {
		t.Run(tt.day, func(t *testing.T) {
			ans := addPadding(tt.day)
			if ans != tt.want {
				t.Errorf("got '%s', want '%s'", ans, tt.want)
			}
		})
	}
}

func TestValidLine(t *testing.T) {
	var lines = []struct {
		line string
		want bool
	}{
		{"canon 7000", true},
		{"lunes 29/9:-4000", true},
		{"martes 30/9: 0", true},
		{"miércoles 1/10", true},
		{"m:2000", true},
		{"t:2000+2000", true},
		{"t:-2000+2000", false},
		{"m:-4500", true},
		{"viernes 3/10 canon 7500", false},
		{"m:2500+2200+2500+6000+2000+4000+4000+2000-14800", true},
		{"t: 2000+5000+2000+2000+2000+3000+2500-3300", true},
		{"canon", false},
		{"lunes 29/9:", false},
		{"domingo 29/9:-4000", false},
		{"jueves 40/10", false},
		{"viernes 3/13", false},
		{"viernes 0/13", false},
		{"viernes -3/13", false},
		{"viernes 3/-13", false},
		{"viernes 3/0", false},
		{"m:2000++2500+4500+4500+4000-2000", false},
	}

	for _, tt := range lines {
		t.Run(tt.line, func(t *testing.T) {
			ans := validLine(tt.line)
			if ans != tt.want {
				t.Errorf("got '%t', want '%t'", ans, tt.want)
			}
		})
	}
}

func TestValidDate(t *testing.T) {
	var dates = []struct {
		date string
		want bool
	}{
		{"13/3", true},
		{"1/12", true},
		{"31/5", true},
		{"-3/7", false},
		{"0/6", false},
		{"9/0", false},
		{"8/13", false},
	}

	for _, tt := range dates {
		t.Run(tt.date, func(t *testing.T) {
			ans := validDate(tt.date)
			if ans != tt.want {
				t.Errorf("got '%t', want '%t'", ans, tt.want)
			}
		})
	}
}

// func TestFormatNote(t *testing.T){
//
// }

func TestFormatLine(t *testing.T) {
	var tests = []struct {
		name     string
		nameNote string
		content  []string
		lineN    int
		wantN    int
		wantLine string
		wantErr  error
	}{
		{"emptyLine", "nameNote", []string{"", "canon 300"}, 0, 1, "", nil},
		{"canonRe->dayWorkRe", "nameNote", []string{"canon 350", "lunes 01/01", "m:300", "t:500", "canon 400", "martes 02/01"}, 4, 5, "canon 400", nil},
		// {"canonRe->invalidDelete", "nameNote", []string{"canon 500", "invalid"}, 0, 2, "", nil},
		// {"", "nameNote", []string{}, 0, 0, "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, line, err := formatLine(tt.nameNote, tt.content, tt.lineN)
			if n != tt.wantN {
				t.Errorf("got '%d', want '%d'", n, tt.wantN)
			}
			if line != tt.wantLine {
				t.Errorf("got '%s', want '%s'", line, tt.wantLine)
			}

			if err != tt.wantErr {
				t.Errorf("got '%v', want '%v'", err, tt.wantErr)
			}
		})
	}
}

func TestInvalidLine(t *testing.T) {
	var tests = []struct {
		name      string
		nameNote  string
		content   string
		readInput func(string) (string, error)
		reader    *bufio.Reader
		wantLine  string
		wantErr   error
	}{
		{
			name:      "invalid->1",
			nameNote:  "nameNote",
			content:   "invalid",
			readInput: func(s string) (string, error) { return s, nil },
			reader:    bufio.NewReader(strings.NewReader("1\n")),
			wantLine:  "",
			wantErr:   nil,
		},
		{
			name:     "invalid->2->valid",
			nameNote: "nameNote",
			content:  "m:-300+400",
			readInput: func(s string) (string, error) {
				return "m:400-300", nil
			},
			reader:   bufio.NewReader(strings.NewReader("2\n")),
			wantLine: "m:400-300\n",
			wantErr:  nil,
		},
		{
			name:      "invalid->3",
			nameNote:  "nameNote",
			content:   "lunes 1",
			readInput: func(s string) (string, error) { return s, nil },
			reader:    bufio.NewReader(strings.NewReader("3\n")),
			wantLine:  "",
			wantErr:   errSkipNote,
		},
		{
			name:      "invalidOption->erase",
			nameNote:  "nameNote",
			content:   "bad data",
			readInput: func(s string) (string, error) { return s, nil },
			reader:    bufio.NewReader(strings.NewReader("4\n1\n")),
			wantLine:  "",
			wantErr:   nil,
		},
		{
			name:     "option2->readInputError->valid",
			nameNote: "nameNote",
			content:  "m:-300+400",
			readInput: func() func(string) (string, error) {
				calls := 0
				return func(s string) (string, error) {
					calls++
					if calls == 1 {
						return "", fmt.Errorf("input error")
					}
					return "m:400-300", nil
				}
			}(),
			reader:   bufio.NewReader(strings.NewReader("2\n")),
			wantLine: "m:400-300\n",
			wantErr:  nil,
		},
		{
			name:     "option2->invalidEdit->valid",
			nameNote: "nameNote",
			content:  "bad data",
			readInput: func() func(string) (string, error) {
				calls := 0
				return func(s string) (string, error) {
					calls++
					if calls == 1 {
						return "canon", nil
					}
					return "canon 5000", nil
				}
			}(),
			reader:   bufio.NewReader(strings.NewReader("2\n")),
			wantLine: "canon 5000\n",
			wantErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, err := invalidLine(tt.nameNote, tt.content, tt.readInput, tt.reader)
			if line != tt.wantLine {
				t.Errorf("got '%s', want '%s'", line, tt.wantLine)
			}
			if err != tt.wantErr {
				t.Errorf("got '%v', want '%v'", err, tt.wantErr)
			}
		})
	}
}
