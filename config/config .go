package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BaseURL       string
	Token         string
	ChannelID     string
	ApplicationID string
	GuildID       string
	GatewayURL    string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		fmt.Println("unable to read .env variables")
	}

	return &Config{
		BaseURL:       os.Getenv("BASE_URL"),
		Token:         os.Getenv("TOKEN"),
		ChannelID:     os.Getenv("CHANNEL_ID"),
		ApplicationID: os.Getenv("APPLICATION_ID"),
		GuildID:       os.Getenv("GUILD_ID"),
		GatewayURL:    os.Getenv("GATEWAY_URL"),
	}
}
