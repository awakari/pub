package model

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrefixes_FindOnePrefix(t *testing.T) {
	p := NewPrefixes[bool]()
	require.Nil(t, p.Put(context.TODO(), "foo", false))
	require.Nil(t, p.Put(context.TODO(), "bar", true))
	cases := map[string]struct {
		in     string
		prefix string
		out    bool
		err    error
	}{
		"empty": {},
		"foo": {
			in:     "foo",
			prefix: "foo",
		},
		"bar42": {
			in:     "bar42",
			prefix: "bar",
			out:    true,
		},
		"fog123": {
			in: "fog123",
		},
		"ba": {
			in: "ba",
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			prefix, out, err := p.FindOnePrefix(context.TODO(), c.in)
			assert.Equal(t, prefix, c.prefix)
			assert.Equal(t, c.out, out)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
