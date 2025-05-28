package providers

import (
	"slices"
	"testing"
)

func TestTokensDelta(t *testing.T) {
	tokens := Tokens{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	tokensLen := uint32(len(tokens))
	ext := Tokens{0, 0, 0}
	extLen := uint32(len(ext))

	list := []struct {
		prev   Tokens
		tokens Tokens

		st  uint32
		del uint32
		ins Tokens
	}{
		{
			prev:   Tokens{},
			tokens: tokens,
			st:     0,
			del:    0,
			ins:    tokens,
		},
		{
			prev:   tokens,
			tokens: Tokens{},
			st:     0,
			del:    tokensLen,
			ins:    Tokens{},
		},
		{
			prev:   tokens,
			tokens: slices.Concat(ext, tokens),
			st:     0,
			del:    0,
			ins:    ext,
		},
		{
			prev:   tokens,
			tokens: slices.Concat(tokens, ext),
			st:     tokensLen,
			del:    0,
			ins:    ext,
		},
		{
			prev:   tokens,
			tokens: slices.Concat(tokens[0:3], ext, tokens[3:]),
			st:     3,
			del:    0,
			ins:    ext,
		},

		{
			prev:   slices.Concat(ext, tokens),
			tokens: tokens,
			st:     0,
			del:    extLen,
			ins:    Tokens{},
		},
		{
			prev:   slices.Concat(tokens, ext),
			tokens: tokens,
			st:     tokensLen,
			del:    extLen,
			ins:    Tokens{},
		},
		{
			prev:   slices.Concat(tokens[0:3], ext, tokens[3:]),
			tokens: tokens,
			st:     3,
			del:    extLen,
			ins:    Tokens{},
		},

		{
			prev:   tokens,
			tokens: slices.Concat(ext, tokens[3:]),
			st:     0,
			del:    extLen,
			ins:    ext,
		},
		{
			prev:   tokens,
			tokens: slices.Concat(tokens[:tokensLen-extLen], ext),
			st:     tokensLen - extLen,
			del:    extLen,
			ins:    ext,
		},
		{
			prev:   tokens,
			tokens: slices.Concat(tokens[:3], ext, tokens[3+extLen:]),
			st:     3,
			del:    extLen,
			ins:    ext,
		},
	}

	for i, it := range list {
		st, del, ins := deltaSemanticTokens(it.prev, it.tokens)

		if st != it.st || del != it.del || slices.Compare(ins, it.ins) != 0 {
			t.Errorf("test %d, st: %d, del: %d, ins: %v", i+1, st, del, ins)
		}
	}
}
