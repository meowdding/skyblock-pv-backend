package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
	"time"
)

const gardenCacheDuration = 10 * time.Minute
const gardenFailedCacheDuration = 5 * time.Minute
const gardenCacheName = "garden"
const gardenHypixelPath = "/v2/skyblock/garden"

func GetGarden(ctx utils.RouteContext, authentication utils.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
	profileId := req.PathValue("profile")
	result, err := ctx.GetFromCache(&authentication, gardenCacheName, profileId)

	if err != nil {
		if ctx.HasErrorCached(gardenCacheName, profileId) {
			res.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			profiles, err := utils.GetFromHypixel(ctx, fmt.Sprintf("%s?profile=%s", gardenHypixelPath, profileId), true)
			if err == nil {
				err = ctx.AddToCache(gardenCacheName, profileId, profiles, gardenCacheDuration)
			} else {
				cacheError := ctx.AddToErrorCache(gardenCacheName, profileId, gardenFailedCacheDuration)
				if cacheError != nil {
					fmt.Printf("Failed to cache garden error: %v\n", cacheError)
				}
			}

			if err != nil || profiles == nil {
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Printf(
					"[/garden/%s] User '%s' with user-agent '%s' failed to fetch or cache garden: %v\n",
					profileId,
					authentication.Requester,
					req.Header.Get("User-Agent"),
					err,
				)
				return
			} else {
				result = *profiles
			}
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(gardenCacheDuration.Seconds())))
	_, _ = io.WriteString(res, result)
}
