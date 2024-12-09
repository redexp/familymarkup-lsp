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

var parsersPool = make([]*ParserWorker, 0)
var lang = familymarkup.GetLanguage()

var FamilyExt = []string{"fml", "family"}
var MarkdownExt = []string{"md", "mdx"}
var AllExt = slices.Concat(FamilyExt, MarkdownExt)

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

func (p *ParserWorker) Parse(text []byte) (tree *Tree, err error) {
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

func Ext(path string) string {
	return strings.ToLower(strings.TrimLeft(filepath.Ext(path), "."))
}

func IsFamilyUri(uri Uri) bool {
	return slices.Contains(FamilyExt, Ext(uri))
}

func IsMarkdownUri(uri Uri) bool {
	return slices.Contains(MarkdownExt, Ext(uri))
}

// "name" || "surname", [Node]
// "name surname|", [Node, Node]
// "name |", [Node]
// "name| surname", [Node, Node]
// "| surname", [Node]
// "nil", [Node]
func GetTypeNode(doc *TextDocument, pos *Position) (t string, nodes []*Node, err error) {
	prev, target, next, err := doc.GetClosestHighlightCaptureByPosition(pos)

	if err != nil {
		return
	}

	caps := []*QueryCapture{prev, target, next}
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
		parentType := ""
		if parent != nil {
			parentType = parent.Type()
		}

		if parentType == "name_aliases" || parentType == "new_surname" {
			if i != 1 {
				continue
			}

			return nt, []*Node{node}, nil
		}

		nodes[i] = cap.Node
	}

	if nodes[0] != nil {
		if nodes[1] != nil {
			return "name surname|", nodes[0:2], nil
		}

		return "name |", nodes[0:1], nil
	}

	node := nodes[1]

	if node != nil {
		if nodes[2] != nil {
			return "name| surname", nodes[1:3], nil
		}

		t = node.Type()
		p := node.Parent()
		nodes = []*Node{node}

		if p != nil && p.Type() == "family_name" {
			t = "surname"
			return
		}

		return
	}

	if nodes[2] != nil {
		return "| surname", nodes[2:3], nil
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

func GetNameSurname(name_ref *Node) (name *Node, surname *Node) {
	name = name_ref.NamedChild(0)
	surname = name_ref.NamedChild(1)

	return
}

func ToNameNode(node *Node) *Node {
	if IsNameRef(node) {
		name, _ := GetNameSurname(node)

		return name
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

func IsFamilyRelation(node *Node) bool {
	rel := GetClosestNode(node, "relation")

	if rel == nil {
		return false
	}

	arrow := rel.ChildByFieldName("arrow")

	return arrow != nil && arrow.Type() == "eq"
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
	return func(yield func(*Node) bool) {
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
