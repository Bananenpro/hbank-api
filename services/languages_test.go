package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLanguageFromAcceptLanguageHeader(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		headerValue string
		want        string
	}{
		{headerValue: "de", want: "de"},
		{headerValue: "de-DE", want: "de"},
		{headerValue: "de-en", want: "de"},
		{headerValue: "de-CH", want: "de"},
		{headerValue: "de,en-US;q=0.7,en;q=0.3", want: "de"},
		{headerValue: "en -US;q = 0.7 , de-CH;q=0.9,en;q = 0.3, ", want: "de"},
		{headerValue: "da, en-gb;q=0.8, en;q=0.7", want: "en"},
		{headerValue: "en-US,fr-CA", want: "en"},
		{headerValue: "*", want: "en"},
		{headerValue: ",,;d;e;,d-e,-;;---;,,q=,,,;;;0.9;", want: "en"},
	}
	for _, tt := range tests {
		t.Run(tt.headerValue, func(t *testing.T) {
			assert.Equal(t, tt.want, GetLanguageFromAcceptLanguageHeader(tt.headerValue))
		})
	}
}
