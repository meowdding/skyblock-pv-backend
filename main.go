package main

import (
	"github.com/redis/go-redis/v9"
	"net/http"
	"skyblock-pv-backend/routes"
	"skyblock-pv-backend/routes/utils"
)

type RequestHandler struct {
	method        string
	authenticated bool
	handler       func(utils.RouteContext, http.ResponseWriter, *http.Request)
}

var routeContext = utils.NewRouteContext(redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
}))

func handleRequests(handlers []RequestHandler) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		isAuthenticated := utils.IsAuthenticated(req.Header.Get("Authorization"))

		requiresAuth := false
		badMethod := false

		for _, handler := range handlers {
			if req.Method != handler.method {
				badMethod = true
			} else if !isAuthenticated && handler.authenticated {
				requiresAuth = true
			} else {
				handler.handler(routeContext, res, req)
				return
			}
		}

		if badMethod {
			res.WriteHeader(http.StatusMethodNotAllowed)
		} else if requiresAuth {
			res.WriteHeader(http.StatusUnauthorized)
		}
	}
}

func handleRequest(method string, handler func(utils.RouteContext, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return handleRequests([]RequestHandler{{method, false, handler}})
}

func handleAuthenticatedRequest(method string, handler func(utils.RouteContext, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return handleRequests([]RequestHandler{{method, true, handler}})
}

func main() {
	http.HandleFunc("/authenticate", handleRequest("GET", routes.Authenticate))
	http.HandleFunc("/profiles/{id}", handleAuthenticatedRequest("GET", routes.GetProfiles))
	http.HandleFunc("/garden/{profile}", handleAuthenticatedRequest("GET", routes.GetGarden))
	http.HandleFunc("/museum/{profile}", handleAuthenticatedRequest("GET", routes.GetMuseum))
	http.HandleFunc("/status/{id}", handleAuthenticatedRequest("GET", routes.GetStatus))

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		panic(err)
	}
}
