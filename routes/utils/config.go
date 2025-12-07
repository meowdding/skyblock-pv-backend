package utils

import (
	"encoding/json"
	"os"
)

type Config struct {
	RedisAddress  string   `json:"redis_address"`
	RedisUsername string   `json:"redis_username"`
	RedisPassword string   `json:"redis_password"`
	JwtToken      string   `json:"jwt_token"`
	HypixelKey    []string `json:"hypixel_key"`
	Port          string   `json:"port"`
	Admins        []string `json:"admins"`
	DevMode       bool     `json:"dev_mode"`
	NoAuth        bool     `json:"no_auth"`
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
