package htmlref

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKind_String(t *testing.T) {
	tests := []struct {
		name string
		kind Kind
		want string
	}{
		{name: "skip", kind: KindSkip, want: "skip"},
		{name: "external", kind: KindExternal, want: "external"},
		{name: "local", kind: KindLocal, want: "local"},
		{name: "out of range", kind: Kind(42), want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test
			got := tt.kind.String()

			// expect
			assert.Equal(t, tt.want, got)
		})
	}
}
