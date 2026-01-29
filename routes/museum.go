package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/internal"
	"time"
)

const museumCacheDuration = 5 * time.Minute
const museumFailedCacheDuration = 3 * time.Minute
const museumCacheName = "museum"
const museumHypixelPath = "/v2/skyblock/museum"

func GetMuseum(ctx internal.RouteContext, authentication internal.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
	profileId := req.PathValue("profile")
	result, err := ctx.GetFromCache(&authentication, museumCacheName, profileId)

	if err != nil {
		if ctx.HasErrorCached(museumCacheName, profileId) {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		profiles, err := internal.GetFromHypixel(ctx, fmt.Sprintf("%s?profile=%s", museumHypixelPath, profileId), true)
		if err == nil && profiles != nil {
			err = ctx.AddToCache(museumCacheName, profileId, profiles, museumCacheDuration)
		} else {
			cacheError := ctx.AddToErrorCache(museumCacheName, profileId, museumFailedCacheDuration)
			if cacheError != nil {
				fmt.Printf("Failed to cache meseum error: %v\n", cacheError)
			}
		}

		if err != nil || profiles == nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf(
				"[/museum/%s] User '%s' with user-agent '%s' failed to fetch or cache museum: %v\n",
				profileId,
				authentication.Requester,
				req.Header.Get("User-Agent"),
				err,
			)
			return
		}
		result = *profiles
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(museumCacheDuration.Seconds())))
	_, _ = io.WriteString(res, result)
}
