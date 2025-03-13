package discord

import (
	"encoding/json"
	"fmt"
)

// Função responsável por lidar com eventos recebidos do WebSocket do Discord.
func (dc *DiscordClient) HandleEvents() error {
	for {
		var payload GatewayPayload

		// Lê um evento do WebSocket e converte para a estrutura GatewayPayload
		if err := dc.WsConn.ReadJSON(&payload); err != nil {
			dc.HandleError(err) // Em caso de erro, trata e tenta reconectar
			return err
		}

		// Verifica o código de operação (Op Code) recebido no evento
		switch payload.Op {
		case 0: // Evento de Dispatch (evento normal)
			dc.Sequence = *payload.S // Atualiza o número da sequência de eventos recebidos

			// Verifica se o evento recebido é uma interação de comando (slash command)
			if payload.T != nil && *payload.T == "INTERACTION_CREATE" {
				data, _ := json.Marshal(payload.D)
				var interactionEvent map[string]interface{}
				json.Unmarshal(data, &interactionEvent)

				// Obtém o tipo da interação
				interactionType := interactionEvent["type"].(float64)

				if interactionType == 4 { // Tipo 4: Autocomplete de comando
					commandData := interactionEvent["data"].(map[string]interface{})
					commandName := commandData["name"].(string)

					// Verifica se o comando existe no registro
					cmdInfo, exists := dc.Registry.GetCommand(commandName)
					if exists {
						// Verifica se o comando implementa a interface de autocomplete
						if autoCmd, ok := cmdInfo.Command.(interface {
							HandleAutocomplete(map[string]interface{}) error
						}); ok {
							autoCmd.HandleAutocomplete(interactionEvent)
						}
					}
				} else if interactionType == 2 { // Tipo 2: Execução normal de um comando
					commandData := interactionEvent["data"].(map[string]interface{})
					commandName := commandData["name"].(string)

					// Verifica se o comando existe no registro
					cmdInfo, exists := dc.Registry.GetCommand(commandName)
					if exists {
						// Executa o comando associado
						cmdInfo.Command.Execute(interactionEvent)
					}
				}
			}
		case 11: // Evento de reconhecimento de Heartbeat
			fmt.Println("Heartbeat recognized: ", payload.Op)
		}
	}
}

// Basicamnete monitora os eventos recebidos do WebSocket do Disc
