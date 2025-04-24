package main

import (
	"github.com/redis/go-redis/v9"
	"net/http"
	"skyblock-pv-backend/routes"
	"skyblock-pv-backend/routes/utils"
)

func handleRequest(method string, handler func(utils.RouteContext, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			res.WriteHeader(http.StatusMethodNotAllowed)
		} else {
			ctx := utils.NewRouteContext(redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			}))

			handler(ctx, res, req)
		}
	}
}

func main() {
	http.HandleFunc("/profiles/{id}", handleRequest("GET", routes.GetProfiles))
	http.HandleFunc("/garden/{profile}", handleRequest("GET", routes.GetGarden))
	http.HandleFunc("/museum/{profile}", handleRequest("GET", routes.GetMuseum))
	http.HandleFunc("/status/{id}", handleRequest("GET", routes.GetStatus))

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		panic(err)
	}
}
