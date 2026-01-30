package handler

import (
	"net/http"
	"skyblock-pv-backend/internal"
	"slices"
)

type RequestHandler interface {
	Handle(internal.RouteContext, http.ResponseWriter, *http.Request)
}

// passthrough, allowing the request to be handled normally

type PassthroughRequestHandler struct {
	Handler func(internal.RouteContext, http.ResponseWriter, *http.Request)
}

func (handler PassthroughRequestHandler) Handle(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	handler.Handler(ctx, res, req)
}

// private, requires an authentication key, could be a guest key

type PrivateRequestHandler struct {
	Handler func(internal.RouteContext, internal.AuthenticationContext, http.ResponseWriter, *http.Request)
}

func (handler PrivateRequestHandler) Handle(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	context := internal.GetAuthenticatedContext(ctx, req.Header.Get("Authorization"))
	if context == nil {
		res.WriteHeader(http.StatusUnauthorized)
	} else {
		handler.Handler(ctx, *context, res, req)
	}
}

// authenticated, requires a full authentication key

type AuthenticatedRequestHandler struct {
	Handler func(internal.RouteContext, internal.AuthenticationContext, http.ResponseWriter, *http.Request)
}

func (handler AuthenticatedRequestHandler) Handle(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	context := internal.GetAuthenticatedContext(ctx, req.Header.Get("Authorization"))
	if context == nil || context.IsGuest {
		res.WriteHeader(http.StatusUnauthorized)
	} else {
		handler.Handler(ctx, *context, res, req)
	}
}

// admin, requires an admin authentication key

type AdminRequestHandler struct {
	Handler func(internal.RouteContext, internal.AuthenticationContext, http.ResponseWriter, *http.Request)
}

func (handler AdminRequestHandler) Handle(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	context := internal.GetAuthenticatedContext(ctx, req.Header.Get("Authorization"))
	if context == nil || context.IsGuest || !slices.Contains(ctx.Config.Admins, context.Requester) {
		res.WriteHeader(http.StatusNotFound)
	} else {
		handler.Handler(ctx, *context, res, req)
	}
}

// not implemented, returns 405 Method Not Allowed

type NotImplementedRequestHandler struct {
}

func (handler NotImplementedRequestHandler) Handle(_ internal.RouteContext, res http.ResponseWriter, _ *http.Request) {
	res.WriteHeader(http.StatusMethodNotAllowed)
}
