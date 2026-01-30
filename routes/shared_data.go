//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -o shared_data.gen.go -package routes -generate types,strict-server,std-http-server shared_data_api.yaml
package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"skyblock-pv-backend/internal"
	"skyblock-pv-backend/utils"
	"strings"
)

type defaults interface {
	setupDefaults()
}

const addData = `
	insert into shared_data(player_id, profile_id, data)
	values ($1, $2, jsonb_set(jsonb_build_object(), $3::text[], $4::jsonb))
	on conflict (player_id, profile_id) do update set data = jsonb_set(shared_data.data, $3::text[], $4::jsonb)
`

const deleteUnknownProfiles = `
	delete from shared_data where player_id = $1 and profile_id <> all($2::uuid[])
`

const getSharedData = `
	select data, profile_id from shared_data where player_id = $1
`

func GetSharedData(ctx internal.RouteContext, authentication internal.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("player_id")

	rows, err := ctx.Pool.Query(*ctx.Context, getSharedData, playerId)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Printf(
			"[/shared_data/%s] User '%s' with user-agent '%s': %v\n",
			playerId,
			authentication.Requester,
			req.Header.Get("User-Agent"),
			err,
		)
		return
	}

	dataMap := make(map[string]interface{})
	defer rows.Close()
	for rows.Next() {
		var id string
		var data map[string]interface{}

		err = rows.Scan(&data, &id)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Printf(
				"[/shared_data/%s] User '%s' with user-agent '%s': %v\n",
				playerId,
				authentication.Requester,
				req.Header.Get("User-Agent"),
				err,
			)
			return
		}
		dataMap[id] = data
	}

	data, err := json.Marshal(dataMap)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Printf(
			"[/shared_data/%s] User '%s' with user-agent '%s': %v\n",
			playerId,
			authentication.Requester,
			req.Header.Get("User-Agent"),
			err,
		)
		return
	}
	_, _ = res.Write(data)
}

const deleteSharedData = `
	delete from shared_data where player_id = $1
`

func DeleteData(ctx internal.RouteContext, authentication internal.AuthenticationContext, res http.ResponseWriter, req *http.Request) {
	playerId := authentication.Requester

	if _, err := ctx.Pool.Exec(*ctx.Context, playerId, playerId); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Printf(
			"[/shared_data] Failed to delete player data for '%s' with user-agent '%s': %v\n",
			authentication.Requester,
			req.Header.Get("User-Agent"),
			err,
		)
	}
}

func CheckData(ctx internal.RouteContext, player string, profileIds []string) {
	if _, err := ctx.Pool.Exec(*ctx.Context, deleteUnknownProfiles, player, "{"+strings.Join(profileIds, ",")+"}"); err != nil {
		fmt.Printf(
			"[Chore] Failed to execute deletion of unknown profiles for %s (%s): %v\n",
			player,
			profileIds,
			err,
		)
	}
}

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
		if utils.Debug {
			fmt.Printf(
				"[/shared_data/%s/%s] Updating shared data for '%s' with user-agent '%s'\n",
				profileId,
				key,
				playerId,
				req.Header.Get("User-Agent"),
			)
		}

		res.WriteHeader(http.StatusOK)
	}
}

type TreeNode struct {
	Id       string `json:"id"`
	Level    int    `json:"level"`
	Disabled bool   `json:"disabled"`
}

type HotfData struct {
	ForestWhispers int64      `json:"forest_whispers"`
	Experience     float32    `json:"experience"`
	Level          int        `json:"level"`
	Nodes          []TreeNode `json:"nodes"`
}

func (h *HotfData) setupDefaults() {
	if h.Nodes == nil {
		h.Nodes = make([]TreeNode, 0)
	}
}

var PutHotfData = putData("hotf", func() defaults {
	return &HotfData{}
})

type HotmData struct {
	Experience float32    `json:"experience"`
	Level      int        `json:"level"`
	Nodes      []TreeNode `json:"nodes"`
}

func (h *HotmData) setupDefaults() {
	if h.Nodes == nil {
		h.Nodes = make([]TreeNode, 0)
	}
}

var PutHotmData = putData("hotm", func() defaults {
	return &HotmData{}
})

type Consumeables map[string]int16

func (c *Consumeables) setupDefaults() {}

var PutConsumeablesData = putData("consumeables", func() defaults {
	return &Consumeables{}
})

type HuntingBox map[string]Attribute

type Attribute struct {
	Owned    int32 `json:"owned"`
	Syphoned int32 `json:"consumed"`
}

func (c *HuntingBox) setupDefaults() {}

var PutHuntingBox = putData("hunting_box", func() defaults {
	return &HuntingBox{}
})

type TimePocket map[string]interface{}

func (c *TimePocket) setupDefaults() {}

var PutTimePocket = putData("time_pocket", func() defaults {
	return &TimePocket{}
})

type GardenChips struct {
	VerminVaporizer int32 `json:"vermin_vaporizer"`
	Synthesis       int32 `json:"synthesis"`
	Sowledge        int32 `json:"sowledge"`
	Mechamind       int32 `json:"mechamind"`
	Hypercharge     int32 `json:"hypercharge"`
	Evergreen       int32 `json:"evergreen"`
	Overdrive       int32 `json:"overdrive"`
	Cropshot        int32 `json:"cropshot"`
	Quickdraw       int32 `json:"quickdraw"`
	Rarefinder      int32 `json:"rarefinder"`
}

func (c *GardenChips) setupDefaults() {}

var PutGardenChips = putData("chips", func() defaults {
	return &GardenChips{}
})

type MutationState string

const (
	Unlocked MutationState = "UNLOCKED"
	Analyzed MutationState = "ANALYZED"
	Unknown  MutationState = "UNKNOWN"
)

type MiscGardenData struct {
	UnlockedGreenhouseTiles int32           `json:"unlocked_greenhouse_tiles,omitempty"`
	GrowthSpeed             int32           `json:"growth_speed"`
	PlantYield              int32           `json:"plant_yield"`
	Mutations               []MutationState `json:"mutations"`
}

func (c *MiscGardenData) setupDefaults() {
	if c.Mutations == nil {
		c.Mutations = make([]MutationState, 0)
	}
}

var PutMiscGardenData = putData("garden_data", func() defaults {
	return &GardenChips{}
})

type MiscForagingData struct {
	HuntingExp                 string      `json:"hunting_exp"`
	HuntingAxeItem             interface{} `json:"hunting_axe_item,omitempty"`
	TempleBuffEnd              int64       `json:"temple_buff_end"`
	BeaconTier                 int32       `json:"beacon_tier"`
	ForestEssence              int32       `json:"forest_essence"`
	AgathaLevelCap             int32       `json:"agatha_level_cap"`
	AgathaPower                int32       `json:"agatha_power"`
	FigFortuneLevel            int32       `json:"fig_fortune_level"`
	FigPersonalBest            bool        `json:"fig_personal_best"`
	FigPersonalBestAmount      int32       `json:"fig_personal_best_amount"`
	MangroveFortuneLevel       int32       `json:"mangrove_fortune_level"`
	MangrovePersonalBest       bool        `json:"mangrove_personal_best"`
	MangrovePersonalBestAmount int32       `json:"mangrove_personal_best_amount"`
}

func (c *MiscForagingData) setupDefaults() {}

var PutMiscForagingData = putData("foraging_data", func() defaults {
	return &GardenChips{}
})

type MelodyData struct {
	HymnToTheJoy       int32 `json:"hymn_to_the_joy"`
	FrereJacques       int32 `json:"frere_jacques"`
	AmazingGrace       int32 `json:"amazing_grace"`
	BrahamsLullaby     int32 `json:"brahams_lullaby"`
	HappyBirthdayToYou int32 `json:"happy_birthday_to_you"`
	Greensleeves       int32 `json:"greensleeves"`
	Geothermy          int32 `json:"geothermy"`
	Minuet             int32 `json:"minuet"`
	JoyToTheWorld      int32 `json:"joy_to_the_world"`
	GodlyImagination   int32 `json:"godly_imagination"`
	LaVieEnRose        int32 `json:"la_vie_en_rose"`
	ThroughTheCampfire int32 `json:"through_the_campfire"`
	Pachelbel          int32 `json:"pachelbel"`
}

func (c *MelodyData) setupDefaults() {}

var PutMelodyData = putData("melody_data", func() defaults {
	return &MelodyData{}
})
