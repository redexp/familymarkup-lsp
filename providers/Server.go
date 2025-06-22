package providers

import (
	"github.com/tliron/glsp"
)

type RequestHandler struct {
	Handlers []glsp.Handler
}

func (req *RequestHandler) Handle(ctx *Ctx) (res any, validMethod bool, validParams bool, err error) {
	for _, h := range req.Handlers {
		res, validMethod, validParams, err = h.Handle(ctx)

		if validMethod {
			return
		}
	}

	return
}

func CreateRequestHandler(handlers ...glsp.Handler) *RequestHandler {
	return &RequestHandler{Handlers: handlers}
}
