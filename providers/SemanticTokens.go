package providers

import (
	"sync"

	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

type Tokens []proto.UInteger

var tokensMap = make(map[string]Tokens)

func SemanticTokensFull(ctx *Ctx, params *proto.SemanticTokensParams) (res *proto.SemanticTokens, err error) {
	tokens, uri, err := getTokens(params.TextDocument.URI)

	if err != nil {
		return
	}

	tokensMap[uri] = tokens

	res = &proto.SemanticTokens{
		Data: tokens,
	}

	return
}

func SemanticTokensDelta(ctx *Ctx, params *proto.SemanticTokensDeltaParams) (res any, err error) {
	tokens, uri, err := getTokens(params.TextDocument.URI)

	if err != nil {
		return
	}

	prevTokens, exist := tokensMap[uri]

	tokensMap[uri] = tokens

	if !exist {
		res = proto.SemanticTokens{
			Data: tokens,
		}

		return
	}

	start, delCount, data := getTokensDelta(prevTokens, tokens)

	res = proto.SemanticTokensEdit{
		Start:       start,
		DeleteCount: delCount,
		Data:        data,
	}

	return
}

func getTokens(docUri string) (tokens Tokens, uri string, err error) {
	uri, err = NormalizeUri(docUri)

	if err != nil {
		return
	}

	doc, err := root.OpenDoc(uri)

	if err != nil {
		return
	}

	tokens, err = doc.ConvertHighlightCaptures(typesMap)

	return
}

func min(a, b uint32) uint32 {
	if a <= b {
		return a
	}

	return b
}

func getTokensDelta(prevTokens, tokens Tokens) (st, delCount uint32, insert Tokens) {
	prevLen := uint32(len(prevTokens))
	curLen := uint32(len(tokens))
	count := min(prevLen, curLen)

	if prevLen == 0 {
		return 0, 0, tokens
	}

	if curLen == 0 {
		return 0, prevLen, tokens
	}

	var wg sync.WaitGroup

	wg.Add(2)

	start := int32(-1)
	prevEnd := int32(-1)
	curEnd := int32(-1)

	go func() {
		defer wg.Done()

		for i := uint32(0); i < count; i++ {
			if prevTokens[i] != tokens[i] {
				start = int32(i)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()

		for i := uint32(1); i <= count; i++ {
			if prevTokens[prevLen-i] != tokens[curLen-i] {
				prevEnd = int32(prevLen - i)
				curEnd = int32(curLen - i)
				return
			}
		}

		prevEnd = int32(prevLen - count - 1)
		curEnd = int32(curLen - count - 1)
	}()

	wg.Wait()

	if start < 0 {
		if prevLen > curLen {
			st = count
			delCount = uint32(prevLen - curLen)
		} else if prevLen < curLen {
			st = count
			insert = tokens[st:]
		} else {
			return 0, prevLen, tokens
		}
	} else {
		st = uint32(start)

		if prevEnd >= start {
			delCount = uint32(prevEnd - start + 1)
		}

		if curEnd >= start {
			insert = tokens[start : curEnd+1]
		}
	}

	if insert == nil {
		insert = Tokens{}
	}

	return
}
