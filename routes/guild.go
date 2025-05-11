package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
	"skyblock-pv-backend/utils/responses"
	"strings"
	"time"
)

const guildCacheDuration = 5 * 24 * time.Hour
const guildCacheName = "guild"
const guildHypixelPath = "/v2/guild"

func cacheGuild(ctx utils.RouteContext, guild string) error {
	var response = responses.GuildResponse{}
	err := json.Unmarshal([]byte(guild), &response)
	if err != nil {
		return err
	}

	if response.Success != true {
		return fmt.Errorf("failed to fetch guild: %s", response.Guild.Name)
	}

	for _, member := range response.Guild.Members {
		realUuid := strings.Join(
			[]string{member.Uuid[0:8], member.Uuid[8:12], member.Uuid[12:16], member.Uuid[16:20], member.Uuid[20:]},
			"-",
		)

		err = ctx.AddToCache(guildCacheName, realUuid, guild, guildCacheDuration)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetGuild(ctx utils.RouteContext, authentication utils.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("id")
	result, err := ctx.GetFromCache(&authentication, guildCacheName, playerId)

	if err != nil {
		guild, err := utils.GetFromHypixel(ctx, fmt.Sprintf("%s?player=%s", guildHypixelPath, playerId), true)
		if err == nil {
			err = cacheGuild(ctx, *guild)
		}

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Failed to fetch or cache guild: %v\n", err)
			return
		} else {
			result = *guild
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(guildCacheDuration.Seconds())))
	_, _ = io.WriteString(res, result)
}
