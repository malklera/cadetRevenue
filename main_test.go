package main

import (
	"testing"
)

func TestCheckPadding(t *testing.T) {
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
			ans := checkPadding(tt.day)
			if ans != tt.want {
				t.Errorf("got '%s', want '%s'", ans, tt.want)
			}
		})
	}
}
