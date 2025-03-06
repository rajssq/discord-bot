package discord

import (
	"bot-map/cmd"
	"bot-map/config"
	"bot-map/shared"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketConn interface {
	ReadJSON(v interface{}) error
	WriteJSON(v interface{}) error
	Close() error
}

type DiscordClient struct {
	Client    shared.HTTPClient
	Config    *config.Config
	WsConn    WebsocketConn
	Heartbeat time.Duration
	Sequence  int
	Registry  *cmd.CommandRegistry
}

type GatewayPayload struct {
	Op int         `json:"op"`
	D  interface{} `json:"d"`
	S  *int        `json:"s,omitempty"`
	T  *string     `json:"t,omitempty"`
}

type HelloEvent struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type IdentifyData struct {
	Token      string            `json:"token"`
	Properties map[string]string `json:"properties"`
	Intents    int               `json:"intents"`
}

var discordClientInstace *DiscordClient

func GetDiscordClient(config *config.Config, registry *cmd.CommandRegistry, client ...shared.HTTPClient) *DiscordClient {
	if discordClientInstace == nil {
		discordClientInstace = &DiscordClient{
			Client: func() shared.HTTPClient {
				if len(client) > 0 {
					return client[0]
				}
				return &http.Client{}
			}(),
			Config:   config,
			Registry: registry,
		}
	}

	return discordClientInstace
}

func (dc *DiscordClient) RegisterSlashCommands() {
	commands := dc.Registry.GetAllCommands()
	for _, cmdInfo := range commands {

		command := map[string]interface{}{
			"name":        cmdInfo.Name,
			"description": cmdInfo.Description,
			"options":     cmdInfo.Options,
		}

		jsonCommand, _ := json.Marshal(command)

		url := fmt.Sprintf("%s/applications/%s/guilds/%s/commands", dc.Config.BaseURL, dc.Config.ApplicationID, dc.Config.GuildID)

		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonCommand))

		req.Header.Set("Authorization", "Bot "+dc.Config.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := dc.Client.Do(req)
		if err != nil {
			fmt.Println("error")
		}
		defer resp.Body.Close()
	}
}

func (dc *DiscordClient) ConnectToGateway() error {
	gatewayURL := dc.Config.GatewayURL

	dialer := websocket.DefaultDialer
	ws, _, err := dialer.Dial(gatewayURL, nil)
	if err != nil {
		return err
	}

	dc.WsConn = ws

	fmt.Println("connected")

	var payload GatewayPayload
	dc.WsConn.ReadJSON(&payload)

	if payload.Op != 10 {
		return fmt.Errorf("op code unexpected: %d", payload.Op)
	}

	var hello HelloEvent

	data, err := json.Marshal(payload.D)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &hello); err != nil {
		return err
	}

	dc.Heartbeat = time.Duration(hello.HeartbeatInterval) * time.Millisecond

	const (
		IntentsGuilds         = 1 << 0
		IntentsGuildMessages  = 1 << 9
		IntentsMessageContent = 1 << 15
	)

	indentifyPayload := GatewayPayload{
		Op: 2,
		D: IdentifyData{
			Token: dc.Config.Token,
			Properties: map[string]string{
				"os":      "linux",
				"browser": "my_bot",
				"device":  "my_bot",
			},
			Intents: IntentsGuilds | IntentsGuildMessages | IntentsMessageContent,
		},
	}

	if err := dc.WsConn.WriteJSON(indentifyPayload); err != nil {
		return err
	}

	go dc.StartHeartbeating()
	go dc.HandleEvents()

	return nil
}

func (dc *DiscordClient) StartHeartbeating() error {
	ticker := time.NewTicker(dc.Heartbeat)

	defer ticker.Stop()

	for range ticker.C {
		heartbeat := GatewayPayload{
			Op: 1,
			D:  dc.Sequence,
		}

		if err := dc.WsConn.WriteJSON(heartbeat); err != nil {
			return err
		}

		fmt.Println("Heartbeat sent")
	}

	return nil
}

func (dc *DiscordClient) HandleError(err error) {
	fmt.Println("error detected", err)

	if closeErr := dc.WsConn.Close(); closeErr != nil {
		fmt.Println("could not close connection: ", closeErr)
	}

	dc.ConnectToGateway()
}
