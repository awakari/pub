package model

import (
	"context"
	"github.com/porfirion/trie"
)

type Prefixes[T any] interface {
	Put(ctx context.Context, prefix string, v T) (err error)
	FindOne(ctx context.Context, input string) (prefix string, v T, err error)
}

type prefixes[T any] struct {
	t *trie.Trie[T]
}

const inputLengthMax = 64

func NewPrefixes[T any]() Prefixes[T] {
	return prefixes[T]{
		t: &trie.Trie[T]{},
	}
}

func (p prefixes[T]) Put(ctx context.Context, prefix string, v T) (err error) {
	p.t.PutString(prefix, v)
	return
}

func (p prefixes[T]) FindOne(ctx context.Context, input string) (prefix string, v T, err error) {
	if len(input) > inputLengthMax {
		input = input[:inputLengthMax]
	}
	var length int
	var ok bool
	v, length, ok = p.t.SearchPrefixInString(input)
	if ok {
		prefix = input[:length]
	}
	return
}
