package main

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCheckNavbar(t *testing.T) {
	pageURL := "https://" + uuid.New().String() + ".example.com/"

	tests := []struct {
		name       string
		body       string
		wantErrSub string
	}{
		{
			name: "non-empty nav passes",
			body: `<nav class="x"><a href="/">Home</a></nav>`,
		},
		{
			name:       "missing nav fails",
			body:       `<div>no nav here</div>`,
			wantErrSub: "missing <nav>",
		},
		{
			name:       "empty nav fails",
			body:       `<nav></nav>`,
			wantErrSub: "empty",
		},
		{
			name:       "whitespace-only nav fails",
			body:       "<nav>   \n  </nav>",
			wantErrSub: "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			err := checkNavbar(pageURL, []byte(tt.body))

			// expect
			if tt.wantErrSub == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErrSub)
			assert.ErrorContains(t, err, pageURL)
		})
	}
}
