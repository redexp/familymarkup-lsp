package providers

import (
	"encoding/json"

	"github.com/redexp/familymarkup-lsp/layout"
	. "github.com/redexp/familymarkup-lsp/types"
)

func SvgFamilies(_ *Ctx, params *SvgFamiliesParams) (SvgFamiliesResult, error) {
	families, relations := layout.Align(root, layout.AlignParams{
		FontRatio: params.FontRatio,
	})

	return SvgFamiliesResult{
		Families:  families,
		Relations: relations,
	}, nil
}

type SvgHandlers struct {
	Families SvgFamiliesFunc
}

func (req *SvgHandlers) Handle(ctx *Ctx) (res any, validMethod bool, validParams bool, err error) {
	switch ctx.Method {
	case SvgFamiliesMethod:
		validMethod = true

		var params SvgFamiliesParams
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.Families(ctx, &params)
		}
	}

	return
}

const SvgFamiliesMethod = "svg/families"

type SvgFamiliesParams struct {
	URI       Uri     `json:"URI"`
	FontRatio float64 `json:"fontRatio"`
}

type SvgFamiliesResult struct {
	Families  []*layout.SvgFamily   `json:"families"`
	Relations []*layout.SvgRelation `json:"relations"`
}

type SvgFamiliesFunc func(*Ctx, *SvgFamiliesParams) (SvgFamiliesResult, error)
