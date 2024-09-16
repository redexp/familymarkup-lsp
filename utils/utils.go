package utils

import (
	"context"
	"iter"
	urlParser "net/url"
	"path/filepath"
	"slices"
	"strings"

	. "github.com/redexp/familymarkup-lsp/types"
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
)

type ParserWorker struct {
	parser *sitter.Parser
	busy   bool
}

var logOnly string
var parsersPool = make([]*ParserWorker, 0)
var lang = familymarkup.GetLanguage()

func CreateParser() *sitter.Parser {
	p := sitter.NewParser()
	p.SetLanguage(familymarkup.GetLanguage())
	return p
}

func GetParser() *ParserWorker {
	var parser *ParserWorker

	for _, p := range parsersPool {
		if !p.busy {
			return p
		}
	}

	parser = &ParserWorker{
		parser: CreateParser(),
		busy:   true,
	}

	parsersPool = append(parsersPool, parser)

	return parser
}

func (p *ParserWorker) Parse(text []byte) (tree *sitter.Tree, err error) {
	p.busy = true

	tree, err = p.parser.ParseCtx(context.Background(), nil, text)

	if len(parsersPool) > 1 {
		p.parser.Close()
		index := slices.Index(parsersPool, p)
		parsersPool = slices.Delete(parsersPool, index, index+1)
	}

	p.busy = false

	return
}

func UriToPath(uri Uri) (string, error) {
	if strings.HasPrefix(uri, "/") {
		return uri, nil
	}

	url, err := urlParser.Parse(uri)

	if err != nil {
		return "", err
	}

	return url.Path, nil
}

func ToUri(path string) Uri {
	if strings.HasPrefix(path, "/") {
		path = "file://" + path
	}

	return path
}

func NormalizeUri(uri Uri) (Uri, error) {
	path, err := UriToPath(uri)

	if err != nil {
		return "", err
	}

	return ToUri(path), nil
}

func RenameUri(uri Uri, name string) (Uri, error) {
	base, err := UriToPath(uri)

	if err != nil {
		return "", err
	}

	return ToUri(filepath.Join(base, "..", name+filepath.Ext(base))), nil
}

func IsUriName(uri Uri, name string) bool {
	base := filepath.Base(uri)
	ext := filepath.Ext(uri)

	return name+ext == base
}

// "surname-name", [2]*Node
// "surname-nil", [1]*Node
// "surname", [2]*Node
// "surname", [1]*Node
// "name", [1]*Node
// "nil-name", [1]*Node
// "nil", [0]*Node
func GetTypeNode(doc *TextDocument, pos *Position) (t string, nodes []*Node, err error) {
	prev, target, next, err := doc.GetClosestHighlightCaptureByPosition(pos)

	if err != nil {
		return
	}

	caps := []*sitter.QueryCapture{prev, target, next}
	nodes = make([]*Node, 3)
	line := pos.Line

	for i, cap := range caps {
		if cap == nil {
			continue
		}

		node := cap.Node
		nt := node.Type()

		if (nt != "name" && nt != "surname") || node.StartPoint().Row != line {
			continue
		}

		parent := node.Parent()

		if parent != nil && parent.Type() == "name_aliases" {
			if i != 1 {
				continue
			}

			return "name", []*Node{cap.Node}, nil
		}

		nodes[i] = cap.Node
	}

	if nodes[0] != nil {
		if nodes[1] != nil {
			return "surname-name", nodes[0:2], nil
		}

		return "surname-nil", nodes[0:1], nil
	}

	if nodes[1] != nil {
		if nodes[2] != nil {
			return "surname", nodes[1:3], nil
		}

		t = "name"
		p := nodes[1].Parent()
		nodes = nodes[1:2]

		if p != nil && p.Type() == "family_name" {
			t = "surname"
			return
		}

		return
	}

	if nodes[2] != nil {
		return "nil-name", nodes[2:3], nil
	}

	return "nil", []*Node{}, nil
}

func GetClosestNode(node *Node, parentType string, fields ...string) *Node {
	for node != nil && node.Type() != parentType {
		node = node.Parent()
	}

	if node != nil && len(fields) > 0 {
		return GetNodeByFields(node, fields...)
	}

	return node
}

func GetNodeByFields(node *Node, fields ...string) *Node {
	if node == nil {
		return nil
	}

	for _, field := range fields {
		node = node.ChildByFieldName(field)

		if node == nil {
			return nil
		}
	}

	return node
}

func GetClosestFamilyName(node *Node) *Node {
	return GetNodeByFields(GetClosestNode(node, "family"), "name", "name")
}

func GetClosestSources(node *Node) *Node {
	return GetClosestNode(node, "relation", "sources")
}

func NameRefName(node *Node) *Node {
	if IsNameRef(node) {
		return node.NamedChild(1)
	}

	return node
}

func IsFamilyName(node *Node) bool {
	return node != nil && node.Type() == "family_name"
}

func IsNameAliases(node *Node) bool {
	return node != nil && node.Type() == "name_aliases"
}

func IsNameRef(node *Node) bool {
	return node != nil && node.Type() == "name_ref"
}

func IsNameDef(node *Node) bool {
	return node != nil && node.Type() == "name_def"
}

func IsNewSurname(node *Node) bool {
	return node != nil && node.Type() == "new_surname"
}

func IsNumUnknown(node *Node) bool {
	return node != nil && node.Type() == "num_unknown"
}

func P[T ~string | ~int32](src T) *T {
	return &src
}

func CreateQuery(pattern string) (*sitter.Query, error) {
	return sitter.NewQuery([]byte(pattern), lang)
}

func CreateCursor(q *sitter.Query, node *Node) *sitter.QueryCursor {
	c := sitter.NewQueryCursor()
	c.Exec(q, node)
	return c
}

func QueryIter(q *sitter.Query, node *Node) iter.Seq2[uint32, *Node] {
	c := CreateCursor(q, node)

	return func(yield func(uint32, *Node) bool) {
		defer c.Close()
		defer q.Close()

		for {
			match, ok := c.NextMatch()

			if !ok {
				break
			}

			for _, cap := range match.Captures {
				if !yield(cap.Index, cap.Node) {
					return
				}
			}
		}
	}
}

func GetErrorNodesIter(root *Node) iter.Seq[*Node] {
	return func(yield func(*sitter.Node) bool) {
		if !root.HasError() {
			return
		}

		c := sitter.NewTreeCursor(root)
		defer c.Close()

		active := true
		var traverse func()

		traverse = func() {
			if !active {
				return
			}

			node := c.CurrentNode()

			if node.IsError() {
				active = yield(node)
				return
			}

			if !node.HasError() {
				return
			}

			if !c.GoToFirstChild() {
				return
			}

			for {
				traverse()

				if !active {
					return
				}

				if !c.GoToNextSibling() {
					break
				}
			}

			c.GoToParent()
		}

		traverse()
	}
}
