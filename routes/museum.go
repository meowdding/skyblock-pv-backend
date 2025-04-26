package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
	"time"
)

const museumCacheDuration = 5 * time.Minute
const museumCacheName = "museum"
const museumHypixelPath = "/v2/skyblock/museum"

func GetMuseum(ctx utils.RouteContext, authentication utils.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
	profileId := req.PathValue("profile")
	result, err := ctx.GetFromCache(&authentication, museumCacheName, profileId)

	if err != nil {
		profiles, err := utils.GetFromHypixel(ctx, fmt.Sprintf("%s?profile=%s", museumHypixelPath, profileId), true)
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
