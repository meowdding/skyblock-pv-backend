package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/auctions"
	"skyblock-pv-backend/internal"
)

func GetLbin(ctx internal.RouteContext, res http.ResponseWriter, _ *http.Request) {
	cachedData, err := auctions.GetCachedAuctions(&ctx)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set("X-Auction-Version", fmt.Sprintf("v%d", auctions.AuthCacheVersion))
	res.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(res, *cachedData)
}
