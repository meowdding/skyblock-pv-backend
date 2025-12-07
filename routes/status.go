package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
	"time"
)

const statusCacheDuration = 5 * time.Minute
const statusFailedCacheDuration = 3 * time.Minute
const statusCacheName = "status"
const statusHypixelPath = "/v2/status"

func GetStatus(ctx utils.RouteContext, authentication utils.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("id")
	result, err := ctx.GetFromCache(&authentication, statusCacheName, playerId)

	if err != nil {
		if ctx.HasErrorCached(statusCacheName, playerId) {
			res.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			profiles, err := utils.GetFromHypixel(ctx, fmt.Sprintf("%s?uuid=%s", statusHypixelPath, playerId), true)
			if err == nil {
				err = ctx.AddToCache(statusCacheName, playerId, profiles, statusCacheDuration)
			} else {
				err = ctx.AddToErrorCache(statusCacheName, playerId, statusFailedCacheDuration)
			}

			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Printf("Failed to fetch or cache status: %v\n", err)
				return
			} else {
				result = *profiles
			}
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(statusCacheDuration.Seconds())))
	_, _ = io.WriteString(res, result)
}
