package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
	"time"
)

const museumCacheDuration = 5 * time.Minute
const museumCacheName = "garden"
const museumHypixelPath = "/v2/skyblock/museum"

func GetMuseum(ctx utils.RouteContext, res http.ResponseWriter, req *http.Request) {
	profileId := req.PathValue("profile")
	result, err := ctx.GetFromCache(museumCacheName, profileId)

	if err != nil {
		profiles, err := utils.GetFromHypixel(fmt.Sprintf("%s?profile=%s", museumHypixelPath, profileId))
		if err == nil {
			err = ctx.AddToCache(museumCacheName, profileId, profiles, museumCacheDuration)
		}

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Failed to fetch or cache meseum: %v\n", err)
			return
		} else {
			result = *profiles
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(museumCacheDuration.Seconds())))
	_, _ = io.WriteString(res, result)
}
