package articlelisting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatLong(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		want    string
		wantErr bool
	}{
		{name: "well-formed date", date: "2025-12-21", want: "21 décembre 2025"},
		{name: "empty date", date: "", wantErr: true},
		{name: "malformed date", date: "not-a-date", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			got, err := formatLong(tt.date)

			// expect
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
