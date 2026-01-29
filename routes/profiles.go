package routes

import (
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/internal"
	"strconv"
	"time"
)

const profileCacheDuration = 5 * time.Minute
const highProfileCacheDuration = 15 * time.Minute
const profileFailedCacheDuration = 3 * time.Minute
const profileCacheName = "profiles"
const profileHypixelPath = "/v2/skyblock/profiles"

func GetProfiles(ctx internal.RouteContext, authentication internal.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("id")
	result, err := ctx.GetFromCache(&authentication, profileCacheName, playerId)

	if err != nil {
		if ctx.HasErrorCached(profileCacheName, playerId) {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		profiles, err := internal.GetFromHypixel(ctx, fmt.Sprintf("%s?uuid=%s", profileHypixelPath, playerId), true)
		if err == nil && profiles != nil {
			cacheDuration := profileCacheDuration
			if ctx.IsHighProfileAccount(playerId) {
				cacheDuration = highProfileCacheDuration
			}
			err = ctx.AddToCache(profileCacheName, playerId, profiles, cacheDuration)
		} else {
			cacheError := ctx.AddToErrorCache(profileCacheName, playerId, profileFailedCacheDuration)
			if cacheError != nil {
				fmt.Printf("Failed to cache profiles error: %v\n", cacheError)
			}
		}

		if err != nil || profiles == nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf(
				"[/profiles/%s] User '%s' with user-agent '%s' failed to fetch or cache profiles: %v\n",
				playerId,
				authentication.Requester,
				req.Header.Get("User-Agent"),
				err,
			)
			return
		}
		result = *profiles
	}

	milli, err := ctx.GetTtlMilli(profileCacheName, playerId)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Failed to fetch or cache profiles: %v\n", err)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(profileCacheDuration.Seconds())))
	res.Header().Set("X-Backend-Expire-In", strconv.FormatInt(int64(milli), 10))
	_, _ = io.WriteString(res, result)
}
