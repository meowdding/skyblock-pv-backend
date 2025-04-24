package routes

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const cacheDuration = 5 * time.Minute
const cacheName = "profiles"
const hypixelPath = "/v2/skyblock/profiles"

func GetProfiles(ctx RouteContext, res http.ResponseWriter, req *http.Request) {
	profileId := req.PathValue("id")
	result, err := ctx.GetFromCache(cacheName, profileId)

	if err != nil {
		println("Cache miss, fetching from Hypixel API")

		profiles, err := GetFromHypixel(fmt.Sprintf("%s?uuid=%s", hypixelPath, profileId))
		if err == nil {
			err = ctx.AddToCache(cacheName, profileId, profiles, cacheDuration)
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
	res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(cacheDuration.Seconds())))
	_, _ = io.WriteString(res, result)
}
