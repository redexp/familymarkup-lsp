package providers

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

func ConfigurationChange(_ *Ctx, config *ClientConfiguration) (err error) {
	if config.Locale != "" {
		err = SetLocale(config.Locale)
	}

	warnChildrenWithoutRelations = config.WarnChildrenWithoutRelations

	for _, doc := range root.Docs {
		doc.NeedDiagnostic = true
	}

	return
}

type ClientConfiguration struct {
	Locale                       string `json:"locale" mapstructure:"locale"`
	WarnChildrenWithoutRelations bool   `json:"warnChildrenWithoutRelations" mapstructure:"warnChildrenWithoutRelations"`
}

func GetClientConfiguration(src any) (res ClientConfiguration, err error) {
	err = mapstructure.Decode(src, &res)

	return
}

type ConfigurationHandlers struct {
	Change ConfigChangeFunc
}

func (req *ConfigurationHandlers) Handle(ctx *Ctx) (res any, validMethod bool, validParams bool, err error) {
	switch ctx.Method {
	case ConfigChangeMethod:
		validMethod = true

		var params ClientConfiguration
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			err = req.Change(ctx, &params)
		}
	}

	return
}

const ConfigChangeMethod = "config/change"

type ConfigChangeFunc func(*Ctx, *ClientConfiguration) error
