package providers

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
	"github.com/tliron/glsp"
)

func ConfigurationChange(ctx *glsp.Context, config *ClientConfiguration) (err error) {
	if config.Locale != "" {
		err = SetLocale(config.Locale)

		if err != nil {
			return
		}

		diagnosticOpenDocs(ctx)
	}

	return
}

type ClientConfiguration struct {
	Locale string `json:"locale"`
}

func GetClientConfiguration(src any) (res ClientConfiguration, err error) {
	err = mapstructure.Decode(src, &res)

	return
}

type ConfigurationHandlers struct {
	Change ConfigChangeFunc
}

func (req *ConfigurationHandlers) Handle(ctx *glsp.Context) (res any, validMethod bool, validParams bool, err error) {
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

type ConfigChangeFunc func(*glsp.Context, *ClientConfiguration) error