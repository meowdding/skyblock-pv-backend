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
	} else {
		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf(mojangAuthUrl, username, server),
			http.NoBody,
		)
		req.Close = true

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Failed to authenticate: %v\n", err)
			return
		}

		r, err := http.DefaultClient.Do(req)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Failed to authenticate: %v\n", err)
			return
		}

		defer r.Body.Close()

		if r.StatusCode != http.StatusOK {
			res.WriteHeader(http.StatusUnauthorized)
			fmt.Printf("Authentication failed: %s\n", r.Status)
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
	}
}
