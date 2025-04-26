package main

import (
	"fmt"
	"net/http"
	"skyblock-pv-backend/auctions"
	"skyblock-pv-backend/routes"
	"skyblock-pv-backend/routes/utils"
	"time"
)

func setDefaults(route *RequestRoute) {
	if route.Get == nil {
		route.Get = NotImplementedRequestHandler{}
	}
	if route.Post == nil {
		route.Post = NotImplementedRequestHandler{}
	}
	if route.Put == nil {
		route.Put = NotImplementedRequestHandler{}
	}
	if route.Delete == nil {
		route.Delete = NotImplementedRequestHandler{}
	}
}

type RequestRoute struct {
	Get    AbstractRequestHandler
	Post   AbstractRequestHandler
	Put    AbstractRequestHandler
	Delete AbstractRequestHandler
}

type AbstractRequestHandler interface {
	handle(http.ResponseWriter, *http.Request)
}

type NotImplementedRequestHandler struct{}

func authenticated(handler func(utils.RouteContext, utils.AuthenticationContext, http.ResponseWriter, *http.Request)) AuthenticatedRequestHandler {
	return AuthenticatedRequestHandler{handler: handler}
}

func public(handler func(utils.RouteContext, http.ResponseWriter, *http.Request)) RequestHandler {
	return RequestHandler{handler: handler}
}

type RequestHandler struct {
	handler func(utils.RouteContext, http.ResponseWriter, *http.Request)
}

type AuthenticatedRequestHandler struct {
	handler func(utils.RouteContext, utils.AuthenticationContext, http.ResponseWriter, *http.Request)
}

var routeContext = utils.NewRouteContext()

func (not NotImplementedRequestHandler) handle(res http.ResponseWriter, _ *http.Request) {
	res.WriteHeader(http.StatusMethodNotAllowed)
}

func (normal RequestHandler) handle(res http.ResponseWriter, req *http.Request) {
	normal.handler(routeContext, res, req)
}

func (authenticated AuthenticatedRequestHandler) handle(res http.ResponseWriter, req *http.Request) {
	context := utils.GetAuthenticatedContext(routeContext, req.Header.Get("Authorization"))
	if context == nil {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}
	authenticated.handler(routeContext, *context, res, req)
}

func create(handlers RequestRoute) func(http.ResponseWriter, *http.Request) {
	setDefaults(&handlers)
	return func(res http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			handlers.Get.handle(res, req)
		case "POST":
			handlers.Post.handle(res, req)
		case "PUT":
			handlers.Put.handle(res, req)
		case "DELETE":
			handlers.Delete.handle(res, req)
		default:
			res.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func fetchData() {
	err := auctions.FetchAll(&routeContext)
	if err != nil {
		panic(err) // panic because we just started
	}
	updateData := time.NewTicker(time.Hour)
	for {
		select {
		case <-updateData.C:
			err = auctions.FetchAll(&routeContext)
			if err != nil {
				fmt.Printf("Error fetching auctions: %v\n", err)
				fmt.Print("Trying again later, using current data.")
			}
		}
	}
}

func main() {
	go fetchData()
	http.HandleFunc("/authenticate", create(RequestRoute{
		Get: public(routes.Authenticate),
	}))
	http.HandleFunc("/profiles/{id}", create(RequestRoute{
		Get: authenticated(routes.GetProfiles),
	}))
	http.HandleFunc("/garden/{profile}", create(RequestRoute{
		Get: authenticated(routes.GetGarden),
	}))
	http.HandleFunc("/museum/{profile}", create(RequestRoute{
		Get: authenticated(routes.GetMuseum),
	}))
	http.HandleFunc("/status/{id}", create(RequestRoute{
		Get: authenticated(routes.GetStatus),
	}))
	http.HandleFunc("/auctions/{profile}", create(RequestRoute{
		Get: authenticated(routes.GetActiveProfileAuctions),
	}))

	http.HandleFunc("/auctions", create(RequestRoute{
		Get: public(routes.GetLbin),
	}))

	fmt.Printf("Listening on 0.0.0.0:%s\n", routeContext.Config.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", routeContext.Config.Port), nil)

	if err != nil {
		panic(err)
	}
}
