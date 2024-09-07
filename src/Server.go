package src

import (
	"encoding/json"

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

type CustomHandlers struct {
	TreeFamilies  TreeFamiliesHandler
	TreeRelations TreeRelationsHandler
	TreeMembers   TreeMembersHandler
}

func (req *CustomHandlers) Handle(ctx *glsp.Context) (res any, validMethod bool, validParams bool, err error) {
	switch ctx.Method {
	case TreeFamiliesMethod:
		validMethod = true
		validParams = true
		res, err = req.TreeFamilies(ctx)

	case TreeRelationsMethod:
		validMethod = true

		var params TreeRelationsParams
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.TreeRelations(ctx, &params)
		}

	case TreeMembersMethod:
		validMethod = true

		var params TreeMembersParams
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.TreeMembers(ctx, &params)
		}
	}

	return
}

// TreeFamilies

const TreeFamiliesMethod = "tree/families"

type TreeFamiliesHandler func(ctx *glsp.Context) ([]*TreeFamily, error)

type TreeFamily struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases,omitempty"`
}

// TreeRelations

const TreeRelationsMethod = "tree/relations"

type TreeRelationsHandler func(ctx *glsp.Context, params *TreeRelationsParams) ([]*TreeRelation, error)

type TreeRelationsParams struct {
	FamilyId string `json:"family_id"`
}

type TreeRelation struct {
	Id    uint32 `json:"id"`
	Label string `json:"label"`
	Arrow string `json:"arrow"`
}

// TreeMembers

const TreeMembersMethod = "tree/members"

type TreeMembersHandler func(ctx *glsp.Context, params *TreeMembersParams) ([]*TreeMember, error)

type TreeMembersParams struct {
	FamilyId   string `json:"family_id"`
	RelationId uint32 `json:"relation_id"`
}

type TreeMember struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases,omitempty"`
}
