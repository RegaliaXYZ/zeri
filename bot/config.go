package bot

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Token      string `json:"token"`
	BotPrefix  string `json:"botPrefix"`
	AppID      string `json:"appID"`
	RiotAPIKey string `json:"RiotApiKey"`
}

func ReadConfig() (*Config, error) {
	data, err := os.ReadFile("./config.json")
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal([]byte(data), &cfg)
	if err != nil {
		fmt.Println("Error unmarshalling config.json")
		return nil, err
	}
	return &cfg, nil
}
