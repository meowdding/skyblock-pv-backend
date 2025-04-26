package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
	"time"
)

const playerAuctionsCacheDuration = 10 * time.Minute
const playerAuctionsCacheName = "player_active_auctions"
const playerAuctionsHypixelPath = "/v2/skyblock/auction"

func GetActiveProfileAuctions(ctx utils.RouteContext, authentication utils.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
	profileId := req.PathValue("profile")
	result, err := ctx.GetFromCache(&authentication, playerAuctionsCacheName, profileId)

	if err != nil {
		auctions, err := utils.GetFromHypixel(ctx, fmt.Sprintf("%s?profile=%s", playerAuctionsHypixelPath, profileId), true)
		if err == nil {
			transformedAuctions, err := transformAuctions(*auctions)
			auctions = &transformedAuctions

			if err == nil {
				err = ctx.AddToCache(playerAuctionsCacheName, profileId, auctions, playerAuctionsCacheDuration)
			}
		}

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Failed to fetch or cache player active auctions: %v\n", err)
			return
		} else {
			result = *auctions
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(playerAuctionsCacheDuration.Seconds())))
	_, _ = io.WriteString(res, result)
}

func transformAuctions(auctionsText string) (string, error) {
	var auctions = make(map[string]interface{})
	err := json.Unmarshal([]byte(auctionsText), &auctions)
	if err != nil {
		return "", err
	}

	if (auctions["success"] == nil) || (auctions["success"] != true) {
		return auctionsText, nil
	} else {
		var currentTime = float64(time.Now().UnixMilli())
		var realAuctions = auctions["auctions"].([]interface{})
		var transformedAuctions = make([]map[string]interface{}, 0)

		for _, auction := range realAuctions {
			a, ok := auction.(map[string]interface{})
			if !ok {
				continue
			}
			end, ok := a["end"].(float64)
			if ok && end > currentTime {
				transformedAuctions = append(transformedAuctions, a)
			}
		}

		auctions["auctions"] = transformedAuctions

		data, err := json.Marshal(auctions)

		if err != nil {
			return "", err
		} else {
			return string(data), nil
		}
	}
}
