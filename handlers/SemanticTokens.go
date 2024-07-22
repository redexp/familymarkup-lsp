package handlers

import (
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func SemanticTokensFull(ctx *glsp.Context, params *proto.SemanticTokensParams) (*proto.SemanticTokens, error) {
	logDebug("SemanticTokens/Full req %s", params)

	doc, err := openDoc(params.TextDocument.URI)

	if err != nil {
		return nil, err
	}

	list, err := GetCaptures(doc.Tree.RootNode())

	if err != nil {
		return nil, err
	}

	tokens, err := CapturesToSemanticTokens(list, doc)

	if err != nil {
		return nil, err
	}

	res := &proto.SemanticTokens{
		Data: *tokens,
	}

	logDebug("SemanticTokens/Full res %s", "res")

	return res, nil
}

func SemanticTokensRange(ctx *glsp.Context, params *proto.SemanticTokensRangeParams) (any, error) {
	logDebug("SemanticTokens/Range req %s", params)

	doc, err := openDoc(params.TextDocument.URI)

	if err != nil {
		return nil, err
	}

	tree := doc.Tree

	nodes := getNodesByRange(tree, &params.Range)
	list := make([]*sitter.QueryCapture, 0)
	startLine := params.Range.Start.Line
	startChar := params.Range.Start.Character
	endLine := params.Range.End.Line
	endChar := params.Range.End.Character

	for _, node := range nodes {
		items, err := GetCaptures(node)

		if err != nil {
			return nil, err
		}

		for _, cap := range items {
			nodePos, err := doc.PointToPosition(cap.Node.StartPoint())

			if err != nil {
				return nil, err
			}

			if nodePos.Line < startLine || (nodePos.Line == startLine && nodePos.Character < startChar) {
				continue
			}

			if endLine < nodePos.Line || (endLine == nodePos.Line && endChar <= nodePos.Character) {
				break
			}

			list = append(list, cap)
		}
	}

	tokens, err := CapturesToSemanticTokens(list, doc)

	if err != nil {
		return nil, err
	}

	res := &proto.SemanticTokens{
		Data: *tokens,
	}

	logDebug("SemanticTokens/Range res %s", "res")

	return res, nil
}

func GetCaptures(root *sitter.Node) ([]*sitter.QueryCapture, error) {
	caps, err := familymarkup.GetHighlightCaptures(root)

	if err != nil {
		return nil, err
	}

	list := []*sitter.QueryCapture{}

	for _, cap := range caps {
		if cap.Node.IsMissing() {
			continue
		}

		list = append(list, cap)
	}

	return list, nil
}

func CapturesToSemanticTokens(list []*sitter.QueryCapture, doc *TextDocument) (*[]proto.UInteger, error) {
	tokens := make([]proto.UInteger, len(list)*5)

	type Token struct {
		proto.Position
		TokenType

		Length uint32
	}

	var prev *proto.Position

	for i, cap := range list {
		node := cap.Node
		start, err1 := doc.PointToPosition(node.StartPoint())
		end, err2 := doc.PointToPosition(node.EndPoint())

		if someError(err1, err2) {
			if err2 != nil {
				p := node.EndPoint()
				logDebug("err EndPoint %s", p)
				logDebug("Text %s", doc.Text)
				logDebug("Tree %s", doc.Tree.RootNode().String())

				m := getMissingChild(node)

				if m != nil {
					logDebug("missing %s", m.String())
				} else {
					logDebug("missing is %s", "nil")
				}
			}

			return nil, findError(err1, err2)
		}

		token := Token{
			Position:  *start,
			TokenType: typesMap[cap.Index],
			Length:    uint32(end.Character - start.Character),
		}

		if prev != nil {
			token.Line = token.Line - prev.Line

			if token.Line == 0 {
				token.Character = token.Character - prev.Character
			}
		}

		prev = start

		n := i * 5

		tokens[n+0] = token.Line
		tokens[n+1] = token.Character
		tokens[n+2] = token.Length
		tokens[n+3] = token.Type
		tokens[n+4] = token.Mod
	}

	return &tokens, nil
}

func getNodesByRange(tree *sitter.Tree, r *proto.Range) (targets []*sitter.Node) {
	selectStartLine := r.Start.Line
	selectEndLine := r.End.Line

	c := sitter.NewTreeCursor(tree.RootNode())
	defer c.Close()

	targets = make([]*sitter.Node, 0)

	if !c.GoToFirstChild() {
		return
	}

	// -1 - node before range
	//  0 - node inside range
	//  1 - node overlaps range
	//  2 - node after range
	getPos := func(node *sitter.Node) int8 {
		startLine := node.StartPoint().Row
		endLine := node.EndPoint().Row

		if endLine < selectStartLine {
			return -1
		}

		if selectStartLine <= startLine && endLine <= selectEndLine {
			return 0
		}

		if selectEndLine < startLine {
			return 2
		}

		return 1
	}

	for {
		family := c.CurrentNode()
		pos := getPos(family)

		if pos == 0 {
			targets = append(targets, family)
		}

		if pos <= 0 {
			if c.GoToNextSibling() {
				continue
			}

			break
		}

		if pos == 2 {
			return
		}

		c.GoToFirstChild()

		for {
			node := c.CurrentNode()
			pos = getPos(node)

			if pos == 0 {
				targets = append(targets, node)
			}

			if pos <= 0 {
				if c.GoToNextSibling() {
					continue
				}

				break
			}

			if pos == 2 {
				return
			}

			if node.Type() == "relations" {
				c.GoToFirstChild()

				for {
					rel := c.CurrentNode()
					pos = getPos(rel)

					if pos == 0 || pos == 1 {
						targets = append(targets, rel)
					}

					if pos == 2 {
						return
					}

					if !c.GoToNextSibling() {
						break
					}
				}

				c.GoToParent()
			} else {
				targets = append(targets, node)
			}

			if !c.GoToNextSibling() {
				break
			}
		}

		c.GoToParent()

		if !c.GoToNextSibling() {
			break
		}
	}

	return targets
}

func getMissingChild(node *sitter.Node) *sitter.Node {
	logDebug("getMissingChild %s", node.Type())

	count := int(node.ChildCount())

	logDebug("getMissingChild %s", count)

	for i := 0; i < count; i++ {
		child := node.Child(i)

		logDebug("getMissingChild %s", child.Type())

		if child.Type() == "MISSING" {
			return child
		}
	}

	return nil
}
