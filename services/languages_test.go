package services

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLanguageFromAcceptLanguageHeader(t *testing.T) {
	supportedTranslations = []string{
		"de",
		"fr",
	}
	tests := []struct {
		headerValue string
		want        string
	}{
		{headerValue: "de", want: "de"},
		{headerValue: "de;q=0.89,fr;q=0.9", want: "fr"},
		{headerValue: "de-DE", want: "de"},
		{headerValue: "de-en", want: "de"},
		{headerValue: "de-CH", want: "de"},
		{headerValue: "de,en-US;q=0.7,en;q=0.3", want: "de"},
		{headerValue: "en -US;q = 0.7 , de-CH;q=0.9,en;q = 0.3, ", want: "de"},
		{headerValue: "da, en-gb;q=0.8, en;q=0.7", want: "en"},
		{headerValue: "*", want: "en"},
		{headerValue: ",,;d;e;,d-e,-;;---;,,q=,,,;;;0.9;", want: "en"},
	}
	for _, tt := range tests {
		t.Run(tt.headerValue, func(t *testing.T) {
			assert.Equal(t, tt.want, GetLanguageFromAcceptLanguageHeader(tt.headerValue))
		})
	}
}

func Test_parseTranslationFile(t *testing.T) {
	tests := []struct {
		content    string
		wantErr    bool
		wantKeys   []string
		wantValues []string
	}{
		{
			content: "\"Hello\"=\"Hallo\"\n   	\n \"I hope this works\"=\"Ich hoffe das funktioniert\" \n\"What about \"quotation marks\"\"=\"Was ist mit \"Anführungszeichen\"\"\n",
			wantErr: false,
			wantKeys: []string{
				"Hello",
				"I hope this works",
				"What about \"quotation marks\"",
			},
			wantValues: []string{
				"Hallo",
				"Ich hoffe das funktioniert",
				"Was ist mit \"Anführungszeichen\"",
			},
		},
		{
			content: "\"Hello\"=\"Hallo\"\n  a 	\n \"I hope this works\"=\"Ich hoffe das funktioniert\" \n\"What about \"quotation marks\"\"=\"Was ist mit \"Anführungszeichen\"\"\n",
			wantErr: true,
		},
		{
			content: "Hello=Hallo",
			wantErr: true,
		},
		{
			content: "\"Hello\"=Hallo",
			wantErr: true,
		},
		{
			content: "Hello=\"Hallo\"",
			wantErr: true,
		},
		{
			content: "\"Hello\"=\"\nHallo\"",
			wantErr: true,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			lang, err := parseTranslationFile(tt.content)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.wantKeys), len(lang))
				for k, v := range lang {
					assert.Contains(t, tt.wantKeys, k)
					assert.Contains(t, tt.wantValues, v)
				}
			}
		})
	}
}
