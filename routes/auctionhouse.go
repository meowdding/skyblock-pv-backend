package routes

import (
	"io"
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
	res.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(res, *cachedData)
}
