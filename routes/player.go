package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
	"time"
)

const playerCacheDuration = 12 * time.Hour
const playerCacheName = "player"
const playerHypixelPath = "/v2/player"

func GetPlayer(ctx utils.RouteContext, authentication utils.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("id")
	result, err := ctx.GetFromCache(&authentication, playerCacheName, playerId)

	if err != nil {
		profiles, err := utils.GetFromHypixel(ctx, fmt.Sprintf("%s?uuid=%s", playerHypixelPath, playerId), true)
		if err == nil {
			err = ctx.AddToCache(playerCacheName, playerId, profiles, playerCacheDuration)
		}

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Failed to fetch or cache player: %v\n", err)
			return
		} else {
			result = *profiles
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(playerCacheDuration.Seconds())))
	_, _ = io.WriteString(res, result)
}
