package main

import (
	"testing"
)

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
		{"m:-4500", true},
		{"viernes 3/10 canon 7500", true}, // this is false
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
		{"m:2000++2500+4500+4500+4000-2000", false}, // TODO: if i update the regex, will have to change this
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
