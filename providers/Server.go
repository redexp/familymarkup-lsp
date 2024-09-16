package providers

import (
	"github.com/tliron/glsp"
	serv "github.com/tliron/glsp/server"
)

type RequestHandler struct {
	Handlers []glsp.Handler
}

func (req *RequestHandler) Handle(ctx *glsp.Context) (res any, validMethod bool, validParams bool, err error) {
	for _, h := range req.Handlers {
		res, validMethod, validParams, err = h.Handle(ctx)

		if validMethod {
			return
		}
	}

	return
}

func CreateServer(handlers ...glsp.Handler) {
	server = serv.NewServer(&RequestHandler{Handlers: handlers}, "familymarkup", false)
	server.RunStdio()
}
