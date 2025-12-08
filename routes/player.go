package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
	"time"
)

const playerCacheDuration = 12 * time.Hour
const playerFailedCacheDuration = 5 * time.Minute
const playerCacheName = "player"
const playerHypixelPath = "/v2/player"

func GetPlayer(ctx utils.RouteContext, authentication utils.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
	if !ctx.Config.Endpoints.Players {
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(playerCacheDuration.Seconds())))
		_, _ = io.WriteString(res, "{}")
	} else {
		playerId := req.PathValue("id")
		result, err := ctx.GetFromCache(&authentication, playerCacheName, playerId)

		if err != nil {
			if ctx.HasErrorCached(playerCacheName, playerId) {
				res.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				profiles, err := utils.GetFromHypixel(ctx, fmt.Sprintf("%s?uuid=%s", playerHypixelPath, playerId), true)
				if err == nil {
					err = ctx.AddToCache(playerCacheName, playerId, profiles, playerCacheDuration)
				} else {
					cacheError := ctx.AddToErrorCache(playerCacheName, playerId, playerFailedCacheDuration)
					if cacheError != nil {
						fmt.Printf("Failed to cache player error: %v\n", cacheError)
					}
				}

				if err != nil || profiles == nil {
					res.WriteHeader(http.StatusInternalServerError)
					fmt.Printf("[/player/%s] User '%s' failed to fetch or cache player: %v\n", playerId, authentication.Requester, err)
					return
				} else {
					result = *profiles
				}
			}
		}

		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(playerCacheDuration.Seconds())))
		_, _ = io.WriteString(res, result)
	}
}
