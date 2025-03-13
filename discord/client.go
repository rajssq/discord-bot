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

// Interface para a conexão WebSocket, permitindo testes mockados.
type WebsocketConn interface {
	ReadJSON(v interface{}) error
	WriteJSON(v interface{}) error
	Close() error
}

// Estrutura que representa o cliente do Discord, contendo configurações, conexão WebSocket e registro de comandos.
type DiscordClient struct {
	Client    shared.HTTPClient    // Cliente HTTP para requisições
	Config    *config.Config       // Configurações do bot
	WsConn    WebsocketConn        // Conexão WebSocket com o gateway do Discord
	Heartbeat time.Duration        // Intervalo de tempo para envio de heartbeat
	Sequence  int                  // Número da sequência de eventos recebidos
	Registry  *cmd.CommandRegistry // Registro de comandos disponíveis
}

// Estrutura que representa um payload enviado para o gateway do Discord.
type GatewayPayload struct {
	Op int         `json:"op"`          // Código da operação
	D  interface{} `json:"d"`           // Dados da operação
	S  *int        `json:"s,omitempty"` // Número da sequência (opcional)
	T  *string     `json:"t,omitempty"` // Tipo do evento (opcional)
}

// Estrutura que representa o evento "Hello" do Discord, contendo o intervalo do heartbeat.
type HelloEvent struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// Estrutura para o payload de identificação do bot no gateway.
type IdentifyData struct {
	Token      string            `json:"token"`      // Token do bot
	Properties map[string]string `json:"properties"` // Propriedades do cliente
	Intents    int               `json:"intents"`    // Intenções (eventos que o bot quer receber)
}

// Instância única do cliente do Discord (Singleton).
var discordClientInstace *DiscordClient

// Função para obter a instância do cliente do Discord, garantindo que apenas uma seja criada.
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

// Função para registrar comandos slash no Discord.
func (dc *DiscordClient) RegisterSlashCommands() {
	commands := dc.Registry.GetAllCommands()
	for _, cmdInfo := range commands {

		command := map[string]interface{}{
			"name":        cmdInfo.Name,
			"description": cmdInfo.Description,
			"options":     cmdInfo.Options,
		}

		jsonCommand, _ := json.Marshal(command)

		// URL para registrar comandos no servidor do Discord
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

// Função para conectar ao gateway do Discord via WebSocket.
func (dc *DiscordClient) ConnectToGateway() error {
	gatewayURL := dc.Config.GatewayURL

	// Cria a conexão WebSocket
	dialer := websocket.DefaultDialer
	ws, _, err := dialer.Dial(gatewayURL, nil)
	if err != nil {
		return err
	}

	dc.WsConn = ws

	fmt.Println("connected")

	// Lê o primeiro payload recebido do gateway
	var payload GatewayPayload
	dc.WsConn.ReadJSON(&payload)

	// Verifica se a operação recebida é "Hello" (código 10)
	if payload.Op != 10 {
		return fmt.Errorf("op code unexpected: %d", payload.Op)
	}

	var hello HelloEvent

	// Converte os dados do payload para a estrutura HelloEvent
	data, err := json.Marshal(payload.D)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &hello); err != nil {
		return err
	}

	// Define o intervalo do heartbeat com base nos dados recebidos
	dc.Heartbeat = time.Duration(hello.HeartbeatInterval) * time.Millisecond

	// Define os intents do bot (quais eventos ele irá receber)
	const (
		IntentsGuilds         = 1 << 0  // Eventos de guildas
		IntentsGuildMessages  = 1 << 9  // Mensagens enviadas nas guildas
		IntentsMessageContent = 1 << 15 // Conteúdo das mensagens
	)

	// Cria o payload de identificação para autenticar no gateway do Discord
	identifyPayload := GatewayPayload{
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

	// Envia o payload de identificação
	if err := dc.WsConn.WriteJSON(identifyPayload); err != nil {
		return err
	}

	// Inicia o envio de heartbeats e o tratamento de eventos em goroutines
	go dc.StartHeartbeating()
	go dc.HandleEvents()

	return nil
}

// Função para enviar heartbeats periodicamente para manter a conexão ativa.
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

// Função para tratar erros e tentar reconectar ao gateway.
func (dc *DiscordClient) HandleError(err error) {
	fmt.Println("error detected", err)

	// Fecha a conexão WebSocket ao detectar um erro
	if closeErr := dc.WsConn.Close(); closeErr != nil {
		fmt.Println("could not close connection: ", closeErr)
	}

	// Tenta se reconectar ao gateway
	dc.ConnectToGateway()
}

// Basicamente esse codigo (cliente.go) gerencia as comunicacoes entre bot e disc via WebSocket
