package handlers

import (
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func SemanticTokensFull(ctx *glsp.Context, params *proto.SemanticTokensParams) (*proto.SemanticTokens, error) {
	logDebug("SemanticTokens/Full req %s", params)

	tree, err := getTree(params.TextDocument.URI)

	if err != nil {
		return nil, err
	}

	list, err := familymarkup.GetHighlightCaptures(tree.RootNode())

	if err != nil {
		return nil, err
	}

	tokens, err := CapturesToSemanticTokens(list)

	if err != nil {
		return nil, err
	}

	res := &proto.SemanticTokens{
		Data: *tokens,
	}

	logDebug("SemanticTokens/Full res %s", res)

	return res, nil
}

func SemanticTokensRange(ctx *glsp.Context, params *proto.SemanticTokensRangeParams) (any, error) {
	logDebug("SemanticTokens/Range req %s", params)

	tree, err := getTree(params.TextDocument.URI)

	if err != nil {
		return nil, err
	}

	nodes := getNodesByRange(tree, &params.Range)
	list := make([]*sitter.QueryCapture, 0)
	startLine := params.Range.Start.Line
	startChar := params.Range.Start.Character
	endLine := params.Range.End.Line
	endChar := params.Range.End.Character

	for _, node := range nodes {
		items, err := familymarkup.GetHighlightCaptures(node)

		if err != nil {
			return nil, err
		}

		for _, cap := range items {
			nodePoint := cap.Node.StartPoint()

			if nodePoint.Row < startLine || (nodePoint.Row == startLine && nodePoint.Column < startChar) {
				continue
			}

			if endLine < nodePoint.Row || (endLine == nodePoint.Row && endChar <= nodePoint.Column) {
				break
			}

			list = append(list, cap)
		}
	}

	tokens, err := CapturesToSemanticTokens(list)

	if err != nil {
		return nil, err
	}

	res := &proto.SemanticTokens{
		Data: *tokens,
	}

	logDebug("SemanticTokens/Range res %s", res)

	return res, nil
}

func CapturesToSemanticTokens(list []*sitter.QueryCapture) (*[]proto.UInteger, error) {
	tokens := make([]proto.UInteger, len(list)*5)

	type Token struct {
		sitter.Point
		TokenType

		Length uint32
	}

	var prev *sitter.Point

	for i, cap := range list {
		node := cap.Node
		start := node.StartPoint()
		end := node.EndPoint()
		token := Token{
			Point:     start,
			TokenType: typesMap[cap.Index],
			Length:    uint32(end.Column - start.Column),
		}

		if prev != nil {
			token.Row = token.Row - prev.Row

			if token.Row == 0 {
				token.Column = token.Column - prev.Column
			}
		}

		prev = &start

		n := i * 5

		tokens[n+0] = token.Row
		tokens[n+1] = token.Column
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
