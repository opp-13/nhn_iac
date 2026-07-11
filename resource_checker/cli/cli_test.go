package cli

import (
	"reflect"
	"testing"
)

func TestExpandShortBoolFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"combined", []string{"-al"}, []string{"-a", "-l"}},
		{"combined reversed", []string{"-la"}, []string{"-l", "-a"}},
		{"already separate", []string{"-a", "-l"}, []string{"-a", "-l"}},
		{"short value flag untouched", []string{"-o", "json"}, []string{"-o", "json"}},
		{"long value flag untouched", []string{"--output", "json"}, []string{"--output", "json"}},
		{"long bool flag untouched", []string{"--all"}, []string{"--all"}},
		{"disallowed letter not expanded", []string{"-ao"}, []string{"-ao"}},
		{"double dash stops expansion", []string{"--", "-al"}, []string{"--", "-al"}},
		{"mixed", []string{"-al", "-o", "json"}, []string{"-a", "-l", "-o", "json"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandShortBoolFlags(tt.args, "al")
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandShortBoolFlags(%v, %q) = %v, want %v", tt.args, "al", got, tt.want)
			}
		})
	}
}
