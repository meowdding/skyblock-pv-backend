package main

import (
	"github.com/redis/go-redis/v9"
	"net/http"
	"skyblock-pv-backend/routes"
)

func handleRequest(method string, handler func(routes.RouteContext, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			res.WriteHeader(http.StatusMethodNotAllowed)
		} else {
			ctx := routes.NewRouteContext(redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			}))

			handler(ctx, res, req)
		}
	}
}

func main() {
	http.HandleFunc("/profiles/{id}", handleRequest("GET", routes.GetProfiles))

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		panic(err)
	}
}
