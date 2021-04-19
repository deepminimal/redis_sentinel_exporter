package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_Validate(t *testing.T) {
	tests := []struct {
		opts  *Options
		valid bool
	}{
		{
			opts: &Options{
				Addr: "",
			},
			valid: false,
		},
		{
			opts: &Options{
				Addr: "127.0.0.1:8080",
			},
			valid: true,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.valid, test.opts.Validate() == nil)
	}
}
