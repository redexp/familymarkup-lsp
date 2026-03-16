package providers

import (
	"encoding/json"

	"github.com/redexp/familymarkup-lsp/layout"
	. "github.com/redexp/familymarkup-lsp/types"
)

func SvgFamilies(_ *Ctx, params *SvgFamiliesParams) ([]*layout.SvgFamily, error) {
	list := layout.Align(root, layout.AlignParams{
		FontRatio: params.FontRatio,
	})

	return list, nil
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

type SvgFamiliesFunc func(*Ctx, *SvgFamiliesParams) ([]*layout.SvgFamily, error)
