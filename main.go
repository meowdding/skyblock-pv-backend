package main

import (
	"fmt"
	"net/http"
	"skyblock-pv-backend/auctions"
	"skyblock-pv-backend/internal"
	"skyblock-pv-backend/routes"
	"skyblock-pv-backend/routes/handler"
	"time"
)

func setDefaults(route *RequestRoute) {
	if route.Get == nil {
		route.Get = handler.NotImplementedRequestHandler{}
	}
	if route.Post == nil {
		route.Post = handler.NotImplementedRequestHandler{}
	}
	if route.Put == nil {
		route.Put = handler.NotImplementedRequestHandler{}
	}
	if route.Delete == nil {
		route.Delete = handler.NotImplementedRequestHandler{}
	}
}

type RequestRoute struct {
	Get    handler.RequestHandler
	Post   handler.RequestHandler
	Put    handler.RequestHandler
	Delete handler.RequestHandler
}

func private(function func(internal.RouteContext, internal.AuthenticationContext, http.ResponseWriter, *http.Request)) handler.PrivateRequestHandler {
	return handler.PrivateRequestHandler{Handler: function}
}

func authenticated(function func(internal.RouteContext, internal.AuthenticationContext, http.ResponseWriter, *http.Request)) handler.AuthenticatedRequestHandler {
	return handler.AuthenticatedRequestHandler{Handler: function}
}

func admin(function func(internal.RouteContext, internal.AuthenticationContext, http.ResponseWriter, *http.Request)) handler.AdminRequestHandler {
	return handler.AdminRequestHandler{Handler: function}
}

func public(function func(internal.RouteContext, http.ResponseWriter, *http.Request)) handler.RequestHandler {
	return handler.PassthroughRequestHandler{Handler: function}
}

var routeContext = internal.NewRouteContext()

func create(handlers RequestRoute) func(http.ResponseWriter, *http.Request) {
	setDefaults(&handlers)
	return func(res http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			handlers.Get.Handle(routeContext, res, req)
		case "POST":
			handlers.Post.Handle(routeContext, res, req)
		case "PUT":
			handlers.Put.Handle(routeContext, res, req)
		case "DELETE":
			handlers.Delete.Handle(routeContext, res, req)
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
		Get: private(routes.GetProfiles),
	}))
	http.HandleFunc("/garden/{profile}", create(RequestRoute{
		Get: private(routes.GetGarden),
	}))
	http.HandleFunc("/museum/{profile}", create(RequestRoute{
		Get: private(routes.GetMuseum),
	}))
	http.HandleFunc("/status/{id}", create(RequestRoute{
		Get: private(routes.GetStatus),
	}))
	http.HandleFunc("/guild/{id}", create(RequestRoute{
		Get: private(routes.GetGuild),
	}))
	http.HandleFunc("/auctions/{profile}", create(RequestRoute{
		Get: private(routes.GetActiveProfileAuctions),
	}))
	http.HandleFunc("/player/{id}", create(RequestRoute{
		Get: private(routes.GetPlayer),
	}))

	http.HandleFunc("/auctions", create(RequestRoute{
		Get: public(routes.GetLbin),
	}))
	http.HandleFunc("/shared_data/{player_id}", create(RequestRoute{
		Get: private(routes.GetSharedData),
	}))

	registerUserData := func(name string, handler func(internal.RouteContext, internal.AuthenticationContext, http.ResponseWriter, *http.Request)) {
		http.HandleFunc("/shared_data/{profile_id}/"+name, create(RequestRoute{
			Put: authenticated(handler),
		}))
	}

	http.HandleFunc("/shared_data", create(RequestRoute{
		Delete: authenticated(routes.DeleteData),
	}))
	registerUserData("hotf", routes.PutHotfData)
	registerUserData("hotm", routes.PutHotmData)
	registerUserData("consumeables", routes.PutConsumeablesData)
	registerUserData("hunting_box", routes.PutHuntingBox)
	registerUserData("hunting_toolkit", routes.PutHuntingToolkit)
	registerUserData("melody", routes.PutMelodyData)
	registerUserData("foraging", routes.PutMiscForagingData)
	registerUserData("garden", routes.PutMiscGardenData)
	registerUserData("time_pocket", routes.PutTimePocket)
	registerUserData("garden_chips", routes.PutGardenChips)

	http.HandleFunc("/_ratelimit", create(RequestRoute{
		Get: admin(routes.GetRateLimit),
	}))

	fmt.Printf("Listening on 0.0.0.0:%s\n", routeContext.Config.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", routeContext.Config.Port), nil)

	if err != nil {
		panic(err)
	}
}
