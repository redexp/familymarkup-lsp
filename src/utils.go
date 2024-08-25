package src

import (
	"encoding/json"
	urlParser "net/url"
	"strings"

	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tliron/glsp"
	serv "github.com/tliron/glsp/server"
)

var logOnly string

func CreateServer(handlers glsp.Handler) {
	server = serv.NewServer(handlers, "familymarkup", false)
	server.RunStdio()
}

func createParser() *sitter.Parser {
	p := sitter.NewParser()
	p.SetLanguage(familymarkup.GetLanguage())
	return p
}

func getParser() *sitter.Parser {
	if parser == nil {
		parser = createParser()
	}

	return parser
}

func logDebug(msg string, data any) {
	if logOnly != "" && !strings.HasPrefix(msg, logOnly) {
		return
	}

	if server == nil || server.Log.GetMaxLevel() < 2 {
		return
	}

	str, _ := json.MarshalIndent(data, "", "  ")
	server.Log.Debugf(msg, str)
}

func Debugf(msg string, args ...any) {
	server.Log.Debugf(msg, args...)
}

func LogOnly(prefix string) {
	logOnly = prefix
}

func uriToPath(uri Uri) (string, error) {
	if strings.HasPrefix(uri, "/") {
		return uri, nil
	}

	url, err := urlParser.Parse(uri)

	if err != nil {
		return "", err
	}

	return url.Path, nil
}

func toUri(path string) Uri {
	if strings.HasPrefix(path, "/") {
		path = "file://" + path
	}

	return path
}

func normalizeUri(uri Uri) (Uri, error) {
	path, err := uriToPath(uri)

	if err != nil {
		return "", err
	}

	return toUri(path), nil
}

// "surname-name", [2]*Node
// "surname-nil", [1]*Node
// "surname", [2]*Node
// "surname", [1]*Node
// "name", [1]*Node
// "nil-name", [1]*Node
// "nil", [0]*Node
func getTypeNode(doc *TextDocument, pos *Position) (t string, nodes []*Node, err error) {
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

func getClosestNode(node *Node, t string) *Node {
	for node != nil && node.Type() != t {
		node = node.Parent()
	}

	return node
}

func getNodeByFields(node *Node, fields ...string) *Node {
	if node == nil {
		return nil
	}

	for _, field := range fields {
		node := node.ChildByFieldName(field)

		if node == nil {
			return nil
		}
	}

	return node
}

func getClosestFamilyName(node *Node) *Node {
	return getNodeByFields(getClosestNode(node, "family"), "name", "name")
}

func nameRefName(node *Node) *Node {
	if isNameRef(node) {
		return node.NamedChild(1)
	}

	return node
}

func isNameAliases(node *Node) bool {
	return node != nil && node.Type() == "name_aliases"
}

func isNameRef(node *Node) bool {
	return node != nil && node.Type() == "name_ref"
}

func isNameDef(node *Node) bool {
	return node != nil && node.Type() == "name_def"
}
