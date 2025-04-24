package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
	"time"
)

const statusCacheDuration = 5 * time.Minute
const statusCacheName = "status"
const statusHypixelPath = "/v2/status"

func GetStatus(ctx utils.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("id")
	result, err := ctx.GetFromCache(statusCacheName, playerId)

	if err != nil {
		profiles, err := utils.GetFromHypixel(fmt.Sprintf("%s?uuid=%s", statusHypixelPath, playerId))
		if err == nil {
			err = ctx.AddToCache(statusCacheName, playerId, profiles, statusCacheDuration)
		}

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Failed to fetch or cache status: %v\n", err)
			return
		} else {
			result = *profiles
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(statusCacheDuration.Seconds())))
	_, _ = io.WriteString(res, result)
}
