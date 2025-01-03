package providers

import (
	"github.com/tliron/glsp"
	serv "github.com/tliron/glsp/server"
)

type RequestHandler struct {
	Handlers []glsp.Handler
}

func (req *RequestHandler) Handle(ctx *Ctx) (res any, validMethod bool, validParams bool, err error) {
	for _, h := range req.Handlers {
		res, validMethod, validParams, err = h.Handle(ctx)

		if validMethod {
			Debugf("method: %s, err: %v", ctx.Method, err != nil)
			return
		}
	}

	return
}

func CreateServer(handlers ...glsp.Handler) *serv.Server {
	server = serv.NewServer(&RequestHandler{Handlers: handlers}, "familymarkup", false)
	return server
}
