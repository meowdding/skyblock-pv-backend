package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
	"slices"
)

const mojangAuthUrl = "https://sessionserver.mojang.com/session/minecraft/hasJoined?username=%s&serverId=%s"

type SessionResponse struct {
	Id string `json:"id"`
}

func Authenticate(ctx utils.RouteContext, res http.ResponseWriter, req *http.Request) {
	username := req.Header.Get("x-minecraft-username")
	server := req.Header.Get("x-minecraft-server")

	if username == "" || server == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if slices.Contains(ctx.Config.Endpoints.Authenticate.BannedAccounts, username) {
		res.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(res, "Your account has been banned from using this service.")
		return
	}

	if ctx.Config.Endpoints.Authenticate.Enabled {
		r, err := http.Get(fmt.Sprintf(mojangAuthUrl, username, server))

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Authentication failed for user '%s' with error: %v\n", username, err)
			return
		}

		defer r.Body.Close()

		if r.StatusCode != http.StatusOK {
			res.WriteHeader(http.StatusUnauthorized)
			fmt.Printf("Authentication failed for user '%s' with status code %d\n", username, r.StatusCode)
			return
		}

		var session SessionResponse

		err = json.NewDecoder(r.Body).Decode(&session)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Failed to decode session response: %v\n", err)
		} else {
			bypassCache := req.URL.Query().Has("bypassCache") && slices.Contains(ctx.Config.Admins, session.Id)
			token, err := utils.CreateAuthenticationKey(ctx, session.Id, bypassCache)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Printf("Failed to create authentication key: %v\n", err)
			} else {
				_, _ = io.WriteString(res, token)
			}
		}
	} else {
		token, err := utils.CreateAuthenticationKey(ctx, "00000000000000000000000000000000", false)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Failed to create authentication key: %v\n", err)
		} else {
			_, _ = io.WriteString(res, token)
		}
	}
}
