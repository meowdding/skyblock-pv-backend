package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/routes/utils"
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
		r, err := http.Get(fmt.Sprintf(mojangAuthUrl, username, server))

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
			token, err := utils.CreateAuthenticationKey(session.Id)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Printf("Failed to create authentication key: %v\n", err)
			} else {
				_, _ = io.WriteString(res, token)
			}
		}
	}
}
