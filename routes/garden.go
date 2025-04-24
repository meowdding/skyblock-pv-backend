package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
	"time"
)

const gardenCacheDuration = 10 * time.Minute
const gardenCacheName = "garden"
const gardenHypixelPath = "/v2/skyblock/garden"

func GetGarden(ctx utils.RouteContext, res http.ResponseWriter, req *http.Request) {
	profileId := req.PathValue("profile")
	result, err := ctx.GetFromCache(gardenCacheName, profileId)

	if err != nil {
		profiles, err := utils.GetFromHypixel(fmt.Sprintf("%s?profile=%s", gardenHypixelPath, profileId))
		if err == nil {
			err = ctx.AddToCache(gardenCacheName, profileId, profiles, gardenCacheDuration)
		}

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Failed to fetch or cache garden: %v\n", err)
			return
		} else {
			result = *profiles
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(gardenCacheDuration.Seconds())))
	_, _ = io.WriteString(res, result)
}
