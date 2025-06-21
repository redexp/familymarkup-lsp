package providers

import (
	"slices"
	"sort"
	"strings"
	"sync"

	fm "github.com/redexp/familymarkup-parser"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

type Tokens []proto.UInteger

type LegendType struct {
	Types     []string
	Modifiers []string
	Map       map[string]Tokens
}

var tokensCache = make(map[Uri]Tokens)

var Legend = CreateLegend([]string{
	"class.declaration.family_name",
	"class.declaration.family_name.alias",
	"property.static.name.ref",
	"class.family_name.ref",
	"property.declaration.static.name.def",
	"property.declaration.static.name.def.alias",
	"string.unknown",
	"comment",
	"number.targets",
	"punctuation.delimiter.sources",
	"punctuation.delimiter.targets",
	"operator.sources.join",
	"operator.arrow",
	"string.label",
})

func SemanticTokensFull(_ *Ctx, params *proto.SemanticTokensParams) (res *proto.SemanticTokens, err error) {
	tokens, uri, err := getSemanticTokens(params.TextDocument.URI)

	if err != nil {
		return
	}

	tokensCache[uri] = tokens

	res = &proto.SemanticTokens{
		Data: tokens,
	}

	return
}

func SemanticTokensDelta(_ *Ctx, params *proto.SemanticTokensDeltaParams) (res any, err error) {
	tokens, uri, err := getSemanticTokens(params.TextDocument.URI)

	if err != nil {
		return
	}

	prevTokens, exist := tokensCache[uri]

	tokensCache[uri] = tokens

	if !exist {
		res = proto.SemanticTokens{
			Data: tokens,
		}

		return
	}

	start, delCount, data := deltaSemanticTokens(prevTokens, tokens)

	res = proto.SemanticTokensEdit{
		Start:       start,
		DeleteCount: delCount,
		Data:        data,
	}

	return
}

func CreateLegend(list []string) (legend *LegendType) {
	types := make([]string, 0)
	modifiers := make([]string, 0)

	legend = &LegendType{
		Map: make(map[string]Tokens),
	}

	for _, name := range list {
		parts := strings.Split(name, ".")
		first := parts[0]
		rest := parts[1:]

		if !slices.Contains(types, first) {
			types = append(types, first)
		}

		modMask := 0

		for _, m := range rest {
			if !slices.Contains(modifiers, m) {
				modifiers = append(modifiers, m)
			}

			modMask = modMask | slices.Index(modifiers, m)
		}

		legend.Map[name] = Tokens{
			proto.UInteger(slices.Index(types, first)),
			proto.UInteger(modMask),
		}
	}

	legend.Types = types
	legend.Modifiers = modifiers

	return
}

func (legend *LegendType) Get(name string) (proto.UInteger, proto.UInteger) {
	list, exist := legend.Map[name]

	if !exist {
		panic("unknown type: " + name)
	}

	return list[0], list[1]
}

func getSemanticTokens(docUri string) (result Tokens, uri string, err error) {
	uri, err = NormalizeUri(docUri)

	if err != nil {
		return
	}

	err = root.UpdateDirty()

	if err != nil {
		return
	}

	doc, ok := root.Docs[uri]

	if !ok {
		return make(Tokens, 0), uri, nil
	}

	type Item struct {
		token *fm.Token
		key   string
	}

	var list []Item

	add := func(token *fm.Token, key string) {
		if token == nil {
			return
		}

		list = append(list, Item{token: token, key: key})
	}

	addArr := func(tokens []*fm.Token, key string) {
		for _, token := range tokens {
			add(token, key)
		}
	}

	addArr(doc.Root.Comments, "comment")

	for _, f := range doc.Root.Families {
		addArr(f.Comments, "comment")
		add(f.Name, "class.declaration.family_name")
		addArr(f.Aliases, "class.declaration.family_name.alias")

		for _, rel := range f.Relations {
			addArr(rel.Comments, "comment")
			add(rel.Arrow, "operator.arrow")
			add(rel.Label, "string.label")

			for listIndex, relList := range []*fm.RelList{rel.Sources, rel.Targets} {
				if relList == nil {
					continue
				}

				for _, p := range relList.Persons {
					add(p.Num, "number.targets")
					addArr(p.Comments, "comment")

					if p.Unknown != nil {
						add(p.Unknown, "string.unknown")
						continue
					}

					key := "property.static.name.ref"

					if p.IsChild {
						key = "property.declaration.static.name.def"
					}

					add(p.Name, key)
					addArr(p.Aliases, "property.declaration.static.name.def.alias")
					add(p.Surname, "class.family_name.ref")
				}

				key := "punctuation.delimiter.sources"

				if listIndex == 1 {
					key = "punctuation.delimiter.targets"
				}

				for _, sep := range relList.Separators {
					if sep.SubType == fm.TokenPlus {
						key = "operator.sources.join"
					}

					add(sep, key)
				}
			}
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].token.Offest < list[j].token.Offest
	})

	result = make(Tokens, len(list)*5)

	var prev *fm.Token

	for i, item := range list {
		token := item.token

		deltaLine := token.Line
		deltaStartChar := token.Char

		if prev != nil {
			deltaLine = token.Line - prev.Line

			if deltaLine == 0 {
				deltaStartChar = token.Char - prev.Char
			}
		}

		t, m := Legend.Get(item.key)

		result[i*5] = proto.UInteger(deltaLine)
		result[i*5+1] = proto.UInteger(deltaStartChar)
		result[i*5+2] = proto.UInteger(token.CharsNum)
		result[i*5+3] = t
		result[i*5+4] = m

		prev = token
	}

	return
}

func min(a, b uint32) uint32 {
	if a <= b {
		return a
	}

	return b
}

func deltaSemanticTokens(prevTokens, tokens Tokens) (st, delCount uint32, insert Tokens) {
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
			delCount = prevLen - curLen
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
