//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -o shared_data.gen.go -package routes -generate types,strict-server,std-http-server shared_data_api.yaml
package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/internal"
)

type HotfData struct {
	ForestWhispers int        `json:"forest_whispers"`
	Experience     float32    `json:"experience"`
	Nodes          []TreeNode `json:"nodes"`
}

func (h *HotfData) setupDefaults() {
	if h.Nodes == nil {
		h.Nodes = make([]TreeNode, 0)
	}
}

type HotmData struct {
	Experience float32    `json:"experience"`
	Nodes      []TreeNode `json:"nodes"`
}

func (h *HotmData) setupDefaults() {
	if h.Nodes == nil {
		h.Nodes = make([]TreeNode, 0)
	}
}

type TreeNode struct {
	Id       string `json:"id"`
	Level    int    `json:"level"`
	Disabled bool   `json:"disabled"`
}

type defaults interface {
	setupDefaults()
}

const addData = `
	insert into shared_data(player_id, profile_id, data)
	values ($1, $2, jsonb_set(jsonb_build_object(), $3::text[], $4::jsonb))
	on conflict (player_id, profile_id) do update set data = jsonb_set(shared_data.data, $3::text[], $4::jsonb)
`

func putData(key string, createData func() defaults) func(ctx internal.RouteContext, authentication internal.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
	dbKey := "{" + key + "}"
	return func(ctx internal.RouteContext, authentication internal.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
		//goland:noinspection GoUnhandledErrorResult
		defer req.Body.Close()
		profileId := req.PathValue("profile_id")
		playerId := authentication.Requester

		data, err := io.ReadAll(req.Body)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf(
				"[/shared_data/%s/%s] User '%s' with user-agent '%s' failed to put %[2]s: %v\n",
				profileId,
				key,
				playerId,
				req.Header.Get("User-Agent"),
				err,
			)
			return
		}

		var userData = createData()
		if err := json.Unmarshal(data, &userData); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		userData.setupDefaults()

		data, err = json.Marshal(userData)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf(
				"[/shared_data/%s/%s] User '%s' with user-agent '%s' failed to put %[2]s: %v\n",
				profileId,
				key,
				playerId,
				req.Header.Get("User-Agent"),
				err,
			)
			return
		}

		if _, err := ctx.Pool.Exec(*ctx.Context, addData, playerId, profileId, dbKey, string(data)); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf(
				"[/shared_data/%s/%s] User '%s' with user-agent '%s' failed to put %[2]s: %v\n",
				profileId,
				key,
				playerId,
				req.Header.Get("User-Agent"),
				err,
			)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

var PutHotfData = putData("hotf", func() defaults {
	return &HotfData{}
})
var PutHotmData = putData("hotm", func() defaults {
	return &HotfData{}
})
