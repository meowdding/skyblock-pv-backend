package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
	"time"
)

const profileCacheDuration = 5 * time.Minute
const profileCacheName = "profiles"
const profileHypixelPath = "/v2/skyblock/profiles"

func GetProfiles(ctx utils.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("id")
	result, err := ctx.GetFromCache(profileCacheName, playerId)

	if err != nil {
		println("Cache miss, fetching from Hypixel API")

		profiles, err := utils.GetFromHypixel(ctx, fmt.Sprintf("%s?uuid=%s", profileHypixelPath, playerId), true)
		if err == nil {
			err = ctx.AddToCache(profileCacheName, playerId, profiles, profileCacheDuration)
		}

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Failed to fetch or cache profiles: %v\n", err)
			return
		} else {
			result = *profiles
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(profileCacheDuration.Seconds())))
	_, _ = io.WriteString(res, result)
}
