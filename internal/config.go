package internal

import (
	"encoding/json"
	"os"
)

type Config struct {
	RedisAddress        string          `json:"redis_address"`
	RedisUsername       *string         `json:"redis_username,omitempty"`
	RedisPassword       *string         `json:"redis_password,omitempty"`
	JwtToken            string          `json:"jwt_token"`
	HypixelKey          []string        `json:"hypixel_key"`
	Port                string          `json:"port"`
	Admins              []string        `json:"admins"`
	DevMode             bool            `json:"dev_mode"`
	HighProfileAccounts []string        `json:"high_profile_accounts"`
	Endpoints           EndpointsConfig `json:"endpoints"`
	PostgresUri         *string         `json:"postgres_uri,omitempty"`
}

type EndpointsConfig struct {
	Players      bool                       `json:"players"`
	RateLimit    bool                       `json:"rate_limit"`
	Authenticate AuthenticateEndpointConfig `json:"authenticate"`
}

type AuthenticateEndpointConfig struct {
	Enabled        bool     `json:"enabled"`
	BannedAccounts []string `json:"banned_accounts"`
}

func NewConfig() Config {
	env := os.Getenv("CONFIG")

	if env == "" {
		panic("CONFIG environment variable not set")
	}

	var config Config
	err := json.Unmarshal([]byte(env), &config)
	if err != nil {
		panic("Failed to parse config: " + err.Error())
	}
	return config
}
