package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrToBool(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		value string
		want  bool
	}{
		{value: "t", want: true},
		{value: "true", want: true},
		{value: "y", want: true},
		{value: "yes", want: true},
		{value: "on", want: true},

		{value: "o", want: false},
		{value: "", want: false},
		{value: "truee", want: false},
		{value: "false", want: false},
		{value: "no", want: false},
		{value: "off", want: false},
		{value: " ", want: false},
		{value: strings.Repeat("a", 64), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			assert.Equal(t, tt.want, StrToBool(tt.value))
		})
	}
}
