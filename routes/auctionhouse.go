package routes

import (
	"net/http"
	"skyblock-pv-backend/auctions"
	"skyblock-pv-backend/routes/utils"
)

func GetLbin(ctx utils.RouteContext, res http.ResponseWriter, _ *http.Request) {
	cachedData, err := auctions.GetCachedAuctions(&ctx)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write([]byte(*cachedData))
}
