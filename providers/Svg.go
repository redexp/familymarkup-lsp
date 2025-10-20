package providers

import (
	"encoding/json"

	"github.com/redexp/familymarkup-lsp/layout"
	. "github.com/redexp/familymarkup-lsp/types"
)

func SvgDocument(_ *Ctx, params *SvgDocumentParams) ([]*layout.SvgFamily, error) {
	uri := NormalizeUri(params.URI)

	list := layout.Align(root, uri, layout.AlignParams{
		FontRatio: params.FontRatio,
	})

	return list, nil
}

type SvgHandlers struct {
	Document SvgDocumentFunc
}

func (req *SvgHandlers) Handle(ctx *Ctx) (res any, validMethod bool, validParams bool, err error) {
	switch ctx.Method {
	case SvgDocumentMethod:
		validMethod = true

		var params SvgDocumentParams
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.Document(ctx, &params)
		}
	}

	return
}

const SvgDocumentMethod = "svg/document"

type SvgDocumentParams struct {
	URI       Uri     `json:"URI"`
	FontRatio float64 `json:"fontRatio"`
}

type SvgDocumentFunc func(*Ctx, *SvgDocumentParams) ([]*layout.SvgFamily, error)
